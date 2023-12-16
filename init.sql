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

CREATE TABLE IF NOT EXISTS user (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(), -- ID used as PK to enable changes to email
    email VARCHAR(300) NOT NULL,
    tenant VARCHAR(300) NOT NULL,
    division VARCHAR(300) NOT NULL,
    department VARCHAR(300) NOT NULL,
    password CHAR(64) NOT NULL, -- SHA-256 generates a 256-bit hash value
    totp_secret_key CHAR(32) NOT NULL, --TOTP key is recommended to have 160 bits, which is 32 base32 characters
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_login TIMESTAMPTZ,

    FOREIGN KEY (tenant, division, department) REFERENCES department(tenant, division, name),
    UNIQUE(tenant, email),
);

CREATE TABLE IF NOT EXISTS appointment (
    title VARCHAR(300) NOT NULL,
    tenant VARCHAR(300) NOT NULL,
    division VARCHAR(300) NOT NULL,
    department VARCHAR(300) NOT NULL,
    user_id VARCHAR(300) NOT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (tenant, division, department, title, user_id),
    FOREIGN KEY (tenant, division, department, user_id) REFERENCES user(tenant, division, department, id)
);