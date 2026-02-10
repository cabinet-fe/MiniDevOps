## ADDED Requirements

### Requirement: Audit log recording
The system SHALL automatically record audit logs for all state-changing API operations via middleware.

#### Scenario: Record project creation
- **WHEN** user creates a project via POST /api/v1/projects
- **THEN** system records audit_log with action="project.create", resource_type="project", resource_id=new project ID, user_id, ip_address, and details (JSON with project name)

#### Scenario: Record build trigger
- **WHEN** user triggers a build
- **THEN** system records audit_log with action="build.trigger", resource_type="build", resource_id=build ID

#### Scenario: Record user management actions
- **WHEN** admin creates, updates, or deletes a user
- **THEN** system records audit_log with action="user.create|update|delete"

### Requirement: Audit log actions coverage
The system SHALL record the following actions: user.create, user.update, user.delete, project.create, project.update, project.delete, project.import, project.export, environment.create, environment.update, environment.delete, server.create, server.update, server.delete, build.trigger, build.cancel, build.deploy, build.rollback, system.backup, system.restore.

#### Scenario: All critical operations logged
- **WHEN** any of the listed operations is performed
- **THEN** an audit_log entry is created with the corresponding action string

### Requirement: Query audit logs
The system SHALL allow admin and ops users to query audit logs with filters.

#### Scenario: Filter by action type
- **WHEN** admin sends GET /api/v1/system/audit-logs?action=build.trigger
- **THEN** system returns only audit logs with action="build.trigger"

#### Scenario: Filter by user
- **WHEN** admin sends GET /api/v1/system/audit-logs?user_id=3
- **THEN** system returns only audit logs for user_id=3

#### Scenario: Filter by date range
- **WHEN** admin sends GET /api/v1/system/audit-logs?from=2026-01-01&to=2026-02-01
- **THEN** system returns audit logs within the date range

#### Scenario: Paginated results
- **WHEN** query returns more than page_size results
- **THEN** system returns paginated results ordered by created_at DESC
