# Web Interface Authentication Plan

## Overview
Add secret key-based authentication to the web interface to secure access to the service management dashboard.

## Requirements
- Environment variable configuration for secret key
- Login page with secret key input
- Session management for authenticated users
- Protection of all existing web routes
- Simple and secure authentication flow

## Implementation Plan

### 1. Backend Changes

#### 1.1 Environment Configuration
- Add support for `WEB_SECRET_KEY` environment variable
- If not set, generate a random key and log it on startup
- Store in service configuration structure

#### 1.2 Session Management
- Implement simple session storage (in-memory with UUID tokens)
- Add session middleware for protected routes
- Set session cookies with appropriate security flags

#### 1.3 Authentication Endpoints
- `POST /api/auth/login` - Accept secret key and create session
- `POST /api/auth/logout` - Destroy session
- `GET /api/auth/status` - Check authentication status

#### 1.4 Route Protection
- Protect existing API routes: `/api/*` (except auth endpoints)
- Protect static files and frontend routes
- Redirect unauthenticated users to login page

### 2. Frontend Changes

#### 2.1 Authentication Service
- Create `authService.ts` for login/logout operations
- Add authentication state management
- Implement auto-logout on session expiry

#### 2.2 Login Component
- Create `Login.vue` component with:
  - Secret key input field (password type)
  - Login button
  - Error message display
  - Simple, clean UI design

#### 2.3 Route Guards
- Add navigation guards to protect all routes
- Redirect to login if not authenticated
- Handle authentication state in router

#### 2.4 UI Updates
- Add logout button to main navigation
- Show authentication status
- Handle session expiry gracefully

### 3. Security Considerations

#### 3.1 Session Security
- Use secure, HttpOnly cookies
- Implement session timeout (default: 24 hours)
- Generate cryptographically secure session tokens

#### 3.2 Secret Key Security
- Log warning if default/weak secret key is used
- Support environment variable configuration
- Clear secret from memory after validation

#### 3.3 HTTPS Recommendation
- Add configuration notes about HTTPS in production
- Set appropriate cookie security flags

### 4. Configuration

#### 4.1 Environment Variables
```bash
WEB_SECRET_KEY=your-secure-secret-key-here
SESSION_TIMEOUT=24h  # Optional, defaults to 24 hours
```

#### 4.2 Default Behavior
- If `WEB_SECRET_KEY` not set, generate random key and log it
- Display warning about security in production
- Provide clear setup instructions

### 5. File Structure

```
auth/
├── middleware.go       # Authentication middleware
├── session.go         # Session management
└── handler.go          # Auth endpoints

web/frontend/src/
├── views/
│   └── Login.vue      # Login page component
├── services/
│   └── authService.ts # Authentication API calls
├── composables/
│   └── useAuth.ts     # Authentication state management
└── router/
    └── guards.ts      # Route protection
```

### 6. Implementation Steps

1. **Backend Authentication Core**
   - Create auth package with session management
   - Add authentication middleware
   - Implement login/logout endpoints

2. **Backend Route Protection**
   - Apply middleware to existing routes
   - Update server startup to include auth configuration
   - Add environment variable support

3. **Frontend Authentication Service**
   - Create authentication service and composables
   - Implement login/logout functionality
   - Add authentication state management

4. **Frontend Login UI**
   - Create login page component
   - Update router with authentication guards
   - Add logout functionality to main UI

5. **Testing & Documentation**
   - Add unit tests for authentication logic
   - Update documentation with setup instructions
   - Test authentication flow end-to-end

### 7. Testing Strategy

- Unit tests for authentication middleware
- Unit tests for session management
- Integration tests for auth endpoints
- E2E tests for login/logout flow
- Test route protection behavior

### 8. Documentation Updates

- Update CLAUDE.md with authentication setup
- Add environment variable documentation
- Include security best practices
- Provide troubleshooting guide

## Success Criteria

- ✅ Web interface requires authentication
- ✅ Secret key configurable via environment variable
- ✅ Simple login page with secret key input
- ✅ Session-based authentication
- ✅ All routes properly protected
- ✅ Graceful handling of session expiry
- ✅ Clear setup documentation
- ✅ Comprehensive test coverage
