## ADDED Requirements

### Requirement: Create project
The system SHALL allow authenticated users to create projects with repository configuration.

#### Scenario: Create project with token auth
- **WHEN** user sends POST /api/v1/projects with name, description, repo_url, repo_auth_type="token", repo_username, repo_password, max_artifacts
- **THEN** system creates the project, encrypts repo_password, generates a webhook_secret, and sets created_by to current user

#### Scenario: Create project with no auth
- **WHEN** user sends POST /api/v1/projects with repo_auth_type="none" (public repo)
- **THEN** system creates the project without repo credentials

#### Scenario: Duplicate project name
- **WHEN** user creates a project with a name that already exists
- **THEN** system returns 409 Conflict

### Requirement: List projects
The system SHALL return projects based on user role: admin and ops see all projects, dev sees only their own.

#### Scenario: Admin lists all projects
- **WHEN** admin sends GET /api/v1/projects
- **THEN** system returns all projects with pagination

#### Scenario: Dev lists own projects
- **WHEN** dev user sends GET /api/v1/projects
- **THEN** system returns only projects where created_by matches the user

### Requirement: Get project detail
The system SHALL return project details including associated environments.

#### Scenario: Get project with environments
- **WHEN** user sends GET /api/v1/projects/:id
- **THEN** system returns project info with list of environments and latest build status per environment

### Requirement: Update project
The system SHALL allow the project owner, ops, or admin to update project configuration.

#### Scenario: Owner updates project
- **WHEN** project owner sends PUT /api/v1/projects/:id with updated fields
- **THEN** system updates the project (re-encrypts password if changed)

#### Scenario: Dev updates other's project
- **WHEN** dev user (not owner) sends PUT /api/v1/projects/:id
- **THEN** system returns 403 Forbidden

### Requirement: Delete project
The system SHALL allow the project owner, ops, or admin to delete a project and all associated data.

#### Scenario: Delete project cascades
- **WHEN** authorized user sends DELETE /api/v1/projects/:id
- **THEN** system deletes the project, its environments, builds, artifacts, and workspace directory

### Requirement: Manage project environments
The system SHALL allow creating, updating, and deleting environments as sub-configurations of a project.

#### Scenario: Create environment
- **WHEN** authorized user sends POST /api/v1/projects/:id/envs with name, branch, build_script, build_output_dir, deploy_server_id, deploy_path, deploy_method, post_deploy_script, env_vars
- **THEN** system creates the environment linked to the project

#### Scenario: Duplicate environment name
- **WHEN** user creates an environment with a name already used in the same project
- **THEN** system returns 409 Conflict

#### Scenario: Update environment
- **WHEN** authorized user sends PUT /api/v1/projects/:id/envs/:envId with updated fields
- **THEN** system updates the environment configuration

#### Scenario: Delete environment
- **WHEN** authorized user sends DELETE /api/v1/projects/:id/envs/:envId
- **THEN** system deletes the environment and associated builds

### Requirement: Environment configuration fields
Each environment SHALL support: name, branch (git branch), build_script (shell commands), build_output_dir (relative path), deploy_server_id (from server pool, nullable), deploy_path (remote directory), deploy_method (rsync/sftp/scp), post_deploy_script (SSH commands), env_vars (JSON key-value pairs), sort_order.

#### Scenario: Environment with full config
- **WHEN** environment is created with all fields including deploy configuration
- **THEN** builds in this environment use the specified branch, script, deployment target, and environment variables

#### Scenario: Environment without deployment
- **WHEN** environment is created without deploy_server_id and deploy_path
- **THEN** builds only execute build script and store artifacts without deploying

### Requirement: Export project configuration
The system SHALL allow admin users to export a project's configuration as JSON.

#### Scenario: Export project
- **WHEN** admin sends GET /api/v1/projects/:id/export
- **THEN** system returns JSON containing project config + environments + env_vars (sensitive fields like passwords excluded)

### Requirement: Import project configuration
The system SHALL allow admin users to import a project from a JSON configuration.

#### Scenario: Import project
- **WHEN** admin sends POST /api/v1/projects/import with valid JSON body
- **THEN** system creates the project and environments from the imported configuration, leaving sensitive fields blank for manual input
