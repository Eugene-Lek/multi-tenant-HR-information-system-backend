# Multi-Tenant HRIS Backend

This is a proof-of-concept multi-tenant HRIS backend.

## Core Functionality

This system facilitates 2 key processes:
1. **Job Requisition**
   * Requestor creates job requisition
   * Supervisor approves/rejects job requisition
   * HR approves/rejects job requisition
2. **Job Application**
   * Candidate submits job application
   * Recruiter shortlists/rejects job application
   * Recruiter sets interview date
   * Hiring manager rejects/makes an offer to the candidate (presumeably after the interview has been conducted)
   * Recruiter accepts/declines the offer on the applicant's behalf

## Supporting Functionality

1. **Two-Factor Authentication**
   * Password (a default 12 character password is generated upon user creation)
   * Time-based One-Time Password (TOTP, use an app like google authenticator to generate the codes)
2. **Authorization**
   * Both Role-Based Access Control (RBAC) and Attribute-Based Access Control (ABAC) are used
   * RBAC is used for resources with broader access (e.g. Any root role admin can create users in any tenant)
   * ABAC is used for resources with stricter access control requirements (e.g. A user can only give supervisor approval to job requisitions which they have been assigned to as the supervisor)
3. **Logging**
   * The following is logged for all events (i.e. default metadata):
     * HostName (fetched from os)
     * RequestId (auto generated uuid)
     * ClientIp (the request's remote address)
     * Url
     * Method
   * The following events are logged
     * Request completion
     * Client-side errors
     * Internal server errors (includes a traceId for debugging purposes)
     * Business events (i.e. the completion of every endpoint)
     * Security events (e.g. attempts to use a revoked session)
     * System events (e.g. Server start up)
4. **Administrative actions**
   * Create a Tenant
   * Create a Division (divisions are a sub-unit of tenants)
   * Create a Department(departments are a sub-unit of divisions)
   * Create a User
   * Create a Position
   * Assign a user to a position
   * Create a authorization role & its corresponding policies (RBAC)
   * Assign a user to an authorization role (RBAC)
   * Create a policy for a particular user (ABAC)

## Project Architecture
 * **httperror package**
   * Defines a struct representing a http error (http status, message, error code)
   * Defines an Internal Server Error constructor
 * **storage package**
   * Defines interfaces for database storage & file storage model providers
   * Defines types for those interfaces
 * **postgres package**
   * Implements the database storage interface defined in the storage package
 * **s3 package**
   * Implements the file storage interface defined in the storage package
 * **routes package**
   * Defines the routes and route handlers
   * Router handlers use the interfaces defined in the storage package
 * **main package**
   * Instantiates dependencies (e.g. postgres & s3 model providers)
   * Passes these dependencies into the router constructor to create a router
   * Starts a server with the router

## Setting up the development environment
1. Install dependencies
```
go get .
```
2. Run the postgres docker container. **(Remember to use your absolute path to init.sql instead!)**
```
docker run --name test -p 5433:5432 -e POSTGRES_PASSWORD=abcd1234 -e POSTGRES_DB=hr_information_system -v absolute/path/to/init.sql:/docker-entrypoint-initdb.d/init.sql -d postgres
```
3. In another terminal, run the server
```
go run .
```

## Running unit and integration tests
1. Run the tests (from the project root directory)
```
go test ./... -p 1 --coverprofile=coverage.out -coverpkg ./...
```
2. View the code coverage of the tests
```
go tool cover -html coverage.out
```
**Note:** If the test exits prematurely, you will have to manually delete a container called "integration_test" before re-running the test