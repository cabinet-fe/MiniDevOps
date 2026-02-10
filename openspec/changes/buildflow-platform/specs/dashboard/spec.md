## ADDED Requirements

### Requirement: Dashboard statistics
The system SHALL provide summary statistics for the dashboard page.

#### Scenario: Get dashboard stats
- **WHEN** authenticated user sends GET /api/v1/dashboard/stats
- **THEN** system returns: total_projects (count), today_builds (builds created today), success_rate (percentage of successful builds in last 7 days), active_builds (currently running builds count)

### Requirement: Active builds list
The system SHALL provide a list of currently running builds with progress indication.

#### Scenario: List active builds
- **WHEN** user sends GET /api/v1/dashboard/active-builds
- **THEN** system returns builds with status in ("pending", "cloning", "building", "deploying"), including project name, environment name, status, triggered_by, and started_at

#### Scenario: No active builds
- **WHEN** no builds are currently running
- **THEN** system returns an empty list

### Requirement: Recent builds list
The system SHALL provide a paginated list of recent builds across all projects.

#### Scenario: List recent builds
- **WHEN** user sends GET /api/v1/dashboard/recent-builds with optional limit (default 20)
- **THEN** system returns recent builds ordered by created_at DESC, including project name, environment name, status, trigger_type, triggered_by username, duration_ms, and created_at

### Requirement: Build trend data
The system SHALL provide build count data grouped by day for charting.

#### Scenario: Get build trend
- **WHEN** user sends GET /api/v1/dashboard/stats with period=7d
- **THEN** system returns daily build counts for the last 7 days, grouped by status (success, failed, cancelled)
