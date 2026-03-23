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

## Verified Server Findings

Checked over SSH on `2026-03-23`:

- SSH access works with the provided key.
- `/var/www/zenbali` exists.
- `/var/www/zenbali` is currently empty.
- Ownership is currently `root:root`.
- The `azlan` user does **not** currently have write access to `/var/www/zenbali`.

## Nginx Findings

- Nginx uses a shared multi-site configuration.
- `/etc/nginx/sites-enabled/gda-s01` exists.
- No obvious dedicated `zenbali` site block was found during the read-only inspection.

## PostgreSQL Findings

Verified over SSH on `2026-03-23`:

- `psql` is installed on the VM
- PostgreSQL service is active
- PostgreSQL is listening on `127.0.0.1:5432`
- This is suitable for a same-VM app deployment using local DB access

## GCS Findings

Verified from this machine on `2026-03-23`:

- `gs://gda-s01-bucket/zenbali/` is accessible
- The prefix currently appears empty

Successful check:

```text
gs://gda-s01-bucket/zenbali/
```

## Deployment Implications

- Server-side deployment will need either:
  - write permission for `azlan` on `/var/www/zenbali`, or
  - a privileged copy/deploy step
- Nginx configuration for Zen Bali should be reviewed carefully before enabling a site
- VM PostgreSQL is available for a same-host deployment
- GCS target exists for `gs://gda-s01-bucket/zenbali/`
- The app now supports `UPLOAD_BACKEND=gcs` for direct image uploads into `gs://gda-s01-bucket/zenbali/`
- The VM must provide Google credentials with write access to that bucket path
- Exact deploy files are prepared in:
  - [zenbali.service](/Users/rogerwoolie/Documents/gaiada_projects/zenbali_saas/reference/systemd/zenbali.service)
  - [zenbali.conf](/Users/rogerwoolie/Documents/gaiada_projects/zenbali_saas/reference/nginx/zenbali.conf)
