# Zen Bali User Acceptance Testing (UAT) Guide

> Current verified local state as of 2026-03-23:
> - App: `http://localhost:8081`
> - API: `http://localhost:8081/api`
> - PostgreSQL host port: `5433`
> - Admin: `admin@zenbali.org` / `Teameditor@123`
> - Creator: `creator@zenbali.org` / `admin123`
> - Event posting fee: `$5 USD` (`500` cents)
> - Seed base state: 1 admin, 1 creator, 1 sample published event, 0 payments


## Overview

This guide provides step-by-step instructions for testing the Zen Bali platform as three different user types:
1. **Visitor** - Public users browsing events
2. **Creator** - Event organizers who post events
3. **Admin** - Platform administrators

## Prerequisites

Before starting UAT:

✅ **Server Running**
```bash
make dev
# Server should be running on http://localhost:8081
```

✅ **Docker Services Active**
- PostgreSQL: Running on port 5433
- Redis: Running on port 6379

✅ **Database Seeded**
- Reference data loaded (locations, event types, entrance types)

✅ **Test Credentials Ready**
- Admin: `admin@zenbali.org` / `Teameditor@123`
- Creator: `creator@zenbali.org` / `admin123`

## Test Environment URLs

| Page | URL |
|------|-----|
| **Home Page** | http://localhost:8081 |
| **Event Details** | http://localhost:8081/event.html?id={event_id} |
| **Creator Login** | http://localhost:8081/creator/login.html |
| **Creator Register** | http://localhost:8081/creator/register.html |
| **Creator Dashboard** | http://localhost:8081/creator/dashboard.html |
| **Admin Login** | http://localhost:8081/admin/login.html |
| **Admin Dashboard** | http://localhost:8081/admin/dashboard.html |

---

# Part 1: Visitor Testing (Public User)

**Role**: Anonymous user browsing events
**No login required**

## Test Case 1.1: Home Page Access

**Objective**: Verify home page loads and displays correctly

**Steps**:
1. Open browser and navigate to http://localhost:8081
2. Observe the page layout

**Expected Results**:
- ✅ Page loads successfully with CSS styles
- ✅ Header shows "🌴 Zen Bali" logo
- ✅ Navigation menu visible
- ✅ Hero section displays welcome message
- ✅ Search/filter section visible
- ✅ Footer displays visitor statistics

**Screenshot Location**: `screenshots/visitor/home-page.png`

---

## Test Case 1.2: View Event Listings

**Objective**: View all published events

**Steps**:
1. On home page, scroll to events section
2. Observe event cards displayed

**Expected Results**:
- ✅ Events displayed in card format
- ✅ Each card shows:
  - Event title
  - Date and time
  - Location
  - Event type
  - Entrance type
  - "View Details" button
- ✅ Only published events are visible
- ✅ Events sorted by date (upcoming first)

**Test Data**: Initially, there may be no events if none have been created yet

---

## Test Case 1.3: Filter Events by Location

**Objective**: Filter events by Bali location

**Steps**:
1. On home page, locate the location filter dropdown
2. Select a location (e.g., "Ubud")
3. Observe filtered results

**Expected Results**:
- ✅ Dropdown shows all locations (25 locations)
- ✅ Events filter to show only selected location
- ✅ Event count updates
- ✅ "Clear filters" option available

**Test Locations**: Ubud, Canggu, Seminyak, Sanur

---

## Test Case 1.4: Filter Events by Type

**Objective**: Filter events by event type

**Steps**:
1. Select an event type (e.g., "Yoga")
2. Observe filtered results

**Expected Results**:
- ✅ Dropdown shows all event types (25 types)
- ✅ Events filter to show only selected type
- ✅ Can combine with location filter
- ✅ Results update dynamically

**Test Event Types**: Yoga, Healing, Meditation, Retreat

---

## Test Case 1.5: Search Events

**Objective**: Search events by keyword

**Steps**:
1. Locate search box
2. Enter search term (e.g., "yoga")
3. Observe search results

**Expected Results**:
- ✅ Search box accepts text input
- ✅ Results filter by title/description
- ✅ Search is case-insensitive
- ✅ Results update as you type (if implemented)

---

## Test Case 1.6: View Event Details

**Objective**: View detailed information about a specific event

**Steps**:
1. Click "View Details" on any event card
2. Observe event details page

**Expected Results**:
- ✅ Navigates to event details page
- ✅ URL contains event ID: `event.html?id={uuid}`
- ✅ Displays complete event information:
  - Title
  - Full description
  - Date and time
  - Duration
  - Location
  - Event type
  - Entrance type and fee
  - Contact information (email, mobile)
  - Event image (if uploaded)
- ✅ Back/Home navigation available

---

## Test Case 1.7: Visitor Tracking

**Objective**: Verify visitor statistics are tracked

**Steps**:
1. Visit home page
2. Check footer for visitor stats
3. Refresh page
4. Check if count increases

**Expected Results**:
- ✅ Footer shows visitor statistics
- ✅ Today's visitor count displayed
- ✅ Total visitors displayed
- ✅ Unique visitors tracked (based on IP)

**API Endpoint**: `GET /api/visitors/stats`

---

# Part 2: Creator Testing (Event Organizer)

**Role**: Event organizer who creates and manages events
**Requires registration and login**

## Test Case 2.1: Creator Registration

**Objective**: Register a new creator account

**Steps**:
1. Navigate to http://localhost:8081/creator/register.html
2. Fill in registration form:
   - **Name**: "Test Creator"
   - **Organization**: "Bali Yoga Studio"
   - **Email**: "testcreator@example.com"
   - **Mobile**: "+62812345678"
   - **Password**: "password123"
3. Click "Register" button

**Expected Results**:
- ✅ Form validates all required fields
- ✅ Email format validated
- ✅ Password minimum 8 characters enforced
- ✅ Registration successful message
- ✅ Account created in database
- ✅ Redirects to login page

**Validation Tests**:
- ❌ Empty fields show error
- ❌ Invalid email format rejected
- ❌ Password < 8 characters rejected
- ❌ Duplicate email shows "Email already registered"

**Database Check**:
```sql
SELECT * FROM creators WHERE email = 'testcreator@example.com';
```

---

## Test Case 2.2: Creator Login

**Objective**: Login with creator credentials

**Steps**:
1. Navigate to http://localhost:8081/creator/login.html
2. Enter credentials:
   - **Email**: "testcreator@example.com"
   - **Password**: "password123"
3. Click "Login" button

**Expected Results**:
- ✅ Login successful
- ✅ JWT token generated and stored
- ✅ Redirects to creator dashboard
- ✅ Dashboard shows creator name

**Validation Tests**:
- ❌ Wrong password shows "Invalid email or password"
- ❌ Non-existent email shows error
- ❌ Empty fields show validation error

**Browser DevTools Check**:
```javascript
// Check localStorage for token
localStorage.getItem('auth_token')
// Should return JWT token
```

---

## Test Case 2.3: View Creator Dashboard

**Objective**: Access and view creator dashboard

**Steps**:
1. After login, observe dashboard page
2. Check displayed information

**Expected Results**:
- ✅ Creator name displayed in header
- ✅ Navigation menu shows:
  - Dashboard
  - My Events
  - Create Event
  - Payments
  - Profile
  - Logout
- ✅ Dashboard shows statistics:
  - Total events created
  - Published events
  - Unpaid events
  - Total revenue
- ✅ Recent events list displayed

---

## Test Case 2.4: Create New Event (Unpaid)

**Objective**: Create a new event listing

**Steps**:
1. Click "Create Event" in navigation
2. Fill in event form:
   - **Title**: "Morning Yoga Session"
   - **Event Date**: "2026-02-15"
   - **Event Time**: "08:00"
   - **Location**: Select "Ubud"
   - **Event Type**: Select "Yoga"
   - **Duration**: "90 minutes"
   - **Entrance Type**: Select "Pay at Site"
   - **Entrance Fee**: "150000" (IDR)
   - **Contact Email**: "testcreator@example.com"
   - **Contact Mobile**: "+62812345678"
   - **Notes**: "Bring your own yoga mat"
3. Click "Create Event" button

**Expected Results**:
- ✅ Form validates all required fields
- ✅ Event created successfully
- ✅ Event status: **Unpublished** (is_published = false)
- ✅ Event status: **Unpaid** (is_paid = false)
- ✅ Redirects to event details or events list
- ✅ Success message displayed
- ✅ Event appears in "My Events" list

**Database Check**:
```sql
SELECT title, is_paid, is_published FROM events
WHERE title = 'Morning Yoga Session';
-- Should show: is_paid = false, is_published = false
```

---

## Test Case 2.5: View My Events

**Objective**: View list of creator's events

**Steps**:
1. Click "My Events" in navigation
2. Observe events list

**Expected Results**:
- ✅ Shows all events created by this creator
- ✅ Each event displays:
  - Title
  - Date
  - Location
  - Event type
  - Payment status (Paid/Unpaid)
  - Published status (Yes/No)
- ✅ Unpaid events marked clearly
- ✅ "Pay to Publish" button for unpaid events
- ✅ "Edit" and "Delete" options available

---

## Test Case 2.6: Edit Event

**Objective**: Modify event details

**Steps**:
1. From "My Events", click "Edit" on an event
2. Modify event details:
   - Change title to "Morning Yoga Session (Updated)"
   - Change time to "09:00"
3. Click "Update Event"

**Expected Results**:
- ✅ Form pre-fills with current event data
- ✅ All fields editable
- ✅ Update successful message
- ✅ Changes saved to database
- ✅ Updated event displayed

**Note**: Cannot edit already paid events (optional feature)

---

## Test Case 2.7: Upload Event Image

**Objective**: Upload an image for the event

**Steps**:
1. From event details or edit page
2. Click "Upload Image" button
3. Select an image file (JPG/PNG, < 5MB)
4. Upload image

**Expected Results**:
- ✅ File picker opens
- ✅ Only allows image formats (.jpg, .jpeg, .png, .webp)
- ✅ Validates file size (max 5MB)
- ✅ Upload progress shown
- ✅ Image uploaded successfully
- ✅ Image displayed on event
- ✅ Image saved to the configured upload backend
- ✅ Image URL stored in database

**Validation Tests**:
- ❌ File > 5MB rejected
- ❌ Non-image files rejected

---

## Test Case 2.8: Create Payment Session (Stripe Checkout)

**Objective**: Initiate payment to publish event

**Steps**:
1. From "My Events", click "Pay to Publish" on unpaid event
2. Observe Stripe Checkout redirect

**Expected Results**:
- ✅ Redirects to Stripe Checkout page
- ✅ Checkout session created in database
- ✅ Payment amount shows $10.00 USD
- ✅ Product name: "Event Posting Fee - {Event Title}"
- ✅ Session metadata includes event_id and creator_id

**Database Check**:
```sql
SELECT * FROM payments WHERE event_id = '{event_uuid}';
-- Should show: status = 'pending'
```

**API Request**:
```bash
POST /api/creator/events/{event_id}/pay
Authorization: Bearer {creator_token}
```

---

## Test Case 2.9: Complete Payment (Stripe Test Mode)

**Objective**: Complete payment using Stripe test card

**Steps**:
1. On Stripe Checkout page, enter test card details:
   - **Card Number**: 4242 4242 4242 4242
   - **Expiry**: 12/34 (any future date)
   - **CVC**: 123
   - **Zip**: 12345
2. Click "Pay"
3. Wait for redirect

**Expected Results**:
- ✅ Payment processes successfully
- ✅ Redirects to success URL
- ✅ Webhook received by backend
- ✅ Payment status updated to "completed"
- ✅ Event marked as paid (is_paid = true)
- ✅ Event published (is_published = true)
- ✅ Event now visible on public site

**Database Check**:
```sql
SELECT is_paid, is_published FROM events WHERE id = '{event_uuid}';
-- Should show: is_paid = true, is_published = true

SELECT status, stripe_payment_intent_id FROM payments
WHERE event_id = '{event_uuid}';
-- Should show: status = 'completed'
```

**Webhook Logs**:
Check server logs for:
```
Successfully processed payment for session: cs_test_xxxxx
```

---

## Test Case 2.10: Cancel Payment

**Objective**: Cancel payment during checkout

**Steps**:
1. Initiate payment
2. On Stripe Checkout, click "Back" or close window
3. Observe result

**Expected Results**:
- ✅ Returns to cancel URL
- ✅ Payment status remains "pending"
- ✅ Event remains unpublished
- ✅ Can retry payment later

---

## Test Case 2.11: View Payment History

**Objective**: View all payments made by creator

**Steps**:
1. Click "Payments" in navigation
2. Observe payment list

**Expected Results**:
- ✅ Shows all payments for this creator
- ✅ Each payment displays:
  - Event title
  - Amount ($10.00)
  - Status (Pending/Completed/Failed)
  - Date
  - Stripe session ID
- ✅ Completed payments show payment intent ID
- ✅ Can filter by status

---

## Test Case 2.12: Delete Unpaid Event

**Objective**: Delete an unpaid event

**Steps**:
1. From "My Events", click "Delete" on an unpaid event
2. Confirm deletion

**Expected Results**:
- ✅ Confirmation dialog appears
- ✅ Event deleted from database
- ✅ Removed from "My Events" list
- ✅ Success message shown

**Note**: May prevent deleting paid events (business rule)

---

## Test Case 2.13: Update Creator Profile

**Objective**: Update creator account information

**Steps**:
1. Click "Profile" in navigation
2. Update fields:
   - **Name**: "Test Creator Updated"
   - **Organization**: "Bali Wellness Center"
   - **Mobile**: "+62898765432"
3. Click "Update Profile"

**Expected Results**:
- ✅ Form pre-fills with current data
- ✅ Can update all fields except email
- ✅ Changes saved successfully
- ✅ Updated info displayed

---

## Test Case 2.14: Creator Logout

**Objective**: Logout from creator account

**Steps**:
1. Click "Logout" in navigation
2. Observe result

**Expected Results**:
- ✅ JWT token removed from localStorage
- ✅ Redirects to login page
- ✅ Cannot access protected pages without login
- ✅ Attempting to access dashboard shows "Unauthorized"

---

# Part 3: Admin Testing (Platform Administrator)

**Role**: Platform administrator with full access
**Login required**

## Test Case 3.1: Admin Login

**Objective**: Login as admin

**Steps**:
1. Navigate to http://localhost:8081/admin/login.html
2. Enter credentials:
   - **Email**: "admin@zenbali.org"
   - **Password**: "Teameditor@123"
3. Click "Login"

**Expected Results**:
- ✅ Login successful
- ✅ JWT token generated (user_type = "admin")
- ✅ Redirects to admin dashboard
- ✅ Admin interface displayed

**Validation Tests**:
- ❌ Creator credentials cannot access admin
- ❌ Wrong password shows error

---

## Test Case 3.2: View Admin Dashboard

**Objective**: Access admin dashboard and view statistics

**Steps**:
1. After login, observe dashboard
2. Check displayed statistics

**Expected Results**:
- ✅ Dashboard shows platform statistics:
  - Total events
  - Published events
  - Total creators
  - Active creators
  - Total payments
  - Total revenue
  - Today's visitors
  - Total visitors
- ✅ Navigation menu shows:
  - Dashboard
  - Events
  - Creators
  - Payments
  - Settings
  - Logout
- ✅ Recent activity displayed

**API Endpoint**: `GET /api/admin/dashboard`

---

## Test Case 3.3: View All Events

**Objective**: View all events (published and unpublished)

**Steps**:
1. Click "Events" in navigation
2. Observe events list

**Expected Results**:
- ✅ Shows ALL events from ALL creators
- ✅ Can filter by:
  - Published/Unpublished
  - Paid/Unpaid
  - Location
  - Event type
  - Date range
- ✅ Search by title
- ✅ Pagination available
- ✅ Each event shows:
  - Title
  - Creator name
  - Date
  - Location
  - Payment status
  - Published status
- ✅ Edit and Delete options

**API Endpoint**: `GET /api/admin/events`

---

## Test Case 3.4: Edit Any Event (Admin Override)

**Objective**: Admin can edit any creator's event

**Steps**:
1. From events list, click "Edit" on any event
2. Modify event details
3. Save changes

**Expected Results**:
- ✅ Can edit events from any creator
- ✅ Can change published status
- ✅ Can modify all event fields
- ✅ Changes saved successfully
- ✅ Audit log created (if implemented)

---

## Test Case 3.5: Delete Any Event

**Objective**: Admin can delete any event

**Steps**:
1. From events list, click "Delete" on an event
2. Confirm deletion

**Expected Results**:
- ✅ Can delete events from any creator
- ✅ Can delete paid events (admin privilege)
- ✅ Confirmation required
- ✅ Event deleted from database
- ✅ Associated payment records handled

---

## Test Case 3.6: View All Creators

**Objective**: View and manage creator accounts

**Steps**:
1. Click "Creators" in navigation
2. Observe creators list

**Expected Results**:
- ✅ Shows all registered creators
- ✅ Each creator displays:
  - Name
  - Organization
  - Email
  - Mobile
  - Verified status
  - Active status
  - Event count
  - Registration date
- ✅ Search by name/email
- ✅ Filter by verified/active status
- ✅ Pagination available

**API Endpoint**: `GET /api/admin/creators`

---

## Test Case 3.7: Update Creator Status

**Objective**: Activate/deactivate creator accounts

**Steps**:
1. From creators list, click "Edit" on a creator
2. Toggle "Active" status to disabled
3. Save changes

**Expected Results**:
- ✅ Can enable/disable creator accounts
- ✅ Disabled creators cannot login
- ✅ Disabled creators' events hidden (optional)
- ✅ Changes saved successfully

---

## Test Case 3.8: Verify Creator

**Objective**: Mark creator as verified

**Steps**:
1. From creator details, toggle "Verified" status
2. Save changes

**Expected Results**:
- ✅ Can verify/unverify creators
- ✅ Verified badge shown on creator events (if implemented)
- ✅ Status saved successfully

---

## Test Case 3.9: View All Payments

**Objective**: View all payment transactions

**Steps**:
1. Click "Payments" in navigation
2. Observe payments list

**Expected Results**:
- ✅ Shows ALL payments from ALL creators
- ✅ Each payment displays:
  - Event title
  - Creator name
  - Amount
  - Currency
  - Status
  - Stripe session ID
  - Payment intent ID
  - Created date
- ✅ Filter by:
  - Status (pending/completed/failed)
  - Date range
  - Creator
- ✅ Search by event or creator
- ✅ Total revenue displayed

**API Endpoint**: `GET /api/admin/payments`

---

## Test Case 3.10: Export Payment Data

**Objective**: Export payments to CSV

**Steps**:
1. On payments page, click "Export" button
2. Select date range (optional)
3. Download file

**Expected Results**:
- ✅ CSV file downloads
- ✅ Contains all payment records
- ✅ Includes all relevant fields
- ✅ Proper CSV formatting
- ✅ Opens in Excel/Google Sheets

**API Endpoint**: `GET /api/admin/payments/export`

---

## Test Case 3.11: Manage Locations

**Objective**: Add, edit, or deactivate locations

**Steps**:
1. Navigate to Settings > Locations
2. Click "Add Location"
3. Enter:
   - **Name**: "Pererenan"
   - **Slug**: "pererenan"
4. Save

**Expected Results**:
- ✅ Can add new locations
- ✅ Slug auto-generated from name
- ✅ Can edit existing locations
- ✅ Can activate/deactivate locations
- ✅ Inactive locations hidden in creator dropdowns
- ✅ Changes reflected immediately

---

## Test Case 3.12: Manage Event Types

**Objective**: Add, edit, or deactivate event types

**Steps**:
1. Navigate to Settings > Event Types
2. Click "Add Event Type"
3. Enter:
   - **Name**: "Art Workshop"
   - **Slug**: "art-workshop"
4. Save

**Expected Results**:
- ✅ Can add new event types
- ✅ Can edit existing types
- ✅ Can activate/deactivate types
- ✅ Changes reflected in creator forms

---

## Test Case 3.13: Admin Logout

**Objective**: Logout from admin account

**Steps**:
1. Click "Logout"
2. Observe result

**Expected Results**:
- ✅ Token removed
- ✅ Redirects to login
- ✅ Cannot access admin pages

---

# Part 4: Integration Testing

## Test Case 4.1: End-to-End Event Publishing Flow

**Objective**: Complete flow from creation to public display

**Steps**:
1. **Creator**: Register and login
2. **Creator**: Create event
3. **Creator**: Upload image
4. **Creator**: Initiate payment
5. **Stripe**: Complete payment with test card
6. **System**: Webhook processes payment
7. **Visitor**: View event on public site

**Expected Results**:
- ✅ Event progresses through all states
- ✅ Payment webhook updates event status
- ✅ Event appears on public site after payment
- ✅ All event details display correctly

---

## Test Case 4.2: Multi-Creator Scenario

**Objective**: Multiple creators managing events simultaneously

**Steps**:
1. Register 3 different creators
2. Each creator creates 2 events
3. Each creator pays for 1 event
4. Admin views all events

**Expected Results**:
- ✅ Each creator sees only their events
- ✅ Admin sees all 6 events
- ✅ Public sees 3 paid events
- ✅ No data leakage between creators

---

## Test Case 4.3: Payment Failure Handling

**Objective**: Test payment failure scenarios

**Steps**:
1. Create event
2. Initiate payment
3. Use Stripe test card for declined payment: 4000 0000 0000 0002
4. Observe result

**Expected Results**:
- ✅ Payment declined by Stripe
- ✅ Payment status remains "pending"
- ✅ Event remains unpublished
- ✅ Creator can retry payment
- ✅ Error message displayed

---

## Test Case 4.4: Visitor Statistics Accuracy

**Objective**: Verify visitor tracking works correctly

**Steps**:
1. Visit site from browser 1
2. Visit site from browser 2 (incognito)
3. Refresh browser 1 multiple times
4. Check visitor stats

**Expected Results**:
- ✅ Two unique visitors counted
- ✅ Total visits = multiple refreshes
- ✅ Stats update in real-time
- ✅ IP-based deduplication works

---

# Part 5: API Testing

## Using cURL for API Testing

### Public Endpoints

**Get All Published Events:**
```bash
curl http://localhost:8081/api/events
```

**Get Single Event:**
```bash
curl http://localhost:8081/api/events/{event_id}
```

**Get Locations:**
```bash
curl http://localhost:8081/api/locations
```

**Get Event Types:**
```bash
curl http://localhost:8081/api/event-types
```

### Creator Endpoints

**Register:**
```bash
curl -X POST http://localhost:8081/api/creator/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "API Test Creator",
    "email": "apitest@example.com",
    "password": "password123"
  }'
```

**Login:**
```bash
curl -X POST http://localhost:8081/api/creator/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "apitest@example.com",
    "password": "password123"
  }'
```

**Create Event (with token):**
```bash
TOKEN="your_jwt_token_here"

curl -X POST http://localhost:8081/api/creator/events \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "API Created Event",
    "event_date": "2026-03-01",
    "location_id": 1,
    "event_type_id": 1,
    "entrance_type_id": 1,
    "contact_email": "apitest@example.com"
  }'
```

**Get My Events:**
```bash
curl http://localhost:8081/api/creator/events \
  -H "Authorization: Bearer $TOKEN"
```

### Admin Endpoints

**Login:**
```bash
curl -X POST http://localhost:8081/api/admin/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@zenbali.org",
    "password": "Teameditor@123"
  }'
```

**Dashboard Stats:**
```bash
ADMIN_TOKEN="admin_jwt_token_here"

curl http://localhost:8081/api/admin/dashboard \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

---

# Part 6: Browser Testing Matrix

## Supported Browsers

Test on the following browsers:

| Browser | Version | Desktop | Mobile |
|---------|---------|---------|--------|
| Chrome | Latest | ✅ | ✅ |
| Firefox | Latest | ✅ | ✅ |
| Safari | Latest | ✅ | ✅ |
| Edge | Latest | ✅ | ❌ |

## Responsive Design Testing

Test on these viewport sizes:

- **Desktop**: 1920x1080, 1366x768
- **Tablet**: 768x1024 (iPad)
- **Mobile**: 375x667 (iPhone), 360x640 (Android)

---

# Part 7: Performance Testing

## Page Load Times

Measure and record:

- Home page: < 2 seconds
- Event details: < 1 second
- Dashboard: < 1.5 seconds

## API Response Times

- GET /api/events: < 200ms
- POST /api/creator/events: < 300ms
- POST /api/creator/login: < 400ms

---

# Part 8: Security Testing

## Authentication Tests

**Test Case 8.1: Access Protected Routes Without Token**
```bash
curl http://localhost:8081/api/creator/events
# Should return: 401 Unauthorized
```

**Test Case 8.2: Use Expired Token**
- Wait 24+ hours or manually create expired token
- Attempt to access protected route
- Should return: 401 Unauthorized

**Test Case 8.3: Cross-User Access**
- Creator A tries to access Creator B's events
- Should only see own events

**Test Case 8.4: SQL Injection Protection**
- Try: `'; DROP TABLE events; --` in search
- Should sanitize input safely

**Test Case 8.5: XSS Protection**
- Create event with title: `<script>alert('XSS')</script>`
- Should escape HTML properly

---

# Test Results Template

## UAT Sign-off Sheet

| Test Case | Status | Tested By | Date | Notes |
|-----------|--------|-----------|------|-------|
| 1.1 Home Page | ⬜ Pass ⬜ Fail | | | |
| 1.2 View Events | ⬜ Pass ⬜ Fail | | | |
| 2.1 Creator Register | ⬜ Pass ⬜ Fail | | | |
| 2.9 Complete Payment | ⬜ Pass ⬜ Fail | | | |
| 3.1 Admin Login | ⬜ Pass ⬜ Fail | | | |

---

# Bug Reporting Template

```markdown
## Bug Report

**Bug ID**: BUG-001
**Severity**: High / Medium / Low
**Test Case**: 2.9
**Environment**: Development (localhost:8081)

**Description**:
Payment webhook not processing

**Steps to Reproduce**:
1. Create event
2. Initiate payment
3. Complete payment with test card

**Expected Result**:
Event should be published

**Actual Result**:
Event remains unpublished

**Screenshots**:
[Attach screenshots]

**Console Errors**:
[Paste console logs]

**Tested By**: [Name]
**Date**: 2026-01-10
```

---

# Appendix

## Test Data Reference

### Test Stripe Cards

| Card Number | Scenario |
|-------------|----------|
| 4242 4242 4242 4242 | Successful payment |
| 4000 0025 0000 3155 | Requires 3D Secure |
| 4000 0000 0000 9995 | Insufficient funds |
| 4000 0000 0000 0002 | Card declined |

### Default Admin Credentials
- **Email**: admin@zenbali.org
- **Password**: Teameditor@123

### Test Creator Credentials
- **Email**: creator@zenbali.org
- **Password**: admin123
- **Name**: Test Creator
- **Organization**: Test Organization

### Pre-seeded Sample Event
A sample event is automatically created by the seed data:
- **Title**: Sample Yoga Session in Ubud
- **Date**: Tomorrow's date (dynamically set)
- **Location**: Ubud
- **Type**: Yoga
- **Status**: Paid and Published (visible on landing page)
- **Creator**: creator@zenbali.org

### Reference Data
- **Locations**: Ubud, Canggu, Seminyak, Sanur, etc. (25 total)
- **Event Types**: Yoga, Healing, Meditation, Retreat, etc. (25 total)
- **Entrance Types**: Free, Prepaid Online, Pay at Site, etc. (6 total)

---

## Troubleshooting

**Issue**: Cannot login
- Check JWT_SECRET is configured
- Clear browser cache and localStorage
- Check server logs for errors

**Issue**: Payment webhook not received
- Ensure Stripe CLI is running
- Check webhook secret in .env
- Verify server is running on port 8081

**Issue**: Events not displaying
- Check is_published = true in database
- Verify API endpoint returns data
- Check browser console for errors

**Issue**: 404 on frontend pages
- Verify frontend files in `frontend/public/`
- Check server path configuration
- Restart server with `make dev`

---

## Sign-off

**UAT Completed By**: _________________
**Date**: _________________
**Overall Result**: ⬜ Pass ⬜ Fail with Issues
**Production Ready**: ⬜ Yes ⬜ No

**Notes**:
_______________________________________
_______________________________________
