# Zen Bali API Documentation

> Current verified local state as of 2026-03-23:
> - App: `http://localhost:8081`
> - API: `http://localhost:8081/api`
> - PostgreSQL host port: `5433`
> - Admin: `admin@zenbali.org` / `Teameditor@123`
> - Creator: `creator@zenbali.org` / `admin123`
> - Event posting fee: `$5 USD` (`500` cents)
> - Seed base state: 1 admin, 1 creator, 1 sample published event, 0 payments


**Version:** 1.0.0
**Last Updated:** 2026-03-27

## Base URL
- **Development:** `http://localhost:8081/api`
- **Production:** `https://zenbali.site/api`

## Authentication
Most creator and admin endpoints require a Bearer Token in the `Authorization` header.

```http
Authorization: Bearer <your_jwt_token>
```

The Zack machine endpoint uses either:

```http
Authorization: Bearer <agent_api_token>
```

or:

```http
X-Agent-Token: <agent_api_token>
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

## Agent Endpoints

These endpoints are intended for Zack/OpenClaw and use `AGENT_API_TOKEN`, not a creator/admin JWT.

### Create And Publish Event
`POST /api/agent/events`

Creates an event for the configured `AGENT_CREATOR_EMAIL` account, marks it `is_paid=true`, and publishes it immediately.

**Body:**
```json
{
  "title": "Ecstatic Dance Ubud",
  "event_date": "2026-04-10",
  "event_time": "19:15",
  "location": "Ubud",
  "event_type": "Dance",
  "duration_days": 0,
  "duration_hours": 2,
  "duration_minutes": 30,
  "entrance_type": "Paid",
  "participant_group_type": "Adults",
  "lead_by": "Maya",
  "venue": "Lotus Studio",
  "contact_email": "hello@example.com",
  "contact_mobile": "+628123456789",
  "event_description": "Sunset dance journey with live DJ.",
  "image_url": "https://example.com/event.jpg",
  "price_thousands": 150,
  "entrance_fee": 150000
}
```

**Notes:**
- `location`, `event_type`, and `entrance_type` can be sent as either the reference name or numeric ID.
- `event_time` must be in 15-minute increments.
- `duration_minutes` must be in 15-minute increments.
- `price_thousands` is the preferred price input. It accepts an integer from `0` to `100000`, where `1` means `IDR 1,000`.
- `entrance_fee` remains supported for backward compatibility and stores the full rupiah amount.
- `image_url` should be a publicly reachable image URL extracted or uploaded by Zack.
- If Zack has only the raw image file, upload it first with `POST /api/agent/uploads/event-image` and use the returned `image_url`.

**Field Mapping For Zack Extraction:**
- `e9` -> `title`
- `e10` -> `event_date`
- `e11` -> `event_time`
- `e109` -> `location`
- `e136` -> `event_type`
- `e163` -> `duration_days`
- `e166` -> `duration_hours`
- `e191` -> `duration_minutes`
- `e196` -> `entrance_type`
- `e253` -> `participant_group_type`
- `e259` -> `lead_by`
- `e260` -> `venue`
- `e261` -> `contact_email`
- `e262` -> `contact_mobile`
- `e263` -> `event_description`
- `e267` -> `image_url`

---

### Upload Event Image
`POST /api/agent/uploads/event-image`

Uploads a raw image file for Zack and returns a public `image_url` suitable for `e267` / `image_url` in the event payload.

**Request:**
- Content type: `multipart/form-data`
- Form field: `image`

**Response:**
```json
{
  "success": true,
  "data": {
    "image_url": "https://storage.googleapis.com/your-bucket/zenbali/abc123.jpg"
  }
}
```

**Notes:**
- Allowed file types: `jpg`, `jpeg`, `png`, `webp`
- Max size follows `MAX_UPLOAD_SIZE_MB`
- This endpoint uses the same agent token as `POST /api/agent/events`

---

## Webhooks

### Stripe
`POST /api/webhooks/stripe`

This endpoint listens for events from Stripe, such as `checkout.session.completed`, to mark events as paid. It is secured by verifying the Stripe signature.
