## ADDED Requirements

### Requirement: Pluggable deployment strategies
The system SHALL support three deployment methods: rsync, SFTP, and SCP, selectable per environment configuration.

#### Scenario: Deploy via rsync
- **WHEN** environment's deploy_method is "rsync" and target server is Linux
- **THEN** system executes rsync to synchronize build output to remote server's deploy_path

#### Scenario: Deploy via SFTP
- **WHEN** environment's deploy_method is "sftp"
- **THEN** system uploads build output files to remote server's deploy_path via SFTP protocol

#### Scenario: Deploy via SCP
- **WHEN** environment's deploy_method is "scp"
- **THEN** system copies build output archive to remote server's deploy_path via SCP

### Requirement: Deployment after successful build
The system SHALL automatically deploy after a successful build if the environment has deployment configured.

#### Scenario: Auto-deploy on build success
- **WHEN** build completes successfully and environment has deploy_server_id and deploy_path configured
- **THEN** system transitions build status to "deploying" and executes deployment, then transitions to "success"

#### Scenario: Skip deployment when not configured
- **WHEN** build succeeds but environment has no deployment configuration
- **THEN** system skips deployment and sets status to "success" directly

### Requirement: Post-deployment script execution
The system SHALL execute a configured post-deploy script on the remote server via SSH after file deployment.

#### Scenario: Execute post-deploy script
- **WHEN** deployment completes and environment has post_deploy_script configured
- **THEN** system connects to remote server via SSH and executes the script, capturing output to build log

#### Scenario: Post-deploy script failure
- **WHEN** post-deploy script exits with non-zero code
- **THEN** system sets build status to "failed" and records the script error output

### Requirement: Manual deploy/redeploy
The system SHALL allow users to manually trigger deployment of a successful build.

#### Scenario: Manual deploy
- **WHEN** user sends POST /api/v1/builds/:id/deploy for a build with status "success" and existing artifact
- **THEN** system re-executes the deployment pipeline (file transfer + post-deploy script)

### Requirement: Deployment logging
The system SHALL log all deployment operations (file transfer progress, remote script output) to the build log.

#### Scenario: Deployment log output
- **WHEN** deployment is executing
- **THEN** rsync/sftp/scp progress and post-deploy script output are appended to build log and broadcast via WebSocket
