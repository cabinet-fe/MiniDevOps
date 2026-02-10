## ADDED Requirements

### Requirement: Admin can list all users
The system SHALL allow admin users to retrieve a paginated list of all users.

#### Scenario: Admin lists users
- **WHEN** admin sends GET /api/v1/users with optional page and page_size params
- **THEN** system returns paginated user list (id, username, display_name, role, email, is_active, created_at)

#### Scenario: Non-admin access denied
- **WHEN** user with role "ops" or "dev" sends GET /api/v1/users
- **THEN** system returns 403 Forbidden

### Requirement: Admin can create users
The system SHALL allow admin users to create new user accounts.

#### Scenario: Successful creation
- **WHEN** admin sends POST /api/v1/users with username, password, display_name, role, email
- **THEN** system creates the user with bcrypt-hashed password and returns user info

#### Scenario: Duplicate username
- **WHEN** admin creates a user with an already existing username
- **THEN** system returns 409 Conflict with error message

### Requirement: Admin can update users
The system SHALL allow admin users to update any user's profile, role, and active status.

#### Scenario: Update user role
- **WHEN** admin sends PUT /api/v1/users/:id with role="ops"
- **THEN** system updates the user's role

#### Scenario: Deactivate user
- **WHEN** admin sends PUT /api/v1/users/:id with is_active=false
- **THEN** system deactivates the user (they can no longer login)

### Requirement: Admin can delete users
The system SHALL allow admin users to delete user accounts.

#### Scenario: Delete non-admin user
- **WHEN** admin sends DELETE /api/v1/users/:id for a non-admin user
- **THEN** system deletes the user record

#### Scenario: Prevent self-deletion
- **WHEN** admin attempts to delete their own account
- **THEN** system returns 400 Bad Request with error message

### Requirement: User can update own profile
The system SHALL allow any authenticated user to update their own display_name, email, avatar, and password.

#### Scenario: Update profile
- **WHEN** authenticated user sends PUT /api/v1/auth/profile with display_name and email
- **THEN** system updates the user's own profile

#### Scenario: Change password
- **WHEN** user sends PUT /api/v1/auth/profile with old_password and new_password
- **THEN** system verifies old password and updates to new bcrypt-hashed password

### Requirement: Default admin account on first startup
The system SHALL create a default admin account (username: admin, password: admin123) on first startup if no users exist.

#### Scenario: First startup initialization
- **WHEN** system starts and the users table is empty
- **THEN** system creates admin user with username="admin", password="admin123", role="admin"
