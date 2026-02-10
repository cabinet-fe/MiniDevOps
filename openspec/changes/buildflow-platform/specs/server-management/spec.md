## ADDED Requirements

### Requirement: Create remote server
The system SHALL allow ops and admin users to register remote servers with SSH connection details.

#### Scenario: Create server with password auth
- **WHEN** ops/admin sends POST /api/v1/servers with name, host, port, username, auth_type="password", password, description, tags
- **THEN** system creates the server record with encrypted password

#### Scenario: Create server with key auth
- **WHEN** ops/admin sends POST /api/v1/servers with auth_type="key" and private_key content
- **THEN** system creates the server record with encrypted private_key

### Requirement: List servers
The system SHALL return server list with role-based visibility.

#### Scenario: Ops/Admin lists servers
- **WHEN** ops or admin sends GET /api/v1/servers
- **THEN** system returns all servers with full details (excluding decrypted passwords/keys)

#### Scenario: Dev lists servers (read-only)
- **WHEN** dev user sends GET /api/v1/servers
- **THEN** system returns server list with only id, name, host, tags (for environment configuration dropdown)

### Requirement: Update server
The system SHALL allow ops and admin to update server configuration.

#### Scenario: Update server host
- **WHEN** ops/admin sends PUT /api/v1/servers/:id with updated host and port
- **THEN** system updates the server record

### Requirement: Delete server
The system SHALL allow ops and admin to delete servers not in use by any environment.

#### Scenario: Delete unused server
- **WHEN** ops/admin sends DELETE /api/v1/servers/:id and no environment references this server
- **THEN** system deletes the server record

#### Scenario: Delete server in use
- **WHEN** ops/admin sends DELETE /api/v1/servers/:id and environments reference this server
- **THEN** system returns 409 Conflict listing the environments using this server

### Requirement: Test server connection
The system SHALL allow testing SSH connectivity to a registered server.

#### Scenario: Successful connection test
- **WHEN** ops/admin sends POST /api/v1/servers/:id/test
- **THEN** system attempts SSH connection and returns success with server OS info

#### Scenario: Failed connection test
- **WHEN** SSH connection fails (wrong credentials, unreachable host)
- **THEN** system returns error details (connection refused, auth failed, timeout)

### Requirement: Server tags
The system SHALL support tagging servers with string labels for grouping and filtering.

#### Scenario: Filter servers by tag
- **WHEN** user sends GET /api/v1/servers?tag=production
- **THEN** system returns only servers whose tags array contains "production"
