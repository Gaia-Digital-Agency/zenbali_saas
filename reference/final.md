# Zen Bali Final Deployment Note

Last updated: `2026-03-27`

## Current Local State

- App: `http://localhost:8081`
- API: `http://localhost:8081/api`
- Admin login: `admin@zenbali.org` / `Teameditor@123`
- Creator login: `creator@zenbali.org`
- Event posting fee: `$1 USD`

## Current VM Deployment State

- Host: `34.124.244.233`
- SSH user: `azlan`
- App path: `/var/www/zenbali`
- App is deployed and running behind Nginx
- Backend app port on VM: `127.0.0.1:8081`
- Public base URL in use: `https://zenbali.site`
- Canonical public URL: `https://zenbali.site`
- Service: `zenbali.service`
- Dedicated Nginx site file: `/etc/nginx/sites-available/zenbali.site`
- Database: VM PostgreSQL on `127.0.0.1:5432`

## Current Production Config on VM

- `UPLOAD_BACKEND=gcs`
- `GCS_BUCKET=gda-s01-bucket`
- `GCS_PREFIX=zenbali`
- `GCS_PUBLIC_BASE_URL=https://storage.googleapis.com/gda-s01-bucket`
- Stripe is configured with live keys
- Stripe webhook secret is configured
- Stripe price is `100` cents

## Verified VM Status

- `azlan` has write access to `/var/www/zenbali`
- PostgreSQL database `zenbali` exists
- PostgreSQL role `zenbali` exists
- Nginx config test passed
- HTTPS is live on `zenbali.site`
- Health check passed
- Admin login works on deployed app
- Creator login works on deployed app
- Stripe webhook endpoint is reachable and rejects invalid signatures as expected

## Remaining Operational Checks

1. Run one real live payment and verify event publication.
2. Run one real GCS image upload from the deployed site.
3. Refresh reference docs if pricing or credentials policy changes again.
