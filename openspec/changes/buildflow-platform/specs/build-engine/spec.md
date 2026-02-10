## ADDED Requirements

### Requirement: Trigger build manually
The system SHALL allow authenticated users to trigger a build for a specific project environment.

#### Scenario: Trigger build
- **WHEN** user sends POST /api/v1/projects/:id/builds with environment_id
- **THEN** system creates a build record with status="pending", assigns build_number (auto-increment per project), sets trigger_type="manual" and triggered_by to current user, and submits it to the scheduler

#### Scenario: Trigger build for nonexistent environment
- **WHEN** user specifies an environment_id not belonging to the project
- **THEN** system returns 400 Bad Request

### Requirement: Build scheduler with concurrency control
The system SHALL schedule builds using a goroutine pool with configurable max concurrent builds.

#### Scenario: Under concurrency limit
- **WHEN** a build is submitted and active builds < max_concurrent
- **THEN** build starts executing immediately

#### Scenario: At concurrency limit
- **WHEN** a build is submitted and active builds >= max_concurrent
- **THEN** build stays in "pending" status until a slot becomes available

### Requirement: Build pipeline execution
The system SHALL execute builds through a sequential pipeline: clone/pull → checkout branch → clean → inject env vars → run build script → collect artifact.

#### Scenario: Full build pipeline success
- **WHEN** build starts executing
- **THEN** system transitions through statuses: pending → cloning → building → (deploying if configured) → success, capturing commit_hash and commit_message from git

#### Scenario: First build clones repository
- **WHEN** workspace directory for the project does not exist
- **THEN** system executes git clone with configured credentials

#### Scenario: Subsequent build pulls updates
- **WHEN** workspace directory already exists
- **THEN** system executes git fetch → git checkout <branch> → git reset --hard origin/<branch>, preserving dependency caches (node_modules, vendor, etc.)

### Requirement: Build workspace cleanup
The system SHALL clean the workspace before building while preserving dependency caches.

#### Scenario: Clean workspace preserving dependencies
- **WHEN** build enters the "cloning" phase on existing workspace
- **THEN** system runs git clean to remove untracked files, excluding common dependency directories (node_modules, vendor, .gradle, target, __pycache__)

### Requirement: Environment variable injection
The system SHALL inject configured environment variables into the build script execution context.

#### Scenario: Build with environment variables
- **WHEN** build script executes and environment has env_vars configured
- **THEN** all key-value pairs from env_vars are set as environment variables for the build process

### Requirement: Build log capture
The system SHALL capture build script stdout and stderr line by line, writing to both a log file and WebSocket broadcast.

#### Scenario: Build log streaming
- **WHEN** build script outputs text to stdout or stderr
- **THEN** each line is written to data/logs/project-{id}/build-{number}.log AND broadcast to connected WebSocket clients on /ws/builds/:id/logs

### Requirement: Build artifact collection
The system SHALL package the build output directory into a tar.gz archive after successful build.

#### Scenario: Collect artifact
- **WHEN** build script completes with exit code 0 and build_output_dir is configured
- **THEN** system archives the output directory to data/artifacts/project-{id}/build-{number}.tar.gz

#### Scenario: No output directory configured
- **WHEN** build succeeds but build_output_dir is empty
- **THEN** system marks build as success without collecting artifact

### Requirement: Artifact retention policy
The system SHALL enforce the project's max_artifacts limit by deleting oldest artifacts.

#### Scenario: Artifact cleanup
- **WHEN** a new artifact is stored and total artifacts for this project exceed max_artifacts
- **THEN** system deletes the oldest artifact files and updates corresponding build records (artifact_path set to null)

### Requirement: Download build artifact
The system SHALL allow authenticated users to download a build's artifact archive.

#### Scenario: Download existing artifact
- **WHEN** user sends GET /api/v1/builds/:id/artifact and artifact exists
- **THEN** system returns the tar.gz file as a download

#### Scenario: Artifact not available
- **WHEN** user requests download but artifact has been cleaned up
- **THEN** system returns 404 Not Found

### Requirement: Cancel build
The system SHALL allow users to cancel a pending, cloning, or building build.

#### Scenario: Cancel running build
- **WHEN** user sends POST /api/v1/builds/:id/cancel for a build in "building" status
- **THEN** system kills the build process, sets status to "cancelled"

#### Scenario: Cancel completed build
- **WHEN** user attempts to cancel a build with status "success" or "failed"
- **THEN** system returns 400 Bad Request

### Requirement: Build history
The system SHALL provide paginated build history for a project, optionally filtered by environment.

#### Scenario: List builds for project
- **WHEN** user sends GET /api/v1/projects/:id/builds with optional environment_id filter
- **THEN** system returns paginated builds ordered by created_at DESC

### Requirement: Build detail
The system SHALL return complete build information including log content.

#### Scenario: Get build detail
- **WHEN** user sends GET /api/v1/builds/:id
- **THEN** system returns build record with status, commit info, duration, error message, and log content

### Requirement: Rollback to previous build
The system SHALL allow ops and admin to redeploy a historical build artifact.

#### Scenario: Rollback deployment
- **WHEN** ops/admin sends POST /api/v1/builds/:id/rollback for a build with existing artifact
- **THEN** system deploys the historical artifact to the environment's configured server using the configured deploy method, creating a new build record with trigger_type="rollback"

#### Scenario: Rollback without artifact
- **WHEN** user attempts rollback for a build whose artifact has been cleaned up
- **THEN** system returns 400 Bad Request with message "构建产物已被清理，无法回滚"

### Requirement: Build failure handling
The system SHALL properly handle and record build failures at each pipeline stage.

#### Scenario: Git clone failure
- **WHEN** git clone/pull fails (auth error, network issue)
- **THEN** system sets build status to "failed", records error_message, and logs the git error output

#### Scenario: Build script failure
- **WHEN** build script exits with non-zero code
- **THEN** system sets build status to "failed", records error_message with exit code, and preserves full log
