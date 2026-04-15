# Live Payment Activation

Stripe payment on production is now configured for live mode.

Current verified production state on `https://zenbali.site`:

- `STRIPE_SECRET_KEY=sk_live_...`
- `STRIPE_PUBLISHABLE_KEY=pk_live_...`
- `STRIPE_WEBHOOK_SECRET=whsec_...`
- `STRIPE_PRICE_CENTS=300`
- `BASE_URL=https://zenbali.site`

## Live Webhook

Stripe webhook endpoint:

```text
https://zenbali.site/api/webhooks/stripe
```

Subscribed events:

- `checkout.session.completed`
- `checkout.session.expired`

The webhook now enforces signature verification in production.

## Current Price

- Event posting fee: `$3 USD`
- Stripe amount: `300` cents

## Recommended Verification

1. Create or use an unpaid event.
2. Start checkout from the creator portal.
3. Confirm Stripe Checkout shows `$3.00 USD`.
4. Complete a real payment.
5. Verify:
   - success page returns to `https://zenbali.site`
   - webhook delivery succeeds in Stripe Dashboard
   - payment row is marked completed
   - event becomes `is_paid=true`
   - event becomes `is_published=true`

## Useful Checks

```bash
grep -nE '^STRIPE_|^BASE_URL' /var/www/zenbali/.env
journalctl -u zenbali.service -f
```
