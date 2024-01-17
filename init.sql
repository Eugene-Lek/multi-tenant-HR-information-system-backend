CREATE USER hr_information_system WITH PASSWORD 'abcd1234';

-- Revoke all Schema-level & Table-level permissions from the Public role (default role assigned to all users)
-- Note that "public" is also the name for the default Schema (a Schema is a namespace containing Tables among other things)
REVOKE ALL ON SCHEMA public FROM public;
REVOKE ALL ON ALL TABLEs IN SCHEMA public FROM public;

-- Grant all NECESSARY Schema-level & Table-level permissions to the api user
-- Schema-level permissions are necessary to access anything inside the Schema (including Tables)
-- Note: The 2nd statement only grants the api user these permissions for Tables created by postgres (user)
GRANT USAGE ON SCHEMA public TO hr_information_system;
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO hr_information_system;


-- Import necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Instantiate the Tables
CREATE TABLE IF NOT EXISTS tenant (
    id UUID PRIMARY KEY NOT NULL,
    name VARCHAR(300) UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS division (
    id UUID PRIMARY KEY NOT NULL,    
    tenant_id UUID NOT NULL,
    name VARCHAR(300) NOT NULL,    
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (tenant_id, name),
    FOREIGN KEY (tenant_id) REFERENCES tenant(id)
);

CREATE TABLE IF NOT EXISTS department (
    id UUID PRIMARY KEY NOT NULL,    
    tenant_id UUID NOT NULL,    
    division_id UUID NOT NULL,
    name VARCHAR(300) NOT NULL,    
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),  

    UNIQUE (division_id, name),
    FOREIGN KEY (tenant_id) REFERENCES tenant(id),
    FOREIGN KEY (division_id) REFERENCES division(id)
);

CREATE TABLE IF NOT EXISTS user_account (
    id UUID PRIMARY KEY NOT NULL, -- ID used as PK to enable changes to email
    tenant_id UUID NOT NULL,    
    email VARCHAR(300) NOT NULL,
    password TEXT NOT NULL,
    totp_secret_key CHAR(32) NOT NULL, --TOTP key is recommended to have 160 bits, which is 32 base32 characters
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_login TIMESTAMPTZ,

    UNIQUE(tenant_id, email),
    FOREIGN KEY (tenant_id) REFERENCES tenant(id)
);

CREATE TABLE IF NOT EXISTS position (
    id UUID PRIMARY KEY NOT NULL,  
    tenant_id UUID NOT NULL,      
    title VARCHAR(300) NOT NULL,
    department_id UUID NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),    

    UNIQUE (title, department_id),
    FOREIGN KEY (tenant_id) REFERENCES tenant(id),    
    FOREIGN KEY (department_id) REFERENCES department(id) -- Every appointment must correspond to a department
);

CREATE TABLE IF NOT EXISTS position_assignment (
    tenant_id UUID NOT NULL,      
    position_id UUID NOT NULL,
    user_account_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL DEFAULT'9999-12-31',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),    

    PRIMARY KEY (position_id, user_account_id),
    FOREIGN KEY (tenant_id) REFERENCES tenant(id),    
    FOREIGN KEY (position_id) REFERENCES position(id),
    FOREIGN KEY (user_account_id) REFERENCES user_account(id)
);

CREATE TABLE IF NOT EXISTS subordinate_supervisor_relationship (
    subordinate_position_id UUID NOT NULL,
    supervisor_position_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),      

    PRIMARY KEY (subordinate_position_id, supervisor_position_id),
    FOREIGN KEY (subordinate_position_id) REFERENCES position(id),
    FOREIGN KEY (supervisor_position_id) REFERENCES position(id),

    CHECK (subordinate_position_id <> supervisor_position_id)
);

-- Seed of root role administrator
-- Password: jU%q837d!QP7
-- Totp Key: OLDFXRMH35A3DU557UXITHYDK4SKLTXZ
INSERT INTO tenant (id, name) 
VALUES ('2ad1dcfc-8867-49f7-87a3-8bd8d1154924', 'HRIS Enterprises');
INSERT INTO division (id, tenant_id, name) 
VALUES ('f8b1551a-71bb-48c4-924a-8a25a6bff71d', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', 'Operations');
INSERT INTO department (id, tenant_id, division_id, name) 
VALUES ('9147b727-1955-437b-be7d-785e9a31f20c', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', 'f8b1551a-71bb-48c4-924a-8a25a6bff71d', 'Administration');

INSERT INTO user_account (id, tenant_id, email, password, totp_secret_key) 
VALUES ('e7f31b70-ae26-42b3-b7a6-01ec68d5c33a', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', 'root-role-admin@hrisEnterprises.org',
'$argon2id$v=19$m=65536,t=1,p=8$cFTNg+YXrN4U0lvwnamPkg$0RDBxH+EouVxDbBlQUNctdWZ+CNKrayPpzTJaWNq83U', 
'OLDFXRMH35A3DU557UXITHYDK4SKLTXZ');

INSERT INTO position (id, tenant_id, title, department_id)
VALUES ('e4edbd37-164d-478d-9625-5b1397ef6e45', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', 'System Administrator', '9147b727-1955-437b-be7d-785e9a31f20c');

INSERT INTO position_assignment (tenant_id, position_id, user_account_id, start_date)
VALUES ('2ad1dcfc-8867-49f7-87a3-8bd8d1154924', 'e4edbd37-164d-478d-9625-5b1397ef6e45', 'e7f31b70-ae26-42b3-b7a6-01ec68d5c33a',	'2024-02-01');

-- Authorization Rule table
CREATE TABLE IF NOT EXISTS casbin_rule (
    ID UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    Ptype VARCHAR(300) CHECK (Ptype IN ('p', 'g')),
    V0 VARCHAR(300),
    V1 VARCHAR(300),
    V2 VARCHAR(300),
    V3 VARCHAR(300),
    V4 VARCHAR(300),
    V5 VARCHAR(300),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),        

    UNIQUE NULLS NOT DISTINCT (Ptype, V0, V1, V2, V3, V4, V5)
);

-- Seed Authorization Rule for Root Role Admin
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'PUBLIC', '*', '/api/session', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'PUBLIC', '*', '/api/session', 'DELETE');
INSERT INTO casbin_rule (Ptype, V0, V1, V2) VALUES ('g', '*', 'PUBLIC', '*');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'ROOT_ROLE_ADMIN', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/*', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'ROOT_ROLE_ADMIN', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/*/divisions/*', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'ROOT_ROLE_ADMIN', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/*/divisions/*/departments/*', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'ROOT_ROLE_ADMIN', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/*/users/*', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'ROOT_ROLE_ADMIN', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/*/positions/*', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'ROOT_ROLE_ADMIN', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/*/users/*/positions/*', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'ROOT_ROLE_ADMIN', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/*/roles/*', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'ROOT_ROLE_ADMIN', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/*/users/*/roles/*', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2) VALUES ('g', 'e7f31b70-ae26-42b3-b7a6-01ec68d5c33a', 'ROOT_ROLE_ADMIN', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924');

