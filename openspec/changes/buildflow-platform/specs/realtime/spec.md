## ADDED Requirements

### Requirement: Real-time build log streaming
The system SHALL stream build logs in real-time to connected WebSocket clients.

#### Scenario: Client connects to active build
- **WHEN** client opens WebSocket connection to /ws/builds/:id/logs for a running build
- **THEN** system sends existing log content first, then streams new log lines as they are produced

#### Scenario: Client connects to completed build
- **WHEN** client opens WebSocket to /ws/builds/:id/logs for a finished build
- **THEN** system sends the complete log content and closes the stream

#### Scenario: Client reconnects
- **WHEN** client disconnects and reconnects during an active build
- **THEN** system sends all log content from the beginning (from log file), then resumes real-time streaming

### Requirement: Real-time notification push
The system SHALL push notifications to authenticated users via WebSocket.

#### Scenario: Build completion notification
- **WHEN** a build completes (success or failure)
- **THEN** system creates notification records for relevant users and broadcasts via /ws/notifications

#### Scenario: Notification delivery
- **WHEN** user is connected to /ws/notifications
- **THEN** user receives new notifications in real-time with type, title, message, and build_id

### Requirement: WebSocket authentication
All WebSocket connections SHALL require a valid JWT Access Token.

#### Scenario: Authenticated WebSocket connection
- **WHEN** client connects to WebSocket endpoint with valid token (as query parameter or header)
- **THEN** connection is established

#### Scenario: Unauthenticated WebSocket connection
- **WHEN** client connects without valid token
- **THEN** server rejects the connection with 401 status

### Requirement: WebSocket Hub management
The system SHALL maintain a WebSocket Hub that manages client connections and message broadcasting.

#### Scenario: Client subscribes to build logs
- **WHEN** client connects to /ws/builds/:id/logs
- **THEN** Hub registers the client in the build's subscriber list

#### Scenario: Client disconnects
- **WHEN** WebSocket connection closes
- **THEN** Hub removes the client from all subscriber lists and cleans up resources

#### Scenario: Broadcast to build subscribers
- **WHEN** a new log line is produced for build ID X
- **THEN** Hub sends the log line to all clients subscribed to build X
