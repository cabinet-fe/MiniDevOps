## ADDED Requirements

### Requirement: System backup export
The system SHALL allow admin users to export a full system backup as a tar.gz archive.

#### Scenario: Export backup
- **WHEN** admin sends POST /api/v1/system/backup
- **THEN** system packages the SQLite database file and config.yaml into a tar.gz and returns it as a download

#### Scenario: Non-admin denied
- **WHEN** non-admin user sends POST /api/v1/system/backup
- **THEN** system returns 403 Forbidden

### Requirement: System restore from backup
The system SHALL allow admin users to restore from a previously exported backup archive.

#### Scenario: Restore from backup
- **WHEN** admin sends POST /api/v1/system/restore with a valid tar.gz backup file
- **THEN** system extracts and replaces the SQLite database file, then restarts the database connection

#### Scenario: Invalid backup file
- **WHEN** admin uploads a file that is not a valid backup archive
- **THEN** system returns 400 Bad Request with error message

### Requirement: Project export as JSON
The system SHALL allow admin users to export a project's configuration as a JSON file.

#### Scenario: Export project config
- **WHEN** admin sends GET /api/v1/projects/:id/export
- **THEN** system returns JSON containing: project fields (name, description, repo_url, repo_auth_type, max_artifacts), environments (all fields except IDs), with sensitive fields (passwords, tokens, keys) replaced by empty strings

### Requirement: Project import from JSON
The system SHALL allow admin users to import a project from an exported JSON file.

#### Scenario: Import project config
- **WHEN** admin sends POST /api/v1/projects/import with valid project JSON
- **THEN** system creates the project and all environments from the JSON, generating new IDs, with sensitive fields left blank for manual configuration

#### Scenario: Import with name conflict
- **WHEN** imported project name already exists
- **THEN** system appends a suffix (e.g., "-imported-1") to avoid conflict
