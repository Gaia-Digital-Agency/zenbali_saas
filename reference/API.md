# Zen Bali API Documentation

**Version:** 1.0.0
**Last Updated:** 2026-01-10

## Base URL
- **Development:** `http://localhost:8080/api`
- **Production:** `https://zenbali.org/api`

## Authentication
Most creator and admin endpoints require a Bearer Token in the `Authorization` header.

```http
Authorization: Bearer <your_jwt_token>
```

---

## Public Endpoints

These endpoints are open and do not require authentication.

### Health Check
Checks the status of the API.

`GET /api/health`

**Response: 200 OK**
```json
{
  "status": "ok",
  "message": "Zen Bali API is running"
}
```

### List Events
Returns a paginated list of published, upcoming events.

`GET /api/events`

**Query Parameters:**
- `page` (int, optional): The page number to retrieve.
- `limit` (int, optional): The number of events per page.
- `location_id` (int, optional): Filter by location ID.
- `event_type_id` (int, optional): Filter by event type ID.
- `date_from` (string, optional): Filter events starting from this date (YYYY-MM-DD).
- `search` (string, optional): Search term to filter events by title or description.

### Get Event Details
Retrieves a single event by its ID.

`GET /api/events/{id}`

### List Reference Data
`GET /api/locations` - Returns a list of all locations.
`GET /api/event-types` - Returns a list of all event types.
`GET /api/entrance-types` - Returns a list of all entrance types.

### Visitor Tracking
`POST /api/visitors` - Tracks a new visitor.
`GET /api/visitors/stats` - Retrieves visitor statistics.

---

## Creator Endpoints

### Authentication

#### Register
`POST /api/creator/register`

**Body:**
```json
{
  "name": "Test Creator",
  "organization_name": "Creative Inc.",
  "email": "creator@example.com",
  "password": "securepassword123",
  "mobile": "+628123456789"
}
```

#### Login
`POST /api/creator/login`

**Body:**
```json
{
  "email": "creator@example.com",
  "password": "securepassword123"
}
```
**Response:**
```json
{
  "token": "your_jwt_token",
  "creator": { ... }
}
```

#### Logout
`POST /api/creator/logout` (Requires Auth)

### Profile Management (Auth Required)

`GET /api/creator/profile` - Get creator's profile.
`PUT /api/creator/profile` - Update creator's profile.

### Event Management (Auth Required)

`GET /api/creator/events` - List events for the authenticated creator.
`POST /api/creator/events` - Create a new event.
`GET /api/creator/events/{id}` - Get details of a specific event.
`PUT /api/creator/events/{id}` - Update an event.
`DELETE /api/creator/events/{id}` - Delete an event.
`POST /api/creator/events/{id}/upload-image` - Upload an image for an event.

### Payments (Auth Required)

`POST /api/creator/events/{id}/pay` - Creates a Stripe Checkout session to pay for an event posting.
`GET /api/creator/payments` - List payment history for the creator.

---

## Admin Endpoints (Admin Auth Required)

### Authentication

`POST /api/admin/login`

**Body:**
```json
{
  "email": "admin@zenbali.org",
  "password": "adminpassword"
}
```

### Dashboard
`GET /api/admin/dashboard` - Get statistics for the admin dashboard.

### Resource Management
`GET /api/admin/events` - List all events.
`PUT /api/admin/events/{id}` - Update an event (e.g., approve, hide).
`DELETE /api/admin/events/{id}` - Delete an event.

`GET /api/admin/creators` - List all creators.
`PUT /api/admin/creators/{id}` - Update a creator (e.g., activate/deactivate).

`GET /api/admin/payments` - List all payments.
`GET /api/admin/payments/export` - Export payments to CSV.

### Settings Management
`GET /api/admin/settings/locations`
`POST /api/admin/settings/locations`
`PUT /api/admin/settings/locations/{id}`

`GET /api/admin/settings/event-types`
`POST /api/admin/settings/event-types`
`PUT /api/admin/settings/event-types/{id}`

---

## Webhooks

### Stripe
`POST /api/webhooks/stripe`

This endpoint listens for events from Stripe, such as `checkout.session.completed`, to mark events as paid. It is secured by verifying the Stripe signature.