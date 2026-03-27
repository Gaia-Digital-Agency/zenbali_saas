# Zen Bali Server Information

> Current verified local state as of 2026-03-23:
> - App: `http://localhost:8081`
> - API: `http://localhost:8081/api`
> - PostgreSQL host port: `5433`
> - Admin: `admin@zenbali.org` / `Teameditor@123`
> - Creator: `creator@zenbali.org` / `admin123`
> - Event posting fee: `$5 USD` (`500` cents)
> - Seed base state: 1 admin, 1 creator, 1 sample published event, 0 payments

## Access

- SSH user: `azlan`
- SSH host: `34.124.244.233`
- SSH key: `~/.ssh/gda-ce01`
- SSH command:

```bash
ssh -i ~/.ssh/gda-ce01 azlan@34.124.244.233
```

## Server Project Path

- Project folder: `/var/www/zenbali`

## Important Caution

- This server is a multi-site setup with Nginx.
- Do not modify shared Nginx files or neighboring site folders casually.
- Validate any Zen Bali deployment change in isolation before touching active site routing.

## Current Deployed Status

Verified over SSH on `2026-03-23`:

- SSH access works with the provided key
- `/var/www/zenbali` is now populated with the deployed app
- `azlan` now has write access to `/var/www/zenbali`
- Zen Bali is installed as `zenbali.service`
- Zen Bali is running on `127.0.0.1:8081`
- Nginx is proxying the server IP to the Zen Bali app
- Health check succeeds at `http://127.0.0.1:8081/api/health`
- Admin login works with `admin@zenbali.org / Teameditor@123`
- Creator login works with `creator@zenbali.org / admin123`

## Nginx Findings

- Nginx uses a shared multi-site configuration.
- `/etc/nginx/sites-enabled/gda-s01` exists.
- Zen Bali now also has a dedicated site file:
  - `/etc/nginx/sites-available/zenbali`
  - `/etc/nginx/sites-enabled/zenbali`
- Port `8080` was already in use by another site, so Zen Bali was moved to backend port `8081`.

## PostgreSQL Findings

Verified over SSH on `2026-03-23`:

- `psql` is installed on the VM
- PostgreSQL service is active
- PostgreSQL is listening on `127.0.0.1:5432`
- Database `zenbali` exists
- Role `zenbali` exists
- Migrations `001`, `002`, and `003` were applied successfully

## GCS Findings

Verified from this machine and the VM on `2026-03-23`:

- `gs://gda-s01-bucket/zenbali/` is accessible
- The VM has active Google credentials via:
  - `292070531785-compute@developer.gserviceaccount.com`
- The app is configured to use GCS-backed uploads on the VM

Successful check:

```text
gs://gda-s01-bucket/zenbali/
```

## Deployment Implications

- Zen Bali is now deployed on the VM and reachable through Nginx on the server IP
- VM PostgreSQL is being used for the deployed app
- GCS target exists for `gs://gda-s01-bucket/zenbali/`
- The app is configured with `UPLOAD_BACKEND=gcs` on the VM
- Exact deploy files are prepared in:
  - [zenbali.service](/Users/rogerwoolie/Documents/gaiada_projects/zenbali_saas/reference/systemd/zenbali.service)
  - [zenbali.conf](/Users/rogerwoolie/Documents/gaiada_projects/zenbali_saas/reference/nginx/zenbali.conf)
- Current gaps before a full production cutover:
  - replace IP-based access with the real domain
  - add TLS
  - replace Stripe test keys with final production Stripe settings
  - set a real `STRIPE_WEBHOOK_SECRET`
