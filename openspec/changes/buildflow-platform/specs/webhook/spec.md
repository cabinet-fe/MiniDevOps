## ADDED Requirements

### Requirement: Webhook endpoint for Git push events
The system SHALL expose a public webhook endpoint that triggers builds on Git push events.

#### Scenario: Valid webhook trigger
- **WHEN** POST /api/v1/webhook/:projectId/:secret is called with matching secret
- **THEN** system triggers a build for all environments of the project whose branch matches the pushed branch, with trigger_type="webhook"

#### Scenario: Invalid secret
- **WHEN** webhook is called with incorrect secret
- **THEN** system returns 401 Unauthorized

#### Scenario: Project not found
- **WHEN** webhook is called with nonexistent projectId
- **THEN** system returns 404 Not Found

### Requirement: Webhook secret generation
The system SHALL generate a unique webhook secret for each project upon creation.

#### Scenario: Project creation generates secret
- **WHEN** a new project is created
- **THEN** system generates a cryptographically random webhook secret (32 hex characters)

### Requirement: Branch matching
The system SHALL only trigger builds for environments whose configured branch matches the pushed branch.

#### Scenario: Branch match triggers build
- **WHEN** webhook payload indicates push to "develop" branch and environment is configured for "develop"
- **THEN** system triggers a build for that environment

#### Scenario: Branch mismatch skips
- **WHEN** webhook payload indicates push to "feature/x" but no environment is configured for that branch
- **THEN** system returns 200 OK but does not trigger any build

### Requirement: Webhook URL display
The system SHALL display the webhook URL in the project settings page for user to copy and configure in Git platform.

#### Scenario: View webhook URL
- **WHEN** user views project detail or edit page
- **THEN** system displays the full webhook URL: {base_url}/api/v1/webhook/{projectId}/{secret}
