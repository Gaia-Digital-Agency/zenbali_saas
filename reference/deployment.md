# Zen Bali - Deployment Guide

**Last Updated:** 2026-01-13

This guide covers deploying the Zen Bali application to Google Cloud Platform (GCP) using Cloud Run, Cloud SQL, and Cloud Storage.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Local Development](#local-development)
3. [Production Deployment](#production-deployment)
4. [Environment Variables](#environment-variables)
5. [Database Migrations](#database-migrations)
6. [Monitoring & Maintenance](#monitoring--maintenance)
7. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required Tools
- **gcloud CLI** - [Install Guide](https://cloud.google.com/sdk/docs/install)
- **Docker Desktop** - [Download](https://www.docker.com/products/docker-desktop)
- **Go 1.22+** - [Download](https://golang.org/dl/)
- **Git** - [Download](https://git-scm.com/downloads)

### GCP Setup
1. Create a Google Cloud Platform account
2. Create a new project: `zenbali-production`
3. Enable billing for the project
4. Enable required APIs:
   ```bash
   gcloud services enable \
     run.googleapis.com \
     sql-component.googleapis.com \
     sqladmin.googleapis.com \
     cloudresourcemanager.googleapis.com \
     storage.googleapis.com \
     secretmanager.googleapis.com
   ```

### Domain Setup
1. Register domain: `zenbali.org`
2. Configure Cloudflare for DNS and CDN
3. Point DNS to Cloud Run (will be configured during deployment)

---

## Local Development

### Quick Start

```bash
# Clone repository
git clone https://github.com/net1io/zenbali.git
cd zenbali

# Start application
./start.sh

# Access at http://localhost:8080
```

### Manual Start

```bash
# Start Docker services
docker-compose up -d

# Run migrations
cd backend
cat internal/database/migrations/001_init.up.sql | docker exec -i zenbali-postgres psql -U zenbali -d zenbali
cat internal/database/migrations/002_seed_data.up.sql | docker exec -i zenbali-postgres psql -U zenbali -d zenbali
cat internal/database/migrations/003_add_participant_group_and_lead_by.up.sql | docker exec -i zenbali-postgres psql -U zenbali -d zenbali

# Start server
go run ./cmd/server
```

### Default Admin Credentials
- Email: `admin@zenbali.org`
- Password: `admin123`

### Stop Services

```bash
./stop.sh
```

---

## Production Deployment

### Step 1: Set Up Cloud SQL (PostgreSQL)

```bash
# Set project
gcloud config set project zenbali-production

# Create Cloud SQL instance
gcloud sql instances create zenbali-db \
  --database-version=POSTGRES_15 \
  --tier=db-f1-micro \
  --region=asia-southeast1 \
  --storage-type=SSD \
  --storage-size=10GB \
  --storage-auto-increase \
  --backup-start-time=03:00 \
  --retained-backups-count=7

# Create database
gcloud sql databases create zenbali --instance=zenbali-db

# Create database user
gcloud sql users create zenbali \
  --instance=zenbali-db \
  --password=GENERATE_STRONG_PASSWORD_HERE
```

### Step 2: Set Up Cloud Storage (Images)

```bash
# Create bucket for event images
gsutil mb -c STANDARD -l asia-southeast1 gs://zenbali-event-images

# Make bucket publicly readable
gsutil iam ch allUsers:objectViewer gs://zenbali-event-images

# Enable CORS
cat > cors.json <<EOF
[
  {
    "origin": ["https://zenbali.org", "http://localhost:8080"],
    "method": ["GET", "POST", "DELETE"],
    "responseHeader": ["Content-Type"],
    "maxAgeSeconds": 3600
  }
]
EOF

gsutil cors set cors.json gs://zenbali-event-images
```

### Step 3: Set Up Secret Manager

```bash
# Create secrets
echo -n "YOUR_DB_PASSWORD" | gcloud secrets create db-password --data-file=-
echo -n "YOUR_JWT_SECRET_MIN_32_CHARS" | gcloud secrets create jwt-secret --data-file=-
echo -n "sk_live_YOUR_STRIPE_SECRET" | gcloud secrets create stripe-secret-key --data-file=-
echo -n "whsec_YOUR_STRIPE_WEBHOOK" | gcloud secrets create stripe-webhook-secret --data-file=-
echo -n "YOUR_ADMIN_PASSWORD" | gcloud secrets create admin-password --data-file=-

# Grant Cloud Run access to secrets
PROJECT_NUMBER=$(gcloud projects describe zenbali-production --format="value(projectNumber)")
gcloud secrets add-iam-policy-binding db-password \
  --member="serviceAccount:${PROJECT_NUMBER}-compute@developer.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"

# Repeat for all secrets
```

### Step 4: Run Database Migrations

```bash
# Connect to Cloud SQL using proxy
cloud_sql_proxy -instances=zenbali-production:asia-southeast1:zenbali-db=tcp:5433

# In another terminal
cd backend
PGPASSWORD=YOUR_DB_PASSWORD psql -h localhost -p 5433 -U zenbali -d zenbali < internal/database/migrations/001_init.up.sql
PGPASSWORD=YOUR_DB_PASSWORD psql -h localhost -p 5433 -U zenbali -d zenbali < internal/database/migrations/002_seed_data.up.sql
PGPASSWORD=YOUR_DB_PASSWORD psql -h localhost -p 5433 -U zenbali -d zenbali < internal/database/migrations/003_add_participant_group_and_lead_by.up.sql
```

### Step 5: Build and Deploy to Cloud Run

```bash
# Build container
gcloud builds submit --tag gcr.io/zenbali-production/zenbali-backend

# Deploy to Cloud Run
gcloud run deploy zenbali-backend \
  --image gcr.io/zenbali-production/zenbali-backend \
  --platform managed \
  --region asia-southeast1 \
  --allow-unauthenticated \
  --memory 512Mi \
  --cpu 1 \
  --min-instances 0 \
  --max-instances 10 \
  --set-env-vars "ENV=production,PORT=8080,BASE_URL=https://zenbali.org" \
  --set-secrets "DB_PASSWORD=db-password:latest,JWT_SECRET=jwt-secret:latest,STRIPE_SECRET_KEY=stripe-secret-key:latest,STRIPE_WEBHOOK_SECRET=stripe-webhook-secret:latest,ADMIN_PASSWORD=admin-password:latest" \
  --add-cloudsql-instances zenbali-production:asia-southeast1:zenbali-db \
  --set-env-vars "DB_HOST=/cloudsql/zenbali-production:asia-southeast1:zenbali-db,DB_USER=zenbali,DB_NAME=zenbali,DB_SSL_MODE=disable"

# Get the Cloud Run URL
gcloud run services describe zenbali-backend --region asia-southeast1 --format="value(status.url)"
```

### Step 6: Configure Domain Mapping

```bash
# Map custom domain
gcloud run domain-mappings create \
  --service zenbali-backend \
  --domain zenbali.org \
  --region asia-southeast1

# Get DNS records to configure in Cloudflare
gcloud run domain-mappings describe \
  --domain zenbali.org \
  --region asia-southeast1
```

### Step 7: Configure Cloudflare

1. Log in to Cloudflare
2. Add DNS records as shown by the previous command
3. Enable SSL/TLS (Full mode)
4. Enable HTTP/2
5. Configure caching rules:
   - Cache static assets (css, js, images)
   - Don't cache API responses

### Step 8: Configure Stripe Webhook

1. Log in to Stripe Dashboard
2. Go to Developers → Webhooks
3. Add endpoint: `https://zenbali.org/api/stripe/webhook`
4. Select events:
   - `checkout.session.completed`
   - `payment_intent.succeeded`
   - `payment_intent.payment_failed`
5. Copy webhook secret and update Secret Manager

---

## Environment Variables

### Local Development (.env)

```env
PORT=8080
ENV=development
BASE_URL=http://localhost:8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=zenbali
DB_PASSWORD=zenbali_dev_password
DB_NAME=zenbali
DB_SSL_MODE=disable

JWT_SECRET=zenbali-dev-secret-key-change-in-production-min-32-chars
JWT_EXPIRY_HOURS=24

STRIPE_SECRET_KEY=sk_test_...
STRIPE_PUBLISHABLE_KEY=pk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
STRIPE_PRICE_CENTS=1000

UPLOAD_DIR=./uploads
MAX_UPLOAD_SIZE_MB=5

ADMIN_EMAIL=admin@zenbali.org
ADMIN_PASSWORD=admin123
```

### Production (Cloud Run)

Set via Cloud Run environment variables and Secret Manager:

```bash
# Environment Variables
ENV=production
PORT=8080
BASE_URL=https://zenbali.org
DB_HOST=/cloudsql/zenbali-production:asia-southeast1:zenbali-db
DB_USER=zenbali
DB_NAME=zenbali
DB_SSL_MODE=disable
DB_MAX_CONNECTIONS=25
JWT_EXPIRY_HOURS=168
STRIPE_PRICE_CENTS=1000
UPLOAD_DIR=gs://zenbali-event-images
MAX_UPLOAD_SIZE_MB=5
ADMIN_EMAIL=admin@zenbali.org

# Secrets (from Secret Manager)
DB_PASSWORD=<secret>
JWT_SECRET=<secret>
STRIPE_SECRET_KEY=<secret>
STRIPE_PUBLISHABLE_KEY=<secret>
STRIPE_WEBHOOK_SECRET=<secret>
ADMIN_PASSWORD=<secret>
```

---

## Database Migrations

### Running New Migrations in Production

```bash
# 1. Connect to Cloud SQL
cloud_sql_proxy -instances=zenbali-production:asia-southeast1:zenbali-db=tcp:5433

# 2. Create new migration file
# backend/internal/database/migrations/004_your_migration.up.sql

# 3. Apply migration
PGPASSWORD=YOUR_DB_PASSWORD psql -h localhost -p 5433 -U zenbali -d zenbali < backend/internal/database/migrations/004_your_migration.up.sql

# 4. Verify migration
PGPASSWORD=YOUR_DB_PASSWORD psql -h localhost -p 5433 -U zenbali -d zenbali -c "\dt"
```

### Rollback Migration

```bash
# Create rollback file
# backend/internal/database/migrations/004_your_migration.down.sql

# Apply rollback
PGPASSWORD=YOUR_DB_PASSWORD psql -h localhost -p 5433 -U zenbali -d zenbali < backend/internal/database/migrations/004_your_migration.down.sql
```

---

## Monitoring & Maintenance

### View Logs

```bash
# Cloud Run logs
gcloud run logs read --service zenbali-backend --region asia-southeast1 --limit 100

# Follow logs in real-time
gcloud run logs tail --service zenbali-backend --region asia-southeast1
```

### Database Backups

```bash
# List backups
gcloud sql backups list --instance=zenbali-db

# Create on-demand backup
gcloud sql backups create --instance=zenbali-db

# Restore from backup
gcloud sql backups restore BACKUP_ID --backup-instance=zenbali-db --instance=zenbali-db
```

### Performance Monitoring

1. **Cloud Run Metrics**
   - Request count
   - Request latency
   - Container instance count
   - Memory usage

2. **Cloud SQL Metrics**
   - CPU utilization
   - Memory usage
   - Connection count
   - Disk I/O

3. **Set Up Alerts**
   ```bash
   # Create alert for high error rate
   gcloud alpha monitoring policies create \
     --notification-channels=CHANNEL_ID \
     --display-name="High Error Rate" \
     --condition-display-name="Error rate > 5%" \
     --condition-threshold-value=0.05 \
     --condition-threshold-duration=300s
   ```

### Scaling Configuration

```bash
# Update Cloud Run scaling
gcloud run services update zenbali-backend \
  --region asia-southeast1 \
  --min-instances 1 \
  --max-instances 50 \
  --concurrency 80
```

---

## Troubleshooting

### Issue: Database Connection Timeout

**Symptoms:** Cloud Run logs show "failed to connect to database"

**Solution:**
```bash
# Check Cloud SQL instance is running
gcloud sql instances describe zenbali-db

# Verify Cloud Run has SQL connection
gcloud run services describe zenbali-backend --region asia-southeast1 | grep cloudsql

# Increase connection timeout in code or environment
```

### Issue: 502 Bad Gateway

**Symptoms:** Cloudflare returns 502 error

**Solution:**
```bash
# Check Cloud Run service health
gcloud run services describe zenbali-backend --region asia-southeast1

# Check logs for startup errors
gcloud run logs read --service zenbali-backend --region asia-southeast1 --limit 50

# Verify environment variables are set correctly
gcloud run services describe zenbali-backend --region asia-southeast1 --format="value(spec.template.spec.containers[0].env)"
```

### Issue: Stripe Webhook Fails

**Symptoms:** Payments succeed but events don't publish

**Solution:**
1. Check Stripe Dashboard → Webhooks for failed events
2. Verify webhook endpoint URL is correct
3. Ensure webhook secret matches in Secret Manager
4. Check Cloud Run logs for webhook processing errors

### Issue: Image Upload Fails

**Symptoms:** "Failed to upload image" error

**Solution:**
```bash
# Verify bucket exists and is accessible
gsutil ls gs://zenbali-event-images

# Check bucket permissions
gsutil iam get gs://zenbali-event-images

# Verify CORS configuration
gsutil cors get gs://zenbali-event-images
```

---

## Cost Optimization

### Estimated Monthly Costs (Low Traffic)

- **Cloud Run:** $0-5 (pay per use, 0-1000 requests/day)
- **Cloud SQL (db-f1-micro):** ~$10/month
- **Cloud Storage:** $0.026/GB + $0.12/GB egress
- **Secret Manager:** $0.06 per secret per month
- **Total:** ~$15-25/month for low traffic

### Optimization Tips

1. **Cloud Run**
   - Set `min-instances=0` for development
   - Use `min-instances=1` for production to avoid cold starts
   - Adjust memory/CPU based on actual usage

2. **Cloud SQL**
   - Start with `db-f1-micro` (shared CPU)
   - Upgrade to `db-g1-small` if needed
   - Enable auto-storage increase
   - Monitor connection pool usage

3. **Cloud Storage**
   - Use Cloud CDN for image delivery
   - Compress images before upload
   - Set lifecycle policies to delete old backups

---

## Security Checklist

- [ ] All secrets stored in Secret Manager (not in code)
- [ ] Database has strong password
- [ ] JWT secret is at least 32 characters
- [ ] Cloud SQL only accessible via private IP or Cloud SQL Proxy
- [ ] HTTPS enforced (Cloudflare SSL)
- [ ] Stripe webhook signature verified
- [ ] CORS configured for Cloud Storage
- [ ] Admin password changed from default
- [ ] Database backups enabled (7-day retention)
- [ ] Cloud Run service account has minimal permissions
- [ ] API rate limiting enabled (if needed)

---

## Rollback Procedure

If a deployment causes issues:

```bash
# 1. List previous revisions
gcloud run revisions list --service zenbali-backend --region asia-southeast1

# 2. Roll back to previous revision
gcloud run services update-traffic zenbali-backend \
  --region asia-southeast1 \
  --to-revisions REVISION_NAME=100

# 3. Verify rollback
gcloud run services describe zenbali-backend --region asia-southeast1
```

---

## Support & Resources

- **GCP Documentation:** https://cloud.google.com/docs
- **Cloud Run Docs:** https://cloud.google.com/run/docs
- **Cloud SQL Docs:** https://cloud.google.com/sql/docs
- **Stripe API Docs:** https://stripe.com/docs/api
- **Project Repository:** https://github.com/net1io/zenbali

---

**Deployment Guide maintained by net1io.com**
**Last Updated: 2026-01-13**
