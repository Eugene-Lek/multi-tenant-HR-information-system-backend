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

    FOREIGN KEY (tenant_id) REFERENCES tenant(id),    
    FOREIGN KEY (department_id) REFERENCES department(id) -- Every position must correspond to a department
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

CREATE TYPE APPROVAL_STATUS AS ENUM (
    'PENDING',
    'APPROVED',
    'REJECTED'
);

CREATE TABLE IF NOT EXISTS job_requisition(
    id UUID PRIMARY KEY NOT NULL,
    tenant_id UUID NOT NULL,
    position_id UUID, -- Null if the job requisition is for a new position. Only filled in after the new position has been created
    title VARCHAR(300), -- Null if the job requisition is for an existing position
    department_id UUID, -- Null if the job requisition is for an existing position
    supervisor_position_ids UUID[], -- Null if the job requisition is for an existing position
    job_description TEXT NOT NULL,
    job_requirements TEXT NOT NULL,
    requestor UUID NOT NULL, 
    supervisor UUID NOT NULL, 
    supervisor_decision APPROVAL_STATUS NOT NULL DEFAULT 'PENDING',
    hr_approver UUID NOT NULL,
    hr_approver_decision APPROVAL_STATUS NOT NULL DEFAULT 'PENDING',    
    recruiter UUID,
    filled_by UUID, 
    filled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    FOREIGN KEY (tenant_id) REFERENCES tenant(id),
    FOREIGN KEY (department_id) REFERENCES department(id),
    FOREIGN KEY (position_id) REFERENCES position(id),
    FOREIGN KEY (requestor) REFERENCES user_account(id),
    FOREIGN KEY (supervisor) REFERENCES user_account(id),
    FOREIGN KEY (hr_approver) REFERENCES user_account(id),    
    FOREIGN KEY (recruiter) REFERENCES user_account(id),      
    FOREIGN KEY (filled_by) REFERENCES user_account(id),    

    -- Prevents hr from approving if supervisor has not approved/has rejected
    CONSTRAINT ck_hr_approval_only_with_supervisor_approval 
        CHECK ( NOT (supervisor_decision <> 'APPROVED' AND hr_approver_decision = 'APPROVED')),
    -- Ensures that recruiter is provided if HR has approved    
    CONSTRAINT ck_recruiter_assignment_made_if_have_hr_approval
        CHECK ( NOT (hr_approver_decision = 'APPROVED' AND recruiter IS NULL) ),      
    -- Prevents job aquisition from being filled if hr has not approved    
    CONSTRAINT ck_req_filled_only_with_hr_approval 
        CHECK ( NOT (hr_approver_decision <> 'APPROVED' AND filled_by IS NOT NULL) ),      
    CONSTRAINT ck_req_filled_at_only_with_hr_approval 
        CHECK ( NOT (hr_approver_decision <> 'APPROVED' AND filled_at IS NOT NULL) )         
);

CREATE TYPE SHORTLIST_STATUS AS ENUM (
    'PENDING',
    'SHORTLISTED',
    'REJECTED'
);

CREATE TYPE OFFER_STATUS AS ENUM (
    'PENDING',
    'OFFERED',
    'REJECTED',
    'RESCINDED'
);

CREATE TYPE ACCEPT_STATUS AS ENUM (
    'PENDING',
    'ACCEPTED',
    'REJECTED',
    'RESCINDED'
);

CREATE TABLE IF NOT EXISTS job_application (
    id UUID PRIMARY KEY NOT NULL,
    tenant_id UUID NOT NULL,
    job_requisition_id UUID NOT NULL,    
    first_name VARCHAR(300) NOT NULL,
    last_name VARCHAR(300) NOT NULL,
    country_code INTEGER NOT NULL,
    phone_number INTEGER NOT NULL,
    email VARCHAR(300) NOT NULL,
    resume_s3_url TEXT NOT NULL, 
    recruiter_decision SHORTLIST_STATUS NOT NULL DEFAULT 'PENDING',
    interview_date DATE,
    hiring_manager_decision OFFER_STATUS NOT NULL DEFAULT 'PENDING',
    offer_start_date DATE,
    offer_end_date DATE,    
    applicant_decision ACCEPT_STATUS NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (email, job_requisition_id),
    UNIQUE (country_code, phone_number, job_requisition_id),

    FOREIGN KEY (tenant_id) REFERENCES tenant(id),
    FOREIGN KEY (job_requisition_id) REFERENCES job_requisition(id),

    -- Prevents the interview date from being set without recruiter shortlisting
    CONSTRAINT ck_recruiter_shortlist_before_setting_interview_date
        CHECK ( NOT (recruiter_decision <> 'SHORTLISTED' AND interview_date IS NOT NULL)),
    -- Prevents the supervisor from making an offer before the interview date has been set
    CONSTRAINT ck_interview_date_set_before_hiring_manager_offer
        CHECK ( NOT ( interview_date IS NULL AND hiring_manager_decision = 'OFFERED')),   
    -- Prevents the applicant from accepting the offer before the hiring manager has made the offer
    CONSTRAINT ck_hiring_manager_offer_before_applicant_acceptance
        CHECK ( NOT ( hiring_manager_decision <> 'OFFERED' AND hiring_manager_decision <> 'RESCINDED' AND applicant_decision = 'ACCEPTED'))
);

-- Credentials of all user accounts
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

-- Test Supervisor Account & position
INSERT INTO user_account (id, tenant_id, email, password, totp_secret_key) 
VALUES ('38d3f831-9a9e-4dfc-ba56-ec68bf2462e0', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', 'administration-manager@hrisEnterprises.org',
'$argon2id$v=19$m=65536,t=1,p=8$cFTNg+YXrN4U0lvwnamPkg$0RDBxH+EouVxDbBlQUNctdWZ+CNKrayPpzTJaWNq83U', 
'OLDFXRMH35A3DU557UXITHYDK4SKLTXZ');

INSERT INTO position (id, tenant_id, title, department_id)
VALUES ('0c55ff72-a23d-440b-b77f-db6b8002f734', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', 'Manager', '9147b727-1955-437b-be7d-785e9a31f20c');

INSERT INTO position_assignment (tenant_id, position_id, user_account_id, start_date)
VALUES ('2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '0c55ff72-a23d-440b-b77f-db6b8002f734', '38d3f831-9a9e-4dfc-ba56-ec68bf2462e0',	'2024-02-01');

-- Test Suboordinate-Supervisor relationship
INSERT INTO subordinate_supervisor_relationship (subordinate_position_id, supervisor_position_id)
VALUES ('e4edbd37-164d-478d-9625-5b1397ef6e45', '0c55ff72-a23d-440b-b77f-db6b8002f734');

-- Test HR Account
INSERT INTO user_account (id, tenant_id, email, password, totp_secret_key) 
VALUES ('9f4c9dd0-7c75-4ea9-a106-948885b6bedf', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', 'hr-director@hrisEnterprises.org',
'$argon2id$v=19$m=65536,t=1,p=8$cFTNg+YXrN4U0lvwnamPkg$0RDBxH+EouVxDbBlQUNctdWZ+CNKrayPpzTJaWNq83U', 
'OLDFXRMH35A3DU557UXITHYDK4SKLTXZ');

-- Test Recruiter Account
INSERT INTO user_account (id, tenant_id, email, password, totp_secret_key) 
VALUES ('ccb2da3b-68ac-419e-b95d-dd6b723035f9', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', 'hr-recruiter@hrisEnterprises.org',
'$argon2id$v=19$m=65536,t=1,p=8$cFTNg+YXrN4U0lvwnamPkg$0RDBxH+EouVxDbBlQUNctdWZ+CNKrayPpzTJaWNq83U', 
'OLDFXRMH35A3DU557UXITHYDK4SKLTXZ');

-- Test Job Requisition
INSERT INTO job_requisition (id, tenant_id, title, department_id, supervisor_position_ids, job_description, job_requirements, requestor, supervisor, hr_approver)
VALUES ('5062a285-e82b-475d-8113-daefd05dcd90', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924',
'Database Administrator', '9147b727-1955-437b-be7d-785e9a31f20c', '{0c55ff72-a23d-440b-b77f-db6b8002f734}',
'Manages databases of HRIS software', '100 years of experience using postgres', 
'e7f31b70-ae26-42b3-b7a6-01ec68d5c33a', '38d3f831-9a9e-4dfc-ba56-ec68bf2462e0', '9f4c9dd0-7c75-4ea9-a106-948885b6bedf');

-- Test Job Application
INSERT INTO job_application (id, tenant_id, job_requisition_id, first_name, last_name, country_code, phone_number, email, resume_s3_url)
VALUES ('5062a285-e82b-475d-8113-daefd05dcd90', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '5062a285-e82b-475d-8113-daefd05dcd90',
'Eugene', 'Lek', '1', '123456789', 'test@gmail.com', '');

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
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'PUBLIC', '*', '/api/tenants/{tenantId}/job-applications/{jobApplicationId}', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2) VALUES ('g', '*', 'PUBLIC', '*');

INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'ROOT_ROLE_ADMIN', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/{tenantId}', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'ROOT_ROLE_ADMIN', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/{tenantId}/divisions/{divisionId}', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'ROOT_ROLE_ADMIN', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/{tenantId}/divisions/{divisionId}/departments/{departmentId}', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'ROOT_ROLE_ADMIN', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/{tenantId}/users/{userId}', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'ROOT_ROLE_ADMIN', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/{tenantId}/positions/{positionId}', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'ROOT_ROLE_ADMIN', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/{tenantId}/users/{userId}/positions/{positionId}', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'ROOT_ROLE_ADMIN', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/{tenantId}/policies', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'ROOT_ROLE_ADMIN', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/{tenantId}/users/{userId}/roles/{roleId}', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2) VALUES ('g', 'e7f31b70-ae26-42b3-b7a6-01ec68d5c33a', 'ROOT_ROLE_ADMIN', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924');

-- Default user rights
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'e7f31b70-ae26-42b3-b7a6-01ec68d5c33a', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/2ad1dcfc-8867-49f7-87a3-8bd8d1154924/users/e7f31b70-ae26-42b3-b7a6-01ec68d5c33a/job-requisitions/role-requestor/{jobReqId}', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'e7f31b70-ae26-42b3-b7a6-01ec68d5c33a', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/2ad1dcfc-8867-49f7-87a3-8bd8d1154924/users/e7f31b70-ae26-42b3-b7a6-01ec68d5c33a/job-requisitions/role-requestor/{jobReqId}/job-applications/{jobAppId}/hiring-manager-decision', 'POST');

-- Default superviosr rights
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', '38d3f831-9a9e-4dfc-ba56-ec68bf2462e0', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/2ad1dcfc-8867-49f7-87a3-8bd8d1154924/users/38d3f831-9a9e-4dfc-ba56-ec68bf2462e0/job-requisitions/role-supervisor/{jobReqId}/supervisor-approval', 'POST');

-- Default hr approver rights
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', '9f4c9dd0-7c75-4ea9-a106-948885b6bedf', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/2ad1dcfc-8867-49f7-87a3-8bd8d1154924/users/9f4c9dd0-7c75-4ea9-a106-948885b6bedf/job-requisitions/role-supervisor/{jobReqId}/supervisor-approval', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', '9f4c9dd0-7c75-4ea9-a106-948885b6bedf', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/2ad1dcfc-8867-49f7-87a3-8bd8d1154924/users/9f4c9dd0-7c75-4ea9-a106-948885b6bedf/job-requisitions/role-hr-approver/{jobReqId}/hr-approval', 'POST');

-- Default recruiter
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'ccb2da3b-68ac-419e-b95d-dd6b723035f9', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/2ad1dcfc-8867-49f7-87a3-8bd8d1154924/users/ccb2da3b-68ac-419e-b95d-dd6b723035f9/job-requisitions/role-recruiter/{jobReqId}/job-applications/{jobAppId}/recruiter-decision', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'ccb2da3b-68ac-419e-b95d-dd6b723035f9', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/2ad1dcfc-8867-49f7-87a3-8bd8d1154924/users/ccb2da3b-68ac-419e-b95d-dd6b723035f9/job-requisitions/role-recruiter/{jobReqId}/job-applications/{jobAppId}/interview-date', 'POST');
INSERT INTO casbin_rule (Ptype, V0, V1, V2, V3) VALUES ('p', 'ccb2da3b-68ac-419e-b95d-dd6b723035f9', '2ad1dcfc-8867-49f7-87a3-8bd8d1154924', '/api/tenants/2ad1dcfc-8867-49f7-87a3-8bd8d1154924/users/ccb2da3b-68ac-419e-b95d-dd6b723035f9/job-requisitions/role-recruiter/{jobReqId}/job-applications/{jobAppId}/applicant-decision', 'POST');
