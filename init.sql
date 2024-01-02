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
    name VARCHAR(300) PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS division (
    name VARCHAR(300) NOT NULL,
    tenant VARCHAR(300) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (tenant, name),
    FOREIGN KEY (tenant) REFERENCES tenant(name)
);

CREATE TABLE IF NOT EXISTS department (
    name VARCHAR(300) NOT NULL,
    tenant VARCHAR(300) NOT NULL,
    division VARCHAR(300) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (tenant, division, name),
    FOREIGN KEY (tenant, division) REFERENCES division(tenant, name)
);

CREATE TABLE IF NOT EXISTS user_account (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(), -- ID used as PK to enable changes to email
    email VARCHAR(300) NOT NULL,
    tenant VARCHAR(300) NOT NULL,
    password TEXT NOT NULL,
    totp_secret_key CHAR(32) NOT NULL, --TOTP key is recommended to have 160 bits, which is 32 base32 characters
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_login TIMESTAMPTZ,

    FOREIGN KEY (tenant) REFERENCES tenant(name),
    UNIQUE(tenant, email)
);

CREATE TABLE IF NOT EXISTS appointment (
    title VARCHAR(300) NOT NULL,
    tenant VARCHAR(300) NOT NULL,
    division VARCHAR(300) NOT NULL,
    department VARCHAR(300) NOT NULL,
    user_account_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL DEFAULT'9999-12-31',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (tenant, division, department, title, user_account_id),
    FOREIGN KEY (user_account_id) REFERENCES user_account(id), -- Every appointment must correspond to a user_account
    FOREIGN KEY (tenant, division, department) REFERENCES department(tenant, division, name) -- Every appointment must correspond to a department
);

-- Seed of root role administrator
-- Password: jU%q837d!QP7
-- Totp Key: OLDFXRMH35A3DU557UXITHYDK4SKLTXZ
INSERT INTO tenant (name) VALUES ('HRIS Enterprises');

INSERT INTO user_account (id, email, tenant, password, totp_secret_key) 
VALUES ('e7f31b70-ae26-42b3-b7a6-01ec68d5c33a', 'root-role-admin@hrisEnterprises.org', 'HRIS Enterprises',
'$argon2id$v=19$m=65536,t=1,p=8$cFTNg+YXrN4U0lvwnamPkg$0RDBxH+EouVxDbBlQUNctdWZ+CNKrayPpzTJaWNq83U', 
'OLDFXRMH35A3DU557UXITHYDK4SKLTXZ');


-- Authorization Rule table
CREATE TABLE IF NOT EXISTS casbin_rule (
    ID UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    Ptype VARCHAR(300),
    V0 VARCHAR(300),
    V1 VARCHAR(300),
    V2 VARCHAR(300),
    V3 VARCHAR(300),
    V4 VARCHAR(300),
    V5 VARCHAR(300)                   
);

-- Seed Authorization Rule for Root Role Admin
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'PUBLIC', '*', '/api/session', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'PUBLIC', '*', '/api/session', 'DELETE');
INSERT INTO casbin_rule (Ptype, V0, V1, V2) VALUES ('g', '*', 'PUBLIC', '*');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'ROOT_ROLE_ADMIN', 'HRIS Enterprises', '/api/tenants/*', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'ROOT_ROLE_ADMIN', 'HRIS Enterprises', '/api/tenants/*/divisions/*', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'ROOT_ROLE_ADMIN', 'HRIS Enterprises', '/api/tenants/*/divisions/*/departments/*', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'ROOT_ROLE_ADMIN', 'HRIS Enterprises', '/api/tenants/*/users/*', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'ROOT_ROLE_ADMIN', 'HRIS Enterprises', '/api/tenants/*/users/*/appointments/*', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2) VALUES ('g', 'e7f31b70-ae26-42b3-b7a6-01ec68d5c33a', 'ROOT_ROLE_ADMIN', 'HRIS Enterprises');

