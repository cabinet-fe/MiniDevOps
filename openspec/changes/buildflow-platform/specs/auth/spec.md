## ADDED Requirements

### Requirement: User login with username and password
The system SHALL authenticate users via username and password, issuing a JWT Access Token (2h TTL) and a Refresh Token (7d TTL, stored in HttpOnly Cookie).

#### Scenario: Successful login
- **WHEN** user submits valid username and password to POST /api/v1/auth/login
- **THEN** system returns Access Token in response body and sets Refresh Token as HttpOnly Cookie

#### Scenario: Invalid credentials
- **WHEN** user submits incorrect username or password
- **THEN** system returns 401 Unauthorized with error message "用户名或密码错误"

#### Scenario: Disabled account
- **WHEN** user with is_active=false attempts to login
- **THEN** system returns 403 Forbidden with error message "账户已被禁用"

### Requirement: Token refresh
The system SHALL allow refreshing an expired Access Token using a valid Refresh Token.

#### Scenario: Successful refresh
- **WHEN** client sends POST /api/v1/auth/refresh with valid Refresh Token cookie
- **THEN** system returns a new Access Token and rotates the Refresh Token

#### Scenario: Expired refresh token
- **WHEN** client sends refresh request with expired Refresh Token
- **THEN** system returns 401 and client MUST re-login

### Requirement: User logout
The system SHALL invalidate the current session on logout.

#### Scenario: Successful logout
- **WHEN** authenticated user sends POST /api/v1/auth/logout
- **THEN** system clears the Refresh Token cookie

### Requirement: Get current user info
The system SHALL return the authenticated user's profile information.

#### Scenario: Authenticated request
- **WHEN** authenticated user sends GET /api/v1/auth/me
- **THEN** system returns user id, username, display_name, role, email, avatar

#### Scenario: Unauthenticated request
- **WHEN** request is sent without valid Access Token
- **THEN** system returns 401 Unauthorized

### Requirement: RBAC middleware
The system SHALL enforce role-based access control on all protected API endpoints. Roles are: admin, ops, dev.

#### Scenario: Sufficient permissions
- **WHEN** user with role "ops" accesses an endpoint requiring "ops" or "admin"
- **THEN** system allows the request to proceed

#### Scenario: Insufficient permissions
- **WHEN** user with role "dev" accesses an endpoint requiring "admin" only
- **THEN** system returns 403 Forbidden

### Requirement: JWT structure
The JWT Access Token payload SHALL contain: sub (user ID), username, role, exp (expiration timestamp).

#### Scenario: Token payload
- **WHEN** system issues an Access Token
- **THEN** the token payload contains sub, username, role, and exp fields
