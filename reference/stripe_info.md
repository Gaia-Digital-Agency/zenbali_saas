# Zen Bali Stripe Payment Integration Guide

## Overview

Zen Bali uses **Stripe** for payment processing. Creators pay **$10 USD per event posting**, and events are automatically published upon successful payment.

## Current Status
✅ **Stripe Integration is FULLY IMPLEMENTED and ACTIVE**

## Business Model

### Payment Flow
```
Creator creates event (unpublished, unpaid)
    ↓
Creator clicks "Pay to Publish"
    ↓
Backend creates Stripe Checkout Session
    ↓
Creator redirected to Stripe payment page
    ↓
Creator enters card details on Stripe (secure, PCI compliant)
    ↓
Payment processed by Stripe
    ↓
Stripe sends webhook to backend
    ↓
Backend marks event as paid and published
    ↓
Event appears on public site
```

### Pricing
- **Event Posting Fee**: $10.00 USD (1000 cents)
- **Payment Method**: Credit/Debit Card via Stripe Checkout
- **Mode**: One-time payment (not subscription)

## Stripe Configuration

### Environment Variables

Located in `.env`:

```bash
# Stripe Configuration (Test Keys)
STRIPE_SECRET_KEY=sk_test_xxxxxxxxxxxxxxxxxxxx
STRIPE_PUBLISHABLE_KEY=pk_test_xxxxxxxxxxxxxxxxxxxx
STRIPE_WEBHOOK_SECRET=whsec_xxxxxxxxxxxxxxxxxxxx
STRIPE_PRICE_CENTS=1000
```

**Key Types:**

**Test Keys** (for development):
- `sk_test_...` - Test Secret Key
- `pk_test_...` - Test Publishable Key
- `whsec_test_...` - Test Webhook Secret

**Production Keys** (for live environment):
- `sk_live_...` - Live Secret Key
- `pk_live_...` - Live Publishable Key
- `whsec_...` - Live Webhook Secret

⚠️ **NEVER commit real API keys to version control!**

### Configuration Structure

Defined in [backend/internal/config/config.go:37-43](backend/internal/config/config.go#L37-L43):

```go
type StripeConfig struct {
    SecretKey      string
    PublishableKey string
    WebhookSecret  string
    PriceCents     int64
}
```

Loaded from environment:
```go
Stripe: StripeConfig{
    SecretKey:      getEnv("STRIPE_SECRET_KEY", ""),
    PublishableKey: getEnv("STRIPE_PUBLISHABLE_KEY", ""),
    WebhookSecret:  getEnv("STRIPE_WEBHOOK_SECRET", ""),
    PriceCents:     int64(getEnvInt("STRIPE_PRICE_CENTS", 1000)),
}
```

### Stripe Initialization

In [backend/cmd/server/main.go:43-44](backend/cmd/server/main.go#L43-L44):

```go
import "github.com/stripe/stripe-go/v76"

// Initialize Stripe
stripe.Key = cfg.Stripe.SecretKey
```

## Payment Service

### Service Structure

**Implementation**: [backend/internal/services/payment_service.go](backend/internal/services/payment_service.go)

```go
type PaymentService struct {
    repos  *repository.Repositories
    config config.StripeConfig
}
```

### Key Methods

**1. CreateCheckoutSession** - Create Stripe payment session
**2. HandleSuccessfulPayment** - Process successful payment webhook
**3. HandleFailedPayment** - Process failed/expired payment
**4. ListByCreator** - List payments for a creator
**5. ListAll** - List all payments (admin)

## Creating a Payment

### API Endpoint

**Endpoint**: `POST /api/creator/events/{id}/pay`
**Authentication**: Required (Creator JWT)
**Handler**: [backend/internal/handlers/creator_handler.go](backend/internal/handlers/creator_handler.go)

### Request

```bash
POST /api/creator/events/550e8400-e29b-41d4-a716-446655440000/pay
Authorization: Bearer {creator_jwt_token}
Content-Type: application/json

{
  "success_url": "http://localhost:8080/creator/payment-success.html?session_id={CHECKOUT_SESSION_ID}",
  "cancel_url": "http://localhost:8080/creator/payment-cancel.html"
}
```

### Response

```json
{
  "success": true,
  "data": {
    "session_id": "cs_test_a1b2c3d4...",
    "session_url": "https://checkout.stripe.com/c/pay/cs_test_a1b2c3d4..."
  }
}
```

### What Happens

**Implementation**: [backend/internal/services/payment_service.go:32-86](backend/internal/services/payment_service.go#L32-L86)

1. **Verify Event**
   - Check if event belongs to authenticated creator
   - Check if event is already paid

2. **Create Stripe Checkout Session**
   ```go
   params := &stripe.CheckoutSessionParams{
       PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
       LineItems: []*stripe.CheckoutSessionLineItemParams{
           {
               PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
                   Currency: stripe.String("usd"),
                   ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
                       Name:        stripe.String("Event Posting Fee - " + event.Title),
                       Description: stripe.String("Zen Bali event posting fee"),
                   },
                   UnitAmount: stripe.Int64(1000), // $10.00
               },
               Quantity: stripe.Int64(1),
           },
       },
       Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
       SuccessURL: stripe.String(successURL),
       CancelURL:  stripe.String(cancelURL),
       Metadata: map[string]string{
           "event_id":   event.ID.String(),
           "creator_id": event.CreatorID.String(),
       },
   }
   ```

3. **Store Payment Record**
   - Create record in `payments` table
   - Status: "pending"
   - Link to event and creator

4. **Return Session URL**
   - Frontend redirects user to `session_url`
   - User completes payment on Stripe's secure page

## Stripe Checkout Session

### What Users See

1. **Redirected to Stripe Checkout**
   - Secure, hosted payment page
   - Stripe-branded interface
   - Mobile-friendly

2. **Payment Form**
   - Card number
   - Expiry date
   - CVC
   - Billing details

3. **Payment Processing**
   - Real-time validation
   - 3D Secure if required
   - Instant confirmation

4. **Success/Cancel**
   - On success: Redirect to `success_url`
   - On cancel: Redirect to `cancel_url`

### Session Metadata

Stripe sessions include metadata to track:
```go
Metadata: map[string]string{
    "event_id":   "550e8400-e29b-41d4-a716-446655440000",
    "creator_id": "770e8400-e29b-41d4-a716-446655440001",
}
```

This allows the webhook handler to identify which event to publish.

## Webhook Integration

### Webhook Endpoint

**Endpoint**: `POST /api/webhooks/stripe`
**Authentication**: Stripe signature verification
**Handler**: [backend/internal/handlers/webhook_handler.go](backend/internal/handlers/webhook_handler.go)

### Webhook Events Handled

**1. checkout.session.completed**
- Payment successful
- Marks payment as "completed"
- Publishes the event

**2. checkout.session.expired**
- Session expired without payment
- Marks payment as "failed"

### Setting Up Webhook Secret for Development

**Order of Operations:**

**1. Install Stripe CLI** (do this first):
```bash
brew install stripe/stripe-cli/stripe
```

**2. Login to Stripe** (one-time setup):
```bash
stripe login
```
This will open your browser to authenticate with your Stripe account.

**3. Start your app server** (must be running first):
```bash
make dev
# or
make run
```
Your server should be running on `http://localhost:8080`

**4. In a separate terminal, start webhook listener** (while app is running):
```bash
stripe listen --forward-to localhost:8080/api/webhooks/stripe
```

This will output something like:
```
Ready! Your webhook signing secret is whsec_xxxxxxxxxxxxxxxxxxxxx
```

**5. Copy the webhook secret** and update your `.env` file:
```bash
STRIPE_WEBHOOK_SECRET=whsec_xxxxxxxxxxxxxxxxxxxxx
```

**6. Restart your app server** (to load the new secret):
```bash
# Stop the server (Ctrl+C) and restart
make dev
```

**Development Workflow:**

You'll need **two terminal windows** running simultaneously:

```
Terminal 1:                    Terminal 2:
─────────────                  ─────────────
$ make dev
Server running on :8080
                              $ stripe listen --forward-to localhost:8080/api/webhooks/stripe
                              Ready! Your webhook signing secret is whsec_xxx...

                              (Webhooks forwarded here)
                              ↓
Webhook received ←────────────┘
Payment processed
```

**Important Notes:**
- ⚠️ Keep both terminals running during development
- ⚠️ The webhook secret from `stripe listen` is only valid while the listener is running
- ⚠️ Each time you restart `stripe listen`, you get a new webhook secret
- ✅ The webhook URL for development is: `http://localhost:8080/api/webhooks/stripe`

**Alternative: Skip Webhook Verification in Dev** (Not Recommended)

If you don't want to use Stripe CLI, you can temporarily leave the webhook secret empty:
```bash
STRIPE_WEBHOOK_SECRET=
```

The webhook handler will skip signature verification when empty. **⚠️ ONLY for local development - NEVER in production!**

### Webhook Security

**Signature Verification**: [backend/internal/handlers/webhook_handler.go:35-44](backend/internal/handlers/webhook_handler.go#L35-L44)

```go
webhookSecret := h.services.Payment.GetWebhookSecret()

if webhookSecret != "" {
    sig := r.Header.Get("Stripe-Signature")
    evt, err := webhook.ConstructEvent(payload, sig, webhookSecret)
    if err != nil {
        log.Printf("Stripe webhook signature verification failed: %v", err)
        utils.BadRequest(w, "Invalid signature")
        return
    }
    // Process verified event
}
```

**Security Benefits:**
- Prevents forged webhook calls
- Ensures webhooks are from Stripe
- Protects against replay attacks

### Successful Payment Handling

**Implementation**: [backend/internal/services/payment_service.go:88-115](backend/internal/services/payment_service.go#L88-L115)

```go
func (s *PaymentService) HandleSuccessfulPayment(ctx context.Context, sessionID string) error {
    // 1. Get payment record by Stripe session ID
    payment, err := s.repos.Payment.GetByStripeSessionID(ctx, sessionID)
    if err != nil {
        return err
    }

    // 2. Get Stripe session details (includes payment_intent)
    sess, err := session.Get(sessionID, nil)
    if err != nil {
        return err
    }

    // 3. Update payment status to "completed"
    if err := s.repos.Payment.UpdateStatus(ctx, payment.ID, models.PaymentStatusCompleted, sess.PaymentIntent.ID); err != nil {
        return err
    }

    // 4. Mark event as paid and published
    if err := s.repos.Event.UpdatePaymentStatus(ctx, payment.EventID, true, true); err != nil {
        return err
    }

    return nil
}
```

**What Happens:**
1. Payment marked as "completed"
2. Payment Intent ID stored
3. Event `is_paid` set to `true`
4. Event `is_published` set to `true`
5. Event now visible on public site

## Payment Database

### Payments Table

**Schema**: [backend/internal/database/migrations/001_init.up.sql:108-125](backend/internal/database/migrations/001_init.up.sql#L108-L125)

```sql
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    creator_id UUID NOT NULL REFERENCES creators(id) ON DELETE CASCADE,
    stripe_session_id VARCHAR(255),
    stripe_payment_intent_id VARCHAR(255),
    amount_cents INTEGER NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_payments_stripe_session ON payments(stripe_session_id);
CREATE INDEX idx_payments_status ON payments(status);
```

### Payment Statuses

**Defined in**: `backend/internal/models/payment.go`

```go
const (
    PaymentStatusPending   = "pending"     // Session created, awaiting payment
    PaymentStatusCompleted = "completed"   // Payment successful
    PaymentStatusFailed    = "failed"      // Payment failed or expired
)
```

### Payment Flow States

```
Event Created (unpaid, unpublished)
    ↓
Payment Session Created (pending)
    ↓
    ├─> Payment Successful (completed) → Event Published
    └─> Payment Failed/Expired (failed) → Event Remains Unpublished
```

## Testing Stripe Integration

### Test Mode

With test API keys, you can test without real money.

**Test Card Numbers** (provided by Stripe):

| Card Number | Description |
|-------------|-------------|
| `4242 4242 4242 4242` | Successful payment |
| `4000 0025 0000 3155` | Requires 3D Secure authentication |
| `4000 0000 0000 9995` | Payment declined (insufficient funds) |
| `4000 0000 0000 0002` | Payment declined (generic decline) |

**Test Card Details:**
- **Expiry**: Any future date (e.g., 12/34)
- **CVC**: Any 3 digits (e.g., 123)
- **Zip**: Any 5 digits (e.g., 12345)

### Testing Webhooks Locally

**Option 1: Stripe CLI** (Recommended)

```bash
# Install Stripe CLI
brew install stripe/stripe-cli/stripe

# Login to Stripe
stripe login

# Forward webhooks to local server
stripe listen --forward-to localhost:8080/api/webhooks/stripe

# This will output a webhook signing secret like:
# whsec_xxxxxxxxxxxxxxxxxxxxx

# Update .env with the webhook secret
STRIPE_WEBHOOK_SECRET=whsec_xxxxxxxxxxxxxxxxxxxxx

# Trigger test webhook
stripe trigger checkout.session.completed
```

**Option 2: Webhook Test Mode**

In development, if `STRIPE_WEBHOOK_SECRET` is empty, the webhook handler accepts unsigned webhooks for testing.

### Manual Testing Flow

1. **Create Event**
   ```bash
   curl -X POST http://localhost:8080/api/creator/events \
     -H "Authorization: Bearer {creator_token}" \
     -H "Content-Type: application/json" \
     -d '{
       "title": "Test Event",
       "event_date": "2026-02-01",
       "location_id": 1,
       "event_type_id": 1,
       "entrance_type_id": 1,
       "contact_email": "test@example.com"
     }'
   ```

2. **Create Payment Session**
   ```bash
   curl -X POST http://localhost:8080/api/creator/events/{event_id}/pay \
     -H "Authorization: Bearer {creator_token}" \
     -H "Content-Type: application/json" \
     -d '{
       "success_url": "http://localhost:8080/success",
       "cancel_url": "http://localhost:8080/cancel"
     }'
   ```

3. **Get session_url from response and open in browser**

4. **Complete payment with test card** (`4242 4242 4242 4242`)

5. **Verify event is published**
   ```bash
   curl http://localhost:8080/api/events
   ```

## Viewing Payments

### Creator Payments

**Endpoint**: `GET /api/creator/payments`
**Authentication**: Creator JWT

**Response:**
```json
{
  "success": true,
  "data": {
    "payments": [
      {
        "id": "payment-uuid",
        "event_id": "event-uuid",
        "amount_cents": 1000,
        "currency": "USD",
        "status": "completed",
        "stripe_session_id": "cs_test_...",
        "stripe_payment_intent_id": "pi_...",
        "created_at": "2026-01-10T10:30:00Z"
      }
    ],
    "total": 5,
    "page": 1,
    "limit": 20,
    "total_pages": 1
  }
}
```

### Admin Payments

**Endpoint**: `GET /api/admin/payments`
**Authentication**: Admin JWT
**Query Params**: `?page=1&limit=20&status=completed`

Admins can view all payments across all creators.

## Stripe Dashboard

### Access
- **Test Mode**: https://dashboard.stripe.com/test
- **Live Mode**: https://dashboard.stripe.com/

### What You Can See

**Payments Tab:**
- All payments and their status
- Payment details and metadata
- Customer information
- Refund capability

**Customers Tab:**
- List of customers who paid
- Payment history per customer

**Events Tab:**
- Webhook event log
- Event delivery status
- Retry failed webhooks

**Webhooks Tab:**
- Configure webhook endpoints
- View webhook signing secrets
- Test webhook delivery

**Logs Tab:**
- API request logs
- Debug API issues

## Common Stripe Operations

### Refund a Payment

**Via Dashboard:**
1. Go to Payments tab
2. Find the payment
3. Click "Refund"
4. Enter amount (partial or full)
5. Confirm refund

**Via API** (if implemented):
```go
import "github.com/stripe/stripe-go/v76/refund"

params := &stripe.RefundParams{
    PaymentIntent: stripe.String("pi_xxxxx"),
    Amount:        stripe.Int64(1000), // Full refund
}
refund, err := refund.New(params)
```

**Note**: Current implementation does not handle refunds automatically. Would need to:
1. Add refund webhook handler
2. Update event status when refunded
3. Unpublish event (optional)

### View Payment Details

```bash
# Using Stripe CLI
stripe payments retrieve pi_xxxxxxxxxxxxx

# Response includes:
# - Amount
# - Status
# - Customer details
# - Metadata (event_id, creator_id)
```

### Export Payment Data

**Via Dashboard:**
1. Go to Payments
2. Click "Export"
3. Select date range
4. Choose format (CSV, Excel)
5. Download

**Via API** (Admin endpoint exists):
```bash
GET /api/admin/payments/export
```

## Error Handling

### Payment Errors

**Already Paid:**
```go
if event.IsPaid {
    return ErrAlreadyPaid // HTTP 400
}
```

**Payment Not Found:**
```go
if payment == nil {
    return ErrPaymentNotFound // HTTP 404
}
```

**Stripe API Error:**
```go
sess, err := session.New(params)
if err != nil {
    // Log error
    return err // HTTP 500
}
```

### Webhook Errors

**Invalid Signature:**
```
HTTP 400 Bad Request
"Invalid signature"
```

**Event Not Found:**
```
HTTP 500 Internal Server Error
"Error processing payment"
```

## Production Deployment

### Before Going Live

**1. Get Live API Keys**
- Log in to Stripe Dashboard
- Switch to Live mode
- Copy live keys from Developers > API Keys

**2. Update Environment Variables**
```bash
STRIPE_SECRET_KEY=sk_live_xxxxxxxxxxxxx
STRIPE_PUBLISHABLE_KEY=pk_live_xxxxxxxxxxxxx
STRIPE_WEBHOOK_SECRET=whsec_xxxxxxxxxxxxx
```

**3. Configure Webhook Endpoint**
- Go to Developers > Webhooks
- Click "Add endpoint"
- URL: `https://zenbali.org/api/webhooks/stripe`
- Events to send:
  - `checkout.session.completed`
  - `checkout.session.expired`
- Copy webhook signing secret

**4. Test in Live Mode**
- Create test event
- Pay with real card (small amount)
- Verify webhook received
- Verify event published
- Refund test payment

**5. Enable Payment Methods**

In Stripe Dashboard, enable:
- Credit/Debit Cards (default)
- Apple Pay
- Google Pay
- Link (Stripe's 1-click checkout)

**6. Configure Radar (Fraud Prevention)**
- Review Stripe Radar rules
- Set up fraud alerts
- Configure risk thresholds

**7. Set Up Email Receipts**
- Configure receipt emails in Stripe
- Customize email template
- Add your logo and branding

## Stripe Fees

### Standard Pricing
- **2.9% + $0.30** per successful card charge
- **$10 event posting** = **$0.59** Stripe fee
- **You receive**: **$9.41** per event

### International Cards
- Additional **1.5%** for international cards
- Additional **1%** for currency conversion

### Example Calculation
```
Event posting fee:     $10.00
Stripe fee (2.9%):     -$0.29
Stripe fee (fixed):    -$0.30
Net received:          $9.41
```

## Security Best Practices

### API Key Management ✅

- ✅ Never commit API keys to Git
- ✅ Use environment variables
- ✅ Separate test and live keys
- ✅ Rotate keys periodically

### Webhook Security ✅

- ✅ Verify webhook signatures
- ✅ Use HTTPS in production
- ✅ Validate event data
- ✅ Implement idempotency

### Payment Security ✅

- ✅ Never store card numbers
- ✅ Use Stripe Checkout (PCI compliant)
- ✅ Validate payment amounts server-side
- ✅ Check payment status before publishing

### Additional Recommendations ⚠️

**1. Implement Idempotency**
```go
// Prevent duplicate payment processing
if payment.Status == "completed" {
    return nil // Already processed
}
```

**2. Add Webhook Retry Logic**
```go
// Store webhook events in database
// Retry failed processing
```

**3. Monitor Failed Payments**
```go
// Alert on high failure rates
// Track common failure reasons
```

**4. Implement Refund Handling**
```go
// Listen for refund webhooks
// Update event status
// Notify creator
```

## Monitoring & Logging

### What to Log

**Payment Creation:**
```go
log.Printf("Payment session created: %s for event %s", sessionID, eventID)
```

**Successful Payment:**
```go
log.Printf("Successfully processed payment for session: %s", session.ID)
```

**Failed Payment:**
```go
log.Printf("Error handling payment: %v", err)
```

**Webhook Events:**
```go
log.Printf("Received Stripe webhook: %s", eventType)
log.Printf("Unhandled Stripe event type: %s", evt.Type)
```

### Monitoring Metrics

Track these in production:
- **Payment success rate**
- **Payment failure rate**
- **Average payment time**
- **Webhook delivery success**
- **Refund rate**

## Troubleshooting

### Payment Not Processing

**Check:**
1. Stripe API keys are correct
2. Webhook endpoint is accessible
3. Webhook signature verification passes
4. Event ID in metadata is valid
5. Payment status in database

**Debug:**
```bash
# Check Stripe logs
stripe logs tail

# Check webhook deliveries
# Go to Dashboard > Webhooks > {endpoint} > Events
```

### Event Not Publishing After Payment

**Check:**
1. Webhook received and processed
2. Payment status updated to "completed"
3. Event `is_paid` and `is_published` set to true
4. Database queries successful

**Query Database:**
```sql
-- Check payment status
SELECT * FROM payments WHERE stripe_session_id = 'cs_test_xxx';

-- Check event status
SELECT id, title, is_paid, is_published FROM events WHERE id = 'event-uuid';
```

### Webhook Signature Verification Failed

**Causes:**
- Wrong webhook secret
- Webhook secret from different mode (test vs live)
- Payload modified in transit

**Fix:**
1. Get correct webhook secret from Stripe Dashboard
2. Update `.env` file
3. Restart server
4. Test with Stripe CLI

## Additional Resources

- **Stripe Documentation**: https://stripe.com/docs
- **Stripe Checkout**: https://stripe.com/docs/payments/checkout
- **Webhooks Guide**: https://stripe.com/docs/webhooks
- **Testing**: https://stripe.com/docs/testing
- **Stripe Go Library**: https://github.com/stripe/stripe-go
- **Stripe Dashboard**: https://dashboard.stripe.com/
- **Stripe CLI**: https://stripe.com/docs/stripe-cli
- **API Reference**: https://stripe.com/docs/api
