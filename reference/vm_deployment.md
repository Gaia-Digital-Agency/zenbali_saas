# Zen Bali VM Deployment Plan

> Current verified local state as of 2026-03-23:
> - App: `http://localhost:8081`
> - API: `http://localhost:8081/api`
> - PostgreSQL host port: `5433`
> - Admin: `admin@zenbali.org` / `Teameditor@123`
> - Creator: `creator@zenbali.org` / `admin123`
> - Event posting fee: `$5 USD` (`500` cents)
> - Seed base state: 1 admin, 1 creator, 1 sample published event, 0 payments

## Target Shape

- App folder: `/var/www/zenbali`
- Host: `34.124.244.233`
- SSH user: `azlan`
- SSH key: `~/.ssh/gda-ce01`
- Database: VM PostgreSQL on `127.0.0.1:5432`
- Image target: `gs://gda-s01-bucket/zenbali/`

## Current Verified Server State

- SSH access works
- `/var/www/zenbali` is populated with the deployed app
- `azlan` has write access to `/var/www/zenbali`
- PostgreSQL is active and listening on `127.0.0.1:5432`
- Database `zenbali` exists
- Role `zenbali` exists
- Nginx is a shared multi-site setup
- `gs://gda-s01-bucket/zenbali/` is accessible
- the VM has active Google credentials through its attached service account
- Zen Bali is installed as `zenbali.service`
- Zen Bali is running on `127.0.0.1:8081`
- Nginx is proxying the server IP to Zen Bali

## Current Upload Behavior

The application now supports both upload backends:
- `UPLOAD_BACKEND=local` saves files under `UPLOAD_DIR` and serves them from `/uploads/...`
- `UPLOAD_BACKEND=gcs` writes files to the configured GCS bucket/prefix and returns public bucket URLs

For this VM target, use:
- `UPLOAD_BACKEND=gcs`
- `GCS_BUCKET=gda-s01-bucket`
- `GCS_PREFIX=zenbali`
- `GCS_PUBLIC_BASE_URL=https://storage.googleapis.com/gda-s01-bucket`

Important runtime requirement:
- the VM must have Google Application Default Credentials with write access to `gs://gda-s01-bucket/zenbali/`
- that can be either a service account attached to the VM, or `GOOGLE_APPLICATION_CREDENTIALS` pointing to a JSON key file

## Current Deployment Notes

- The VM does not have Go installed
- The safe deployment flow is:
  - build the Linux binary locally
  - sync the repo to `/var/www/zenbali`
  - copy the binary into `/var/www/zenbali/bin/`
  - restart `zenbali.service`
- Backend port `8080` was already occupied by another site on the VM
- Zen Bali therefore uses backend port `8081`

## Recommended Deployment Sequence

1. Prepare server permissions.
- Either grant `azlan` write access to `/var/www/zenbali`
- Or deploy with a privileged copy step

2. Prepare PostgreSQL on the VM.
- Create DB user if needed
- Create database `zenbali` if needed
- Confirm app credentials match the production env file

3. Upload application code to `/var/www/zenbali`.
- Preferred: clone/pull repo there
- Alternative: rsync/copy release bundle

4. Create production env file.
- Use [vm.env.example](/Users/rogerwoolie/Documents/gaiada_projects/zenbali_saas/reference/vm.env.example) as the baseline
- Save it on the server as `/var/www/zenbali/.env`

5. Build and run the app on the VM.
- Build the Go binary on the VM or upload a Linux build
- Run under `systemd`, not `nohup`

6. Configure Nginx carefully.
- Add a dedicated Zen Bali server block
- Reverse proxy to the app port
- Do not disturb other sites on the VM

7. Run migrations.
- Apply:
  - `001_init.up.sql`
  - `002_seed_data.up.sql`
  - `003_add_participant_group_and_lead_by.up.sql`

8. Verify production basics.
- `/api/health`
- homepage
- admin login
- creator login
- event create/edit
- image upload to GCS
- Stripe payment

## Suggested VM Runtime

Use:
- app port `8081`
- Nginx reverse proxy in front
- `systemd` service for the Go binary

Exact service file:
- [zenbali.service](/Users/rogerwoolie/Documents/gaiada_projects/zenbali_saas/reference/systemd/zenbali.service)

## Suggested Nginx Shape

Use a dedicated site file for Zen Bali and proxy to `127.0.0.1:8081`.

Exact site file:
- [zenbali.conf](/Users/rogerwoolie/Documents/gaiada_projects/zenbali_saas/reference/nginx/zenbali.conf)

## PostgreSQL Preparation

Example commands on the VM:

```bash
sudo -u postgres psql
```

```sql
CREATE USER zenbali WITH PASSWORD 'CHANGE_ME_DB_PASSWORD';
CREATE DATABASE zenbali OWNER zenbali;
GRANT ALL PRIVILEGES ON DATABASE zenbali TO zenbali;
```

## Migration Commands

From `/var/www/zenbali/backend`:

```bash
PGPASSWORD=CHANGE_ME_DB_PASSWORD psql -h 127.0.0.1 -p 5432 -U zenbali -d zenbali < internal/database/migrations/001_init.up.sql
PGPASSWORD=CHANGE_ME_DB_PASSWORD psql -h 127.0.0.1 -p 5432 -U zenbali -d zenbali < internal/database/migrations/002_seed_data.up.sql
PGPASSWORD=CHANGE_ME_DB_PASSWORD psql -h 127.0.0.1 -p 5432 -U zenbali -d zenbali < internal/database/migrations/003_add_participant_group_and_lead_by.up.sql
```

## First Deploy Checklist

- `/var/www/zenbali` writable by deploy user or deploy process
- `.env` present
- Google credentials present if the VM is not already using a Storage-capable service account
- `bin/zenbali-server` built
- PostgreSQL database ready
- Nginx site isolated from other sites
- systemd service enabled
- Stripe webhook endpoint configured after domain is live

## GCS Notes

Target bucket path:
- `gs://gda-s01-bucket/zenbali/`

Behavior:
- new image uploads are stored in the configured bucket prefix
- replacing an event image deletes the previous object
- deleting an event deletes its uploaded object

## Current Live Gaps

- public access is still IP-based, not domain-based
- TLS is not configured yet
- the VM still uses Stripe test keys
- `STRIPE_WEBHOOK_SECRET` is currently blank on the VM
- the sample seeded event image URL was cleared on the VM because the seed pointed to a missing local file
