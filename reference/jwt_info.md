# Zen Bali JWT Authentication Guide

## Overview

Zen Bali uses **JWT (JSON Web Tokens)** for stateless authentication. The application implements a custom JWT-based authentication system using the `github.com/golang-jwt/jwt/v5` library.

## Current Status
✅ **JWT Authentication is FULLY IMPLEMENTED and ACTIVE**

## How JWT Works in Zen Bali

### Authentication Flow

```
1. User Registration/Login
   ├─> User submits credentials (email + password)
   ├─> Backend validates credentials against PostgreSQL
   ├─> Password verified using bcrypt
   └─> If valid, generate JWT token

2. JWT Token Generation
   ├─> Create Claims (user_id, email, user_type, expiry)
   ├─> Sign with HS256 algorithm using JWT_SECRET
   └─> Return signed token to client

3. Client Stores Token
   ├─> Store in localStorage/sessionStorage
   └─> Include in Authorization header for future requests

4. Protected Route Access
   ├─> Client sends: "Authorization: Bearer {token}"
   ├─> Middleware extracts and validates token
   ├─> Verify signature with JWT_SECRET
   ├─> Check expiration time
   ├─> Load user from database
   └─> Attach user to request context

5. Logout
   └─> Client deletes token (client-side only)
```

## JWT Configuration

### Environment Variables

Located in `.env`:

```bash
# JWT Configuration
JWT_SECRET=zenbali-dev-secret-key-change-in-production-min-32-chars
JWT_EXPIRY_HOURS=24
```

### Development vs Production JWT_SECRET

**For Development (Current Setup):**

✅ **Use the existing JWT_SECRET** - It's already configured and ready to use:
```bash
JWT_SECRET=zenbali-dev-secret-key-change-in-production-min-32-chars
```

**Why this is fine for development:**
- ✅ Long enough (32+ characters) - meets security requirements
- ✅ Works immediately - no additional setup needed
- ✅ Internal only - not exposed to external services
- ✅ Easy to test - can start developing right away

**No action needed for local development!** You can start testing authentication immediately.

---

**For Production (Before Deployment):**

⚠️ **MUST generate a new, cryptographically secure secret**

**How to Generate a Production JWT_SECRET:**

**Option 1: Using OpenSSL (Recommended)**
```bash
# Generate a 64-character base64 secret (strong)
openssl rand -base64 64

# Example output:
# XkJ8vN2pQw9rLmYtZa1bC3dE4fG5hI6jK7lM8nO9pQ0rS1tU2vW3xY4zA5bC6dE7f==
```

**Option 2: Using OpenSSL with Hex**
```bash
# Generate a 64-byte hex secret (even stronger)
openssl rand -hex 64

# Example output:
# 3a7f9c2e1b4d8f6a0c5e7b9d2f4a6c8e1b3d5f7a9c2e4b6d8f0a2c4e6b8d0f2a4c
```

**Option 3: Using Node.js**
```bash
# If you have Node.js installed
node -e "console.log(require('crypto').randomBytes(64).toString('base64'))"
```

**Option 4: Using Python**
```bash
# If you have Python installed
python3 -c "import secrets; print(secrets.token_urlsafe(64))"
```

**Option 5: Online Generator** (Use with caution)
- Only use on trusted machines
- Never use on shared/public computers
- Site: https://generate-secret.vercel.app/

---

**Updating Production JWT_SECRET:**

1. **Generate the secret** using one of the methods above

2. **Update your production `.env` file:**
   ```bash
   JWT_SECRET=XkJ8vN2pQw9rLmYtZa1bC3dE4fG5hI6jK7lM8nO9pQ0rS1tU2vW3xY4zA5bC6dE7f==
   JWT_EXPIRY_HOURS=2
   ```

3. **Use environment variable management:**
   - **Google Cloud Run**: Set as environment variable in Cloud Run config
   - **Google Secret Manager**: Store secret securely
   - **Kubernetes**: Use Kubernetes secrets
   - **Docker**: Pass via `-e` flag or `docker-compose.yml`

4. **Important Production Settings:**
   ```bash
   # Use shorter expiry in production (1-2 hours instead of 24)
   JWT_EXPIRY_HOURS=2

   # Never use the development secret
   JWT_SECRET=<your-generated-production-secret>
   ```

---

**Security Best Practices:**

- ⚠️ **Never commit JWT_SECRET to Git** - Always use environment variables
- ⚠️ **Use different secrets for dev/staging/production**
- ⚠️ **Rotate secrets periodically** (every 90 days recommended)
- ⚠️ **Keep secrets in secure storage** (Secret Manager, 1Password, etc.)
- ⚠️ **Minimum 32 characters** (64+ recommended for production)
- ⚠️ **Never share secrets** via email, Slack, or insecure channels
- ⚠️ **Invalidate old secrets** after rotation

**Important:**
- ⚠️ `JWT_SECRET` MUST be changed in production
- ⚠️ Minimum 32 characters for security (64+ recommended)
- ⚠️ Never commit production secrets to version control

### Configuration Structure

Defined in [backend/internal/config/config.go:33-36](backend/internal/config/config.go#L33-L36):

```go
type JWTConfig struct {
    Secret      string
    ExpiryHours int
}
```

Loaded from environment:
```go
JWT: JWTConfig{
    Secret:      getEnv("JWT_SECRET", "default-dev-secret-change-in-production-min-32-chars"),
    ExpiryHours: getEnvInt("JWT_EXPIRY_HOURS", 24),
}
```

## JWT Token Structure

### Claims Structure

Defined in [backend/internal/services/auth_service.go:34-39](backend/internal/services/auth_service.go#L34-L39):

```go
type Claims struct {
    UserID   uuid.UUID `json:"user_id"`     // Unique user identifier
    Email    string    `json:"email"`       // User's email
    UserType string    `json:"user_type"`   // "creator" or "admin"
    jwt.RegisteredClaims                     // Standard JWT claims
}
```

### Standard JWT Claims

Automatically included via `jwt.RegisteredClaims`:
- **ExpiresAt**: Token expiration timestamp
- **IssuedAt**: Token creation timestamp
- **Issuer**: Set to "zenbali"

### Example JWT Token

**Header:**
```json
{
  "alg": "HS256",
  "typ": "JWT"
}
```

**Payload:**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "creator@example.com",
  "user_type": "creator",
  "exp": 1704067200,
  "iat": 1703980800,
  "iss": "zenbali"
}
```

**Signature:**
```
HMACSHA256(
  base64UrlEncode(header) + "." + base64UrlEncode(payload),
  JWT_SECRET
)
```

## User Types

The system supports two user types:

### 1. Creator
- **Role**: Event organizers who create and manage events
- **UserType**: `"creator"`
- **Database Table**: `creators`
- **Access**: Creator dashboard, event management, payment management

### 2. Admin
- **Role**: Platform administrators
- **UserType**: `"admin"`
- **Database Table**: `admins`
- **Access**: Admin dashboard, all events, all creators, platform settings

## Authentication Endpoints

### Creator Registration

**Endpoint**: `POST /api/creator/register`

**Request:**
```json
{
  "name": "John Doe",
  "organization_name": "Bali Yoga Studio",
  "email": "john@example.com",
  "mobile": "+62812345678",
  "password": "securepassword123"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "John Doe",
    "email": "john@example.com",
    "is_verified": false,
    "is_active": true,
    "created_at": "2026-01-10T10:30:00Z"
  }
}
```

**Implementation**: [backend/internal/handlers/auth_handler.go:35-64](backend/internal/handlers/auth_handler.go#L35-L64)

### Creator Login

**Endpoint**: `POST /api/creator/login`

**Request:**
```json
{
  "email": "john@example.com",
  "password": "securepassword123"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "creator": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "John Doe",
      "email": "john@example.com"
    }
  }
}
```

**Implementation**: [backend/internal/handlers/auth_handler.go:66-96](backend/internal/handlers/auth_handler.go#L66-L96)

### Admin Login

**Endpoint**: `POST /api/admin/login`

**Request:**
```json
{
  "email": "admin@zenbali.org",
  "password": "admin123"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "admin": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "admin@zenbali.org",
      "name": "Admin User"
    }
  }
}
```

**Implementation**: [backend/internal/handlers/auth_handler.go:103-133](backend/internal/handlers/auth_handler.go#L103-L133)

### Logout

**Endpoint**: `POST /api/creator/logout`

**Response:**
```json
{
  "success": true,
  "message": "Logged out successfully"
}
```

**Note**: Since JWT is stateless, logout is handled client-side by deleting the token. The server doesn't maintain session state.

**Implementation**: [backend/internal/handlers/auth_handler.go:98-101](backend/internal/handlers/auth_handler.go#L98-L101)

## Using JWT Tokens

### Client-Side Usage

**Storing the Token (JavaScript):**
```javascript
// After successful login
const response = await fetch('/api/creator/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ email, password })
});

const data = await response.json();
const token = data.data.token;

// Store in localStorage
localStorage.setItem('auth_token', token);
```

**Sending Authenticated Requests:**
```javascript
// Get token from storage
const token = localStorage.getItem('auth_token');

// Include in Authorization header
const response = await fetch('/api/creator/events', {
  method: 'GET',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  }
});
```

**Logout:**
```javascript
// Remove token from storage
localStorage.removeItem('auth_token');

// Optionally call logout endpoint
await fetch('/api/creator/logout', {
  method: 'POST',
  headers: { 'Authorization': `Bearer ${token}` }
});
```

### Server-Side Token Validation

**Authorization Header Format:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Token Extraction**: [backend/internal/handlers/auth_handler.go:232-244](backend/internal/handlers/auth_handler.go#L232-L244)

```go
func extractToken(r *http.Request) string {
    auth := r.Header.Get("Authorization")
    if auth == "" {
        return ""
    }

    parts := strings.Split(auth, " ")
    if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
        return ""
    }

    return parts[1]
}
```

## Authentication Middleware

### Creator Authentication Middleware

**Implementation**: [backend/internal/handlers/auth_handler.go:135-169](backend/internal/handlers/auth_handler.go#L135-L169)

**Flow:**
1. Extract token from Authorization header
2. Validate token signature and expiration
3. Verify user_type is "creator"
4. Load creator from database
5. Check if account is active
6. Attach creator to request context
7. Proceed to handler

**Usage in Routes:**
```go
r.Group(func(r chi.Router) {
    r.Use(h.Auth.CreatorAuthMiddleware)

    r.Get("/creator/events", h.Creator.ListEvents)
    r.Post("/creator/events", h.Creator.CreateEvent)
    // ... more protected routes
})
```

### Admin Authentication Middleware

**Implementation**: [backend/internal/handlers/auth_handler.go:171-205](backend/internal/handlers/auth_handler.go#L171-L205)

**Similar to Creator middleware but:**
- Requires user_type "admin"
- Loads admin from database
- Attaches admin to context

**Usage in Routes:**
```go
r.Group(func(r chi.Router) {
    r.Use(h.Auth.AdminAuthMiddleware)

    r.Get("/admin/dashboard", h.Admin.Dashboard)
    r.Get("/admin/events", h.Admin.ListEvents)
    // ... more admin routes
})
```

## Accessing Authenticated User

### In Handlers

**Get Creator from Context:**
```go
func (h *CreatorHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
    creator := handlers.GetCreatorFromContext(r.Context())
    if creator == nil {
        utils.Unauthorized(w, "Not authenticated")
        return
    }

    // Use creator.ID, creator.Email, etc.
    utils.Success(w, creator.ToResponse())
}
```

**Get Admin from Context:**
```go
func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
    admin := handlers.GetAdminFromContext(r.Context())
    if admin == nil {
        utils.Unauthorized(w, "Not authenticated")
        return
    }

    // Use admin.ID, admin.Email, etc.
}
```

**Get User ID (Generic):**
```go
func SomeHandler(w http.ResponseWriter, r *http.Request) {
    userID := handlers.GetUserIDFromContext(r.Context())
    if userID == uuid.Nil {
        utils.Unauthorized(w, "Not authenticated")
        return
    }
}
```

**Helper Functions**: [backend/internal/handlers/auth_handler.go:207-229](backend/internal/handlers/auth_handler.go#L207-L229)

## Password Security

### Password Hashing

The application uses **bcrypt** for password hashing:

**Registration** (Hash password):
```go
hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
```

**Login** (Verify password):
```go
err := bcrypt.CompareHashAndPassword([]byte(creator.PasswordHash), []byte(password))
```

**Cost Factor**: `bcrypt.DefaultCost` (currently 10)
- Higher cost = more secure but slower
- 10 is a good balance for 2026

**Implementation**: [backend/internal/services/auth_service.go:52-55](backend/internal/services/auth_service.go#L52-L55) and [backend/internal/services/auth_service.go:86-88](backend/internal/services/auth_service.go#L86-L88)

## Token Generation Process

**Implementation**: [backend/internal/services/auth_service.go:147-161](backend/internal/services/auth_service.go#L147-L161)

```go
func (s *AuthService) generateToken(userID uuid.UUID, email, userType string) (string, error) {
    claims := &Claims{
        UserID:   userID,
        Email:    email,
        UserType: userType,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.config.ExpiryHours) * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "zenbali",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(s.config.Secret))
}
```

**Key Points:**
- Algorithm: HS256 (HMAC with SHA-256)
- Secret: From environment variable
- Expiry: Configurable (default 24 hours)
- Issuer: "zenbali"

## Token Validation Process

**Implementation**: [backend/internal/services/auth_service.go:123-137](backend/internal/services/auth_service.go#L123-L137)

```go
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(s.config.Secret), nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, errors.New("invalid token")
}
```

**Validation Checks:**
1. Token signature matches (using JWT_SECRET)
2. Token is not expired
3. Token format is valid
4. Claims can be parsed

## Error Handling

### Common Authentication Errors

**Invalid Credentials:**
```go
ErrInvalidCredentials = errors.New("invalid credentials")
// HTTP 401 Unauthorized
```

**Account Disabled:**
```go
ErrAccountDisabled = errors.New("account is disabled")
// HTTP 403 Forbidden
```

**Email Already Exists:**
```go
ErrEmailExists = errors.New("email already registered")
// HTTP 400 Bad Request
```

**Missing Token:**
```
"Missing authorization token"
// HTTP 401 Unauthorized
```

**Invalid/Expired Token:**
```
"Invalid or expired token"
// HTTP 401 Unauthorized
```

**Wrong User Type:**
```
"Access denied" (Creator accessing admin route)
"Admin access required" (Non-admin accessing admin route)
// HTTP 403 Forbidden
```

## Sessions Table

While JWT is stateless, the application maintains a `sessions` table for tracking:

**Table Structure:**
```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    user_type VARCHAR(20) NOT NULL,  -- 'creator' or 'admin'
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**Purpose:**
- Track active sessions
- Enable session revocation (future feature)
- Audit login activity
- Force logout if needed

**Note**: Currently, the sessions table is created but not actively used for validation. JWT validation is purely token-based.

## Security Best Practices

### Current Implementation ✅

- ✅ Passwords hashed with bcrypt (cost 10)
- ✅ JWT tokens signed with HS256
- ✅ Token expiration enforced
- ✅ User type validation in middleware
- ✅ Account active status check
- ✅ Password minimum length (8 characters)
- ✅ CORS configured properly

### Production Recommendations ⚠️

**1. Change JWT Secret**
```bash
# Generate strong secret (64+ characters)
JWT_SECRET=$(openssl rand -base64 64)
```

**2. Use Shorter Token Expiry**
```bash
# Instead of 24 hours, use 1-2 hours for production
JWT_EXPIRY_HOURS=2
```

**3. Implement Refresh Tokens**
- Short-lived access tokens (1-2 hours)
- Long-lived refresh tokens (7-30 days)
- Store refresh tokens in httpOnly cookies

**4. Add Rate Limiting**
- Limit login attempts per IP (e.g., 5 per 15 minutes)
- Protect against brute force attacks

**5. Enable HTTPS Only**
- Never send tokens over HTTP
- Use secure cookies for tokens

**6. Implement Token Revocation**
- Use sessions table to track valid tokens
- Invalidate tokens on logout
- Invalidate all tokens on password change

**7. Add CSRF Protection**
- Use CSRF tokens for state-changing operations
- Especially important if using cookies

**8. Monitor & Log**
- Log failed login attempts
- Alert on suspicious activity
- Track token usage patterns

## Testing Authentication

### Using cURL

**Creator Registration:**
```bash
curl -X POST http://localhost:8080/api/creator/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Creator",
    "email": "test@example.com",
    "password": "password123"
  }'
```

**Creator Login:**
```bash
curl -X POST http://localhost:8080/api/creator/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

**Access Protected Route:**
```bash
# Save token from login response
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

curl -X GET http://localhost:8080/api/creator/events \
  -H "Authorization: Bearer $TOKEN"
```

### Decoding JWT Token

**Online**: https://jwt.io/

**Command Line**:
```bash
# Extract payload (base64 decode)
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
echo $TOKEN | cut -d'.' -f2 | base64 -d | jq
```

## Troubleshooting

### Token Expired
**Error**: "Invalid or expired token"
**Solution**: Login again to get a new token

### Wrong User Type
**Error**: "Access denied"
**Solution**: Use the correct endpoint for your user type (creator vs admin)

### Missing Authorization Header
**Error**: "Missing authorization token"
**Solution**: Include `Authorization: Bearer {token}` header

### Invalid Token Format
**Error**: "Invalid or expired token"
**Solution**: Ensure format is `Bearer {token}`, not just `{token}`

### Account Disabled
**Error**: "Account is disabled"
**Solution**: Contact admin to reactivate account

## API Routes Summary

### Public (No Auth)
- `POST /api/creator/register` - Register new creator
- `POST /api/creator/login` - Creator login
- `POST /api/admin/login` - Admin login
- `GET /api/events` - List public events
- `GET /api/events/{id}` - Get single event

### Creator Protected (Requires Creator JWT)
- `GET /api/creator/profile` - Get profile
- `PUT /api/creator/profile` - Update profile
- `GET /api/creator/events` - List my events
- `POST /api/creator/events` - Create event
- `PUT /api/creator/events/{id}` - Update event
- `DELETE /api/creator/events/{id}` - Delete event
- `POST /api/creator/events/{id}/pay` - Create payment session
- `GET /api/creator/payments` - List my payments

### Admin Protected (Requires Admin JWT)
- `GET /api/admin/dashboard` - Dashboard stats
- `GET /api/admin/events` - List all events
- `PUT /api/admin/events/{id}` - Update any event
- `DELETE /api/admin/events/{id}` - Delete any event
- `GET /api/admin/creators` - List all creators
- `PUT /api/admin/creators/{id}` - Update creator
- `GET /api/admin/payments` - List all payments

## Additional Resources

- **JWT.io**: https://jwt.io/ - JWT debugger and documentation
- **golang-jwt/jwt**: https://github.com/golang-jwt/jwt - Go JWT library
- **bcrypt**: https://pkg.go.dev/golang.org/x/crypto/bcrypt - Password hashing
- **OWASP Auth Cheatsheet**: https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html
