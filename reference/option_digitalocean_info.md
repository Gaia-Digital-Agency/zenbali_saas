# Digital Ocean Deployment Option for Zen Bali

## Overview

This guide provides complete information for deploying Zen Bali to Digital Ocean as an alternative to Google Cloud Platform. Your application is currently configured for GCP but is easily portable to Digital Ocean.

**Current Status**: Configured for GCP Cloud Run
**Alternative**: Digital Ocean App Platform or Droplet

---

# Part 1: Changes Required for Digital Ocean

## Current GCP-Specific Code

Your app is already **99% portable**. Only minimal changes needed:

### Files That Reference GCP

**1. Makefile (Deploy Target)**

Current [Makefile:116-127](Makefile#L116-L127):
```makefile
deploy:
	@echo "Building Docker image..."
	@docker build -t gcr.io/$(GCP_PROJECT)/zenbali:latest .
	@echo "Pushing to GCR..."
	@docker push gcr.io/$(GCP_PROJECT)/zenbali:latest
	@echo "Deploying to Cloud Run..."
	@gcloud run deploy zenbali \
		--image gcr.io/$(GCP_PROJECT)/zenbali:latest \
		--platform managed \
		--region asia-southeast1 \
		--allow-unauthenticated
	@echo "Deployment complete"
```

**Digital Ocean Version** (add new target):
```makefile
deploy-do:
	@echo "Deploying to Digital Ocean..."
	@doctl apps create-deployment $(DO_APP_ID)
	@echo "Deployment complete"
```

**2. README.md Architecture Diagram**

Update architecture section to mention both options:
```markdown
## Deployment Options

### Option 1: Google Cloud Platform (Default)
- Cloud Run (auto-scaling containers)
- Cloud SQL (PostgreSQL)
- Cloud Storage (images)

### Option 2: Digital Ocean (Alternative)
- App Platform (managed containers)
- Managed PostgreSQL
- Spaces (S3-compatible storage)
```

**3. Dockerfile**

✅ **No changes needed** - Works on both platforms!

Current Dockerfile is perfect for Digital Ocean.

### Environment Variables

**GCP-Specific** (`.env` currently):
```bash
# No GCP-specific variables currently
# All env vars are platform-agnostic ✅
```

**Digital Ocean** (same variables work):
```bash
# All current variables work as-is
PORT=8080
DB_HOST=<managed-db-host>
DB_PASSWORD=<from-DO>
JWT_SECRET=<generated>
STRIPE_SECRET_KEY=<your-key>
```

### Storage for Uploaded Images

**Current** (Local filesystem):
```go
// backend/cmd/server/main.go:155
fileServer := http.FileServer(http.Dir(cfg.Upload.Dir))
```

**Problem**: Local storage doesn't persist across deployments

**Solution for Digital Ocean**:

**Option A: Digital Ocean Spaces** (Recommended)
```go
// Use S3-compatible storage
// Install: go get github.com/aws/aws-sdk-go/service/s3

// Configure for DO Spaces
s3Client := s3.New(session.New(&aws.Config{
    Region:      aws.String("sgp1"),
    Endpoint:    aws.String("https://sgp1.digitaloceanspaces.com"),
    Credentials: credentials.NewStaticCredentials(key, secret, ""),
}))
```

**Option B: Volume Mount** (Simpler)
```yaml
# In .do/app.yaml
services:
- name: web
  instance_size_slug: basic-xs
  volumes:
  - name: uploads
    mount_path: /app/uploads
```

**Option C: Keep Local** (Works but uploads lost on redeploy)
- Acceptable for development/testing
- Use Spaces for production

### Database Connection

**GCP Cloud SQL**:
```bash
DB_HOST=<cloud-sql-ip>
DB_SSL_MODE=require
```

**Digital Ocean Managed Database**:
```bash
DB_HOST=<managed-db-host>
DB_PORT=25060  # DO uses custom port
DB_SSL_MODE=require
```

✅ **No code changes needed** - Just update env vars

### Summary of Required Changes

| Component | Changes Needed | Complexity |
|-----------|----------------|------------|
| **Dockerfile** | ✅ None | N/A |
| **Go Code** | ✅ None | N/A |
| **Frontend** | ✅ None | N/A |
| **Environment Variables** | ⚠️ Update values only | Easy |
| **File Storage** | ⚠️ Optional: Add Spaces | Medium |
| **Makefile** | ⚠️ Add `deploy-do` target | Easy |
| **Documentation** | ⚠️ Update README | Easy |

**Total Effort**: ~30 minutes to 1 hour

---

# Part 2: Digital Ocean Deployment Options

## Option A: App Platform (Recommended)

**What is it**: Managed container platform (like Cloud Run/Heroku)

### Pros
- ✅ Auto-deploy from GitHub
- ✅ Zero server management
- ✅ Auto-scaling (within limits)
- ✅ Free SSL certificates
- ✅ Easy to use
- ✅ Built-in CI/CD

### Cons
- ❌ Less control than droplet
- ❌ Fixed pricing tiers
- ❌ Limited customization

### Pricing
- **Basic**: $5/month (512MB RAM)
- **Professional**: $12/month (1GB RAM) ← **Recommended for Zen Bali**
- **Database**: $15/month (managed PostgreSQL)
- **Total**: ~$27/month

---

## Option B: Droplet (VPS)

**What is it**: Virtual private server (full control)

### Pros
- ✅ Full root access
- ✅ Maximum flexibility
- ✅ Cheaper (can self-host DB)
- ✅ More control

### Cons
- ❌ Manual server management
- ❌ No auto-deploy (must configure)
- ❌ You handle updates/security
- ❌ More setup work

### Pricing
- **Droplet**: $12/month (2GB RAM, 50GB SSD)
- **Database**: Self-hosted (free) or Managed ($15/month)
- **Total**: $12-27/month

---

# Part 3: Deployment Steps

## Option A: Deploy to App Platform

### Prerequisites

**1. GitHub Repository**
```bash
# Create repo on GitHub
# Push your code
git remote add origin https://github.com/yourusername/zenbali.git
git push -u origin main
```

**2. Digital Ocean Account**
- Sign up at https://www.digitalocean.com/
- Get $200 free credit (new accounts)

**3. Generate Secrets**
```bash
# JWT Secret
openssl rand -base64 64

# Save securely (you'll need for DO)
```

### Step-by-Step Deployment

#### Step 1: Create App Platform App

**Via Dashboard**:

1. Go to https://cloud.digitalocean.com/apps
2. Click **"Create App"**
3. Choose **GitHub** as source
4. Authorize Digital Ocean to access GitHub
5. Select **zenbali** repository
6. Select **main** branch
7. Auto-deploy: **ON** ✅
8. Click **Next**

**Via CLI**:

Create `.do/app.yaml`:
```yaml
name: zenbali
region: sgp  # Singapore (closest to Bali)

services:
- name: web
  github:
    repo: yourusername/zenbali
    branch: main
    deploy_on_push: true

  dockerfile_path: Dockerfile

  http_port: 8080

  instance_count: 1
  instance_size_slug: basic-xs  # $12/month

  health_check:
    http_path: /api/health
    initial_delay_seconds: 10
    period_seconds: 10
    timeout_seconds: 3
    success_threshold: 1
    failure_threshold: 3

  routes:
  - path: /

  envs:
  - key: ENV
    value: production
    scope: RUN_TIME

  - key: PORT
    value: "8080"
    scope: RUN_TIME

  - key: BASE_URL
    value: ${APP_URL}
    scope: RUN_TIME

  # Database connection (populated after DB creation)
  - key: DB_HOST
    value: ${db.HOSTNAME}
    scope: RUN_TIME

  - key: DB_PORT
    value: ${db.PORT}
    scope: RUN_TIME

  - key: DB_USER
    value: ${db.USERNAME}
    scope: RUN_TIME

  - key: DB_PASSWORD
    value: ${db.PASSWORD}
    scope: RUN_TIME

  - key: DB_NAME
    value: ${db.DATABASE}
    scope: RUN_TIME

  - key: DB_SSL_MODE
    value: require
    scope: RUN_TIME

  # Secrets (add via dashboard)
  - key: JWT_SECRET
    scope: RUN_TIME
    type: SECRET

  - key: STRIPE_SECRET_KEY
    scope: RUN_TIME
    type: SECRET

  - key: STRIPE_PUBLISHABLE_KEY
    scope: RUN_TIME
    type: SECRET

  - key: STRIPE_WEBHOOK_SECRET
    scope: RUN_TIME
    type: SECRET

  - key: ADMIN_EMAIL
    value: admin@zenbali.org
    scope: RUN_TIME

  - key: ADMIN_PASSWORD
    scope: RUN_TIME
    type: SECRET

databases:
- name: db
  engine: PG
  version: "15"
  production: false  # Development tier ($15/month)
  cluster_name: zenbali-db
```

Deploy:
```bash
doctl apps create --spec .do/app.yaml
```

#### Step 2: Add Database

**Via Dashboard**:

1. In app settings, scroll to **"Database"**
2. Click **"Add Database"**
3. Choose **PostgreSQL 15**
4. Choose **Development** tier ($15/month)
5. Click **"Add Database"**

**Via CLI** (already in app.yaml above):
```bash
# Database is created automatically from spec
```

#### Step 3: Configure Environment Variables

**Via Dashboard**:

1. Go to app → **Settings** → **Environment Variables**
2. Add secrets (click **"Edit"**):

```bash
JWT_SECRET=<your-generated-secret>
STRIPE_SECRET_KEY=sk_live_xxxxx
STRIPE_PUBLISHABLE_KEY=pk_live_xxxxx
STRIPE_WEBHOOK_SECRET=whsec_xxxxx
ADMIN_PASSWORD=<secure-password>
```

3. Click **"Save"**

**Via CLI**:
```bash
# Update spec file with secrets
# Then update app
doctl apps update <app-id> --spec .do/app.yaml
```

#### Step 4: Deploy

**Automatic** (if auto-deploy ON):
```bash
git push origin main
# Digital Ocean automatically deploys
```

**Manual**:
```bash
doctl apps create-deployment <app-id>
```

#### Step 5: Run Database Migrations

**Connect to database**:

1. Get database connection details:
```bash
doctl databases connection <db-id> --format URI
```

2. Connect via psql:
```bash
psql "postgresql://user:pass@host:25060/db?sslmode=require"
```

3. Run migrations:
```bash
# Option 1: Via psql
\i backend/internal/database/migrations/001_init.up.sql
\i backend/internal/database/migrations/002_seed_data.up.sql

# Option 2: Using migration tool
# Add connection string to .env temporarily
DB_HOST=<managed-db-host>
DB_PORT=25060
make migrate-up
```

#### Step 6: Configure Custom Domain (Optional)

**Via Dashboard**:

1. Go to app → **Settings** → **Domains**
2. Click **"Add Domain"**
3. Enter: `zenbali.org`
4. Follow DNS configuration instructions:

```bash
# Add CNAME record in your DNS:
Type: CNAME
Name: @  (or www)
Value: <your-app>.ondigitalocean.app
```

5. Wait for SSL certificate (automatic, ~5 minutes)

**Via CLI**:
```bash
doctl apps create-domain <app-id> --domain zenbali.org
```

#### Step 7: Configure Stripe Webhook

1. Go to Stripe Dashboard → **Webhooks**
2. Click **"Add endpoint"**
3. URL: `https://zenbali.org/api/webhooks/stripe`
4. Events: `checkout.session.completed`, `checkout.session.expired`
5. Copy webhook signing secret
6. Update in DO environment variables:
```bash
STRIPE_WEBHOOK_SECRET=whsec_xxxxx
```

#### Step 8: Test Deployment

```bash
# Health check
curl https://zenbali.org/api/health

# Expected response
{"success":true,"data":{"service":"zenbali","status":"healthy"}}

# Test frontend
open https://zenbali.org
```

### Post-Deployment

**Monitor logs**:
```bash
doctl apps logs <app-id> --follow
```

**View app info**:
```bash
doctl apps get <app-id>
```

**Check deployment status**:
```bash
doctl apps list-deployments <app-id>
```

---

## Option B: Deploy to Droplet

### Step-by-Step Droplet Deployment

#### Step 1: Create Droplet

```bash
doctl compute droplet create zenbali \
  --size s-2vcpu-4gb \
  --image ubuntu-22-04-x64 \
  --region sgp1 \
  --ssh-keys $(doctl compute ssh-key list --format ID --no-header)
```

Or via dashboard:
1. Create → Droplets
2. Choose **Ubuntu 22.04**
3. **2GB RAM / 2 vCPUs** ($12/month)
4. Region: **Singapore**
5. Add SSH key
6. Create

#### Step 2: Initial Server Setup

```bash
# SSH into droplet
ssh root@<droplet-ip>

# Update system
apt update && apt upgrade -y

# Install Docker
apt install docker.io docker-compose -y
systemctl enable docker
systemctl start docker

# Create app user
adduser zenbali
usermod -aG docker zenbali
su - zenbali
```

#### Step 3: Clone Repository

```bash
cd /home/zenbali
git clone https://github.com/yourusername/zenbali.git
cd zenbali
```

#### Step 4: Configure Environment

```bash
cp .env.example .env
vim .env

# Update for production:
ENV=production
PORT=8080
BASE_URL=https://zenbali.org

# Database (use managed or local)
DB_HOST=localhost  # or managed DB host
DB_PORT=5432
DB_USER=zenbali
DB_PASSWORD=<secure-password>
DB_NAME=zenbali

# Secrets
JWT_SECRET=<generated-secret>
STRIPE_SECRET_KEY=sk_live_xxx
```

#### Step 5: Start with Docker Compose

```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Check status
docker-compose ps
```

#### Step 6: Setup Nginx Reverse Proxy

```bash
# Install Nginx
apt install nginx -y

# Configure
vim /etc/nginx/sites-available/zenbali
```

Add configuration:
```nginx
server {
    listen 80;
    server_name zenbali.org www.zenbali.org;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Enable site:
```bash
ln -s /etc/nginx/sites-available/zenbali /etc/nginx/sites-enabled/
nginx -t
systemctl restart nginx
```

#### Step 7: Add SSL with Let's Encrypt

```bash
# Install Certbot
apt install certbot python3-certbot-nginx -y

# Get certificate
certbot --nginx -d zenbali.org -d www.zenbali.org

# Auto-renewal (already configured)
systemctl status certbot.timer
```

#### Step 8: Setup Auto-Deploy (Optional)

Create webhook endpoint:
```bash
# Install webhook
apt install webhook -y

# Create hook script
vim /home/zenbali/deploy.sh
```

Add script:
```bash
#!/bin/bash
cd /home/zenbali/zenbali
git pull origin main
docker-compose down
docker-compose up -d --build
```

Make executable:
```bash
chmod +x /home/zenbali/deploy.sh
```

Configure GitHub webhook to call your server.

---

# Part 4: GCP vs Digital Ocean Comparison

## Feature Comparison

| Feature | Google Cloud Platform | Digital Ocean |
|---------|----------------------|---------------|
| **Compute** | Cloud Run (serverless) | App Platform (managed) or Droplet (VPS) |
| **Auto-Scaling** | ✅ Automatic | ⚠️ Limited (App Platform) / ❌ Manual (Droplet) |
| **Database** | Cloud SQL (fully managed) | Managed PostgreSQL |
| **Storage** | Cloud Storage (global CDN) | Spaces (S3-compatible) |
| **Deployment** | `gcloud` CLI / Cloud Build | `doctl` CLI / GitHub integration |
| **Auto-Deploy** | Requires setup | ✅ Built-in (App Platform) |
| **Learning Curve** | Steep | Gentle |
| **Documentation** | Excellent | Good |
| **Region** | Asia-Southeast1 (Singapore) | SGP1 (Singapore) |

## Cost Comparison (Monthly)

### Scenario 1: Small App (100-500 visitors/day)

**Google Cloud Platform**:
```
Cloud Run:        $5-10
Cloud SQL:        $10-15
Cloud Storage:    $2-3
Bandwidth:        $5
Secret Manager:   $1
─────────────────────
Total:            $25-35/month
```

**Digital Ocean App Platform**:
```
App (Basic):      $12
Database:         $15
Spaces:           $5
─────────────────────
Total:            $32/month
```

**Digital Ocean Droplet** (Self-Managed):
```
Droplet (2GB):    $12
Database:         $0 (self-hosted)
Backup:           $1
─────────────────────
Total:            $13/month
```

**Winner**: Droplet ($13) < GCP ($25) < App Platform ($32)

### Scenario 2: Medium App (5000+ visitors/day)

**Google Cloud Platform**:
```
Cloud Run:        $20-40 (auto-scales)
Cloud SQL:        $25 (larger tier)
Cloud Storage:    $10
Bandwidth:        $15
─────────────────────
Total:            $70-90/month
```

**Digital Ocean App Platform**:
```
App (Pro):        $24 (manual scale)
Database:         $40 (larger tier)
Spaces:           $10
Load Balancer:    $10
─────────────────────
Total:            $84/month
```

**Digital Ocean Droplet**:
```
Droplet (4GB):    $24
Droplet (DB):     $24
Load Balancer:    $10
Backup:           $2
─────────────────────
Total:            $60/month
```

**Winner**: Droplet ($60) < GCP ($70) < App Platform ($84)

## When to Choose Digital Ocean

✅ **Choose Digital Ocean if:**
- Budget-conscious (20-40% cheaper for small apps)
- Want simpler deployment process
- Need quick setup (App Platform = 10 minutes)
- Comfortable with limited auto-scaling
- Prefer straightforward pricing
- Want full control (Droplet option)

## When to Choose GCP

✅ **Choose GCP if:**
- Need true serverless auto-scaling
- Expect unpredictable traffic spikes
- Want global CDN built-in
- Need enterprise features (compliance, SLAs)
- Integrating with other GCP services
- Budget > $50/month is acceptable

## Recommendation for Zen Bali

**Phase 1** (MVP, First 6 months):
→ **Digital Ocean App Platform** ($27/month)
- Simple setup
- Auto-deploy from GitHub
- Good enough performance
- Easy to manage

**Phase 2** (Growing, 6-12 months):
→ **Stay on DO or Move to GCP**
- If traffic is steady: Stay on DO Droplet ($12-24/month)
- If traffic is spiky: Move to GCP Cloud Run ($40-80/month)

**Phase 3** (Scale, 12+ months):
→ **GCP for auto-scaling**
- Unpredictable traffic
- Global expansion
- Enterprise features needed

---

# Part 5: Managing Digital Ocean via CLI

## Installation

```bash
# macOS
brew install doctl

# Linux
cd ~
wget https://github.com/digitalocean/doctl/releases/download/v1.104.0/doctl-1.104.0-linux-amd64.tar.gz
tar xf ~/doctl-1.104.0-linux-amd64.tar.gz
sudo mv ~/doctl /usr/local/bin

# Verify
doctl version
```

## Authentication

```bash
# Get API token from DO dashboard
# Settings → API → Generate New Token

# Authenticate
doctl auth init

# Paste token when prompted
# Saved to ~/.config/doctl/config.yaml

# Test authentication
doctl account get
```

## App Platform Commands

### List Apps
```bash
doctl apps list
```

### Get App Details
```bash
doctl apps get <app-id>
```

### View Logs (Real-time)
```bash
# All logs
doctl apps logs <app-id> --follow

# Runtime logs only
doctl apps logs <app-id> --follow --type run

# Build logs
doctl apps logs <app-id> --follow --type build
```

### Create Deployment
```bash
doctl apps create-deployment <app-id>
```

### List Deployments
```bash
doctl apps list-deployments <app-id>
```

### Get Deployment Status
```bash
doctl apps get-deployment <app-id> <deployment-id>
```

### Update App
```bash
doctl apps update <app-id> --spec .do/app.yaml
```

### Delete App
```bash
doctl apps delete <app-id>
```

## Database Commands

### List Databases
```bash
doctl databases list
```

### Get Database Info
```bash
doctl databases get <db-id>
```

### Get Connection Details
```bash
doctl databases connection <db-id>

# Get connection string
doctl databases connection <db-id> --format URI
```

### Create Backup
```bash
doctl databases backups create <db-id>
```

### List Backups
```bash
doctl databases backups list <db-id>
```

### Resize Database
```bash
doctl databases resize <db-id> --size db-s-2vcpu-4gb
```

## Droplet Commands

### List Droplets
```bash
doctl compute droplet list
```

### Create Droplet
```bash
doctl compute droplet create zenbali \
  --size s-2vcpu-4gb \
  --image ubuntu-22-04-x64 \
  --region sgp1 \
  --ssh-keys <key-id>
```

### SSH into Droplet
```bash
doctl compute ssh <droplet-id>
```

### Resize Droplet
```bash
doctl compute droplet-action resize <droplet-id> --size s-4vcpu-8gb
```

### Delete Droplet
```bash
doctl compute droplet delete <droplet-id>
```

## Domain Commands

### List Domains
```bash
doctl compute domain list
```

### Add Domain
```bash
doctl compute domain create zenbali.org
```

### Add DNS Record
```bash
doctl compute domain records create zenbali.org \
  --record-type A \
  --record-name @ \
  --record-data <droplet-ip>
```

### Add CNAME
```bash
doctl compute domain records create zenbali.org \
  --record-type CNAME \
  --record-name www \
  --record-data @
```

## Useful Aliases

Add to `~/.zshrc` or `~/.bashrc`:

```bash
# Digital Ocean
alias do='doctl'
alias doa='doctl apps'
alias dod='doctl databases'
alias dodrop='doctl compute droplet'

# Zen Bali specific (update with your app-id)
export ZENBALI_APP_ID="your-app-id-here"
export ZENBALI_DB_ID="your-db-id-here"

alias zenlogs='doctl apps logs $ZENBALI_APP_ID --follow'
alias zendeploy='doctl apps create-deployment $ZENBALI_APP_ID'
alias zeninfo='doctl apps get $ZENBALI_APP_ID'
alias zendb='doctl databases connection $ZENBALI_DB_ID'
alias zenstatus='doctl apps list-deployments $ZENBALI_APP_ID'
```

Then use:
```bash
zenlogs      # View real-time logs
zendeploy    # Trigger deployment
zeninfo      # Get app info
zendb        # Get DB connection
zenstatus    # Check deployment status
```

## Complete CLI Workflow Example

```bash
# 1. Check app status
doctl apps get <app-id>

# 2. View recent deployments
doctl apps list-deployments <app-id> | head -5

# 3. Watch logs
doctl apps logs <app-id> --follow

# 4. Manual deploy
doctl apps create-deployment <app-id>

# 5. Check deployment status
doctl apps list-deployments <app-id> | head -1

# 6. Connect to database
DB_URI=$(doctl databases connection <db-id> --format URI | tail -1)
psql "$DB_URI"

# 7. View app URL
doctl apps get <app-id> --format LiveURL --no-header
```

## Automation Scripts

**Auto-deploy script** (`scripts/deploy-do.sh`):
```bash
#!/bin/bash
set -e

APP_ID="your-app-id"

echo "🚀 Deploying Zen Bali to Digital Ocean..."

# Create deployment
DEPLOYMENT_ID=$(doctl apps create-deployment $APP_ID --format ID --no-header)

echo "📦 Deployment ID: $DEPLOYMENT_ID"
echo "⏳ Watching deployment..."

# Watch logs
doctl apps logs $APP_ID --follow --type deploy &
LOG_PID=$!

# Wait for deployment to complete
while true; do
  STATUS=$(doctl apps get-deployment $APP_ID $DEPLOYMENT_ID --format Phase --no-header)

  if [ "$STATUS" = "ACTIVE" ]; then
    echo "✅ Deployment successful!"
    kill $LOG_PID 2>/dev/null || true
    break
  elif [ "$STATUS" = "ERROR" ]; then
    echo "❌ Deployment failed!"
    kill $LOG_PID 2>/dev/null || true
    exit 1
  fi

  sleep 5
done

# Get app URL
URL=$(doctl apps get $APP_ID --format LiveURL --no-header)
echo "🌐 App is live at: $URL"
```

Make executable:
```bash
chmod +x scripts/deploy-do.sh
```

Use:
```bash
./scripts/deploy-do.sh
```

---

# Part 6: Migration Path

## From Development to Digital Ocean

**Current**: Local development
**Target**: Digital Ocean App Platform

**Steps**:
1. ✅ Push code to GitHub
2. ✅ Create DO account
3. ✅ Create App Platform app
4. ✅ Add managed database
5. ✅ Configure environment variables
6. ✅ Deploy and test
7. ✅ Add custom domain
8. ✅ Configure Stripe webhook

**Time**: ~2-3 hours (first time)

## From Digital Ocean to GCP (Future)

If you need to migrate to GCP later:

**Steps**:
1. Export database from DO
2. Import to Cloud SQL
3. Update environment variables
4. Deploy to Cloud Run
5. Update DNS
6. Test thoroughly
7. Decommission DO resources

**Time**: ~4-6 hours

---

# Part 7: Troubleshooting

## Common Issues

**Build Fails**:
```bash
# View build logs
doctl apps logs <app-id> --type build

# Common causes:
# - Missing dependencies in Dockerfile
# - Wrong Dockerfile path
# - Build timeout (increase in settings)
```

**App Crashes on Start**:
```bash
# View runtime logs
doctl apps logs <app-id> --type run --follow

# Common causes:
# - Missing environment variables
# - Database connection failed
# - Port mismatch (ensure PORT=8080)
```

**Database Connection Issues**:
```bash
# Test connection
doctl databases connection <db-id>

# Check firewall rules
# DO manages automatically for App Platform
# For droplet: Add droplet IP to trusted sources
```

**Deployment Stuck**:
```bash
# Check deployment status
doctl apps get-deployment <app-id> <deployment-id>

# Cancel and retry
doctl apps cancel-deployment <app-id> <deployment-id>
doctl apps create-deployment <app-id>
```

---

# Summary

## Quick Decision Matrix

| Requirement | Recommended Option |
|-------------|-------------------|
| **Easiest setup** | DO App Platform |
| **Cheapest** | DO Droplet (self-managed) |
| **Auto-scaling** | GCP Cloud Run |
| **Full control** | DO Droplet |
| **Best for learning** | DO App Platform |
| **Production-ready now** | Either (both work well) |

## Final Recommendation

**Start with**: Google Cloud Platform (your current choice) ✅

**Why stick with GCP**:
- Already configured
- Better auto-scaling for growth
- Superior managed services
- Your architecture diagram shows GCP
- Worth the extra $10-20/month for features

**Keep Digital Ocean as backup**:
- Document this alternative (this file)
- Can migrate in 2-3 hours if needed
- Good fallback if GCP costs get too high

---

# Additional Resources

## Digital Ocean Documentation
- **App Platform**: https://docs.digitalocean.com/products/app-platform/
- **Managed Databases**: https://docs.digitalocean.com/products/databases/
- **Spaces (Storage)**: https://docs.digitalocean.com/products/spaces/
- **doctl CLI**: https://docs.digitalocean.com/reference/doctl/

## Tutorials
- **Deploy Go App**: https://www.digitalocean.com/community/tutorials/how-to-deploy-a-go-web-application-using-nginx-on-ubuntu-20-04
- **App Platform Guide**: https://docs.digitalocean.com/products/app-platform/getting-started/
- **Database Setup**: https://docs.digitalocean.com/products/databases/postgresql/

## Support
- **Community**: https://www.digitalocean.com/community/
- **Support Tickets**: Available for paid customers
- **Status Page**: https://status.digitalocean.com/
