# External Integrations

**Analysis Date:** 2026-03-03

## APIs & External Services

**Google Cloud Platform (Compute):**
- Google Compute Engine (GCE) - VM instance management
  - SDK/Client: google.golang.org/api/compute/v1
  - Auth: Application Default Credentials (ADC)
  - Features: List instances across zones, start/stop instances, SSH access
  - Code: `internal/services/gce/api.go`

- Google Kubernetes Engine (GKE) - Kubernetes cluster management
  - SDK/Client: google.golang.org/api/container/v1
  - Auth: ADC with cloud-platform scope
  - Features: List clusters across locations, launch k9s tool
  - Code: `internal/services/gke/api.go`

- Cloud Run - Serverless container runtime
  - SDK/Client: google.golang.org/api/run/v1 and google.golang.org/api/cloudfunctions/v2
  - Auth: ADC
  - Features: List Cloud Run services and Cloud Functions across locations
  - Code: `internal/services/cloudrun/api.go`

**Google Cloud Platform (Data):**
- BigQuery - Data warehouse and analytics
  - SDK/Client: cloud.google.com/go/bigquery and google.golang.org/api
  - Auth: ADC
  - Features: List datasets, browse table schemas
  - Code: `internal/services/bigquery/api.go`

- Cloud SQL - Managed relational databases
  - SDK/Client: google.golang.org/api/sqladmin/v1beta4
  - Auth: ADC with sql-service-admin scope
  - Features: List instances, start/stop instances
  - Code: `internal/services/cloudsql/api.go`

- Firestore - NoSQL document database
  - SDK/Client: google.golang.org/api/firestore/v1 and google.golang.org/api/datastore/v1
  - Auth: ADC
  - Features: List databases and collections
  - Code: `internal/services/firestore/api.go`

- Cloud Bigtable - Wide-column NoSQL database
  - SDK/Client: google.golang.org/api/bigtableadmin/v2
  - Auth: ADC
  - Features: List instances and clusters
  - Code: `internal/services/bigtable/api.go`

- Cloud Spanner - Globally distributed relational database
  - SDK/Client: google.golang.org/api/spanner/v1
  - Auth: ADC
  - Features: List instances and databases
  - Code: `internal/services/spanner/api.go`

**Google Cloud Platform (Storage & Networking):**
- Google Cloud Storage (GCS) - Object storage
  - SDK/Client: cloud.google.com/go/storage
  - Auth: ADC
  - Features: List buckets, browse objects with directory navigation
  - Code: `internal/services/gcs/api.go`

- Persistent Disks - Block storage for VMs
  - SDK/Client: google.golang.org/api/compute/v1
  - Auth: ADC
  - Features: List disks across zones
  - Code: `internal/services/disks/api.go`

- Cloud Networking (VPCs/Subnets/Firewalls) - Virtual private networks
  - SDK/Client: google.golang.org/api/compute/v1
  - Auth: ADC
  - Features: List networks, subnets, firewall rules
  - Code: `internal/services/net/api.go`

**Google Cloud Platform (Messaging & Events):**
- Cloud Pub/Sub - Message queuing and event streaming
  - SDK/Client: google.golang.org/api/pubsub/v1
  - Auth: ADC
  - Features: List topics and subscriptions
  - Code: `internal/services/pubsub/api.go`

- Cloud Dataflow - Unified stream and batch processing
  - SDK/Client: google.golang.org/api/dataflow/v1b3
  - Auth: ADC
  - Features: List dataflow jobs across regions
  - Code: `internal/services/dataflow/api.go`

- Cloud Dataproc - Managed Spark and Hadoop clusters
  - SDK/Client: google.golang.org/api/dataproc/v1
  - Auth: ADC
  - Features: List dataproc clusters per region
  - Code: `internal/services/dataproc/api.go`

**Google Cloud Platform (Security & Monitoring):**
- Cloud IAM - Identity and Access Management
  - SDK/Client: google.golang.org/api/iam/v1
  - Auth: ADC with cloud-platform scope
  - Features: List service accounts and roles
  - Code: `internal/services/iam/api.go`

- Secret Manager - Secrets storage and retrieval
  - SDK/Client: google.golang.org/api/secretmanager/v1
  - Auth: ADC
  - Features: List secrets (values read-only, requires additional permissions)
  - Code: `internal/services/secrets/api.go`

- Cloud Logging - Application and infrastructure logs
  - SDK/Client: google.golang.org/api/logging/v2
  - Auth: ADC with logging.read scope
  - Features: Query and view logs for resources
  - Code: `internal/services/logging/api.go`

- Cloud Monitoring (referenced via recommendations) - Metrics and alerting
  - SDK/Client: google.golang.org/api via monitoring API (imported transitive)
  - Auth: ADC

**Google Cloud Platform (Infrastructure & Cost):**
- Cloud Billing - Cost management and budgets
  - SDK/Client: google.golang.org/api/cloudbilling/v1 and google.golang.org/api/billingbudgets/v1
  - Auth: ADC with cloud-platform scope
  - Features: List billing accounts and budgets
  - Code: `internal/services/overview/api.go`

- Recommender API - Cost and performance recommendations
  - SDK/Client: google.golang.org/api/recommender/v1
  - Auth: ADC
  - Features: Display cost optimization and sustainability recommendations
  - Code: `internal/services/overview/api.go`

**Google Cloud Platform (DevOps):**
- Cloud Build - CI/CD build service
  - SDK/Client: google.golang.org/api/cloudbuild/v1
  - Auth: ADC
  - Features: List build triggers, view build history and status
  - Code: `internal/services/cloudbuild/api.go`

- Artifact Registry - Container and package registry
  - SDK/Client: google.golang.org/api/artifactregistry/v1
  - Auth: ADC
  - Features: List repositories, browse container images and package versions
  - Code: `internal/services/artifactregistry/api.go`

**Redis:**
- Cloud Memorystore for Redis - Managed Redis cache
  - SDK/Client: google.golang.org/api/redis/v1
  - Auth: ADC
  - Features: List Redis instances across regions
  - Code: `internal/services/redis/api.go`

## Data Storage

**Databases:**
- Google Cloud SQL - Relational (MySQL, PostgreSQL, SQL Server)
  - Connection: Requires ADC with appropriate service account roles
  - Client: google.golang.org/api/sqladmin/v1beta4 (admin API, not direct connections)

- BigQuery - Columnar warehouse
  - Connection: ADC credentials
  - Client: cloud.google.com/go/bigquery

- Cloud Bigtable - Wide-column NoSQL
  - Connection: ADC credentials
  - Client: google.golang.org/api/bigtableadmin/v2

- Firestore/Datastore - Document NoSQL
  - Connection: ADC credentials
  - Clients: google.golang.org/api/firestore/v1 and google.golang.org/api/datastore/v1

- Cloud Spanner - Relational distributed
  - Connection: ADC credentials
  - Client: google.golang.org/api/spanner/v1

**File Storage:**
- Google Cloud Storage (GCS) - Object storage
  - Client: cloud.google.com/go/storage
  - Connection: ADC credentials
  - No local filesystem caching of downloaded files (read-only UI)

**Caching:**
- Cloud Memorystore for Redis - Managed Redis service (observed only, not written to)
  - Client: google.golang.org/api/redis/v1
  - Application-level caching: In-memory cache (not persistent)

## Authentication & Identity

**Auth Provider:**
- Google Cloud ADC (Application Default Credentials)
  - Implementation: oauth2/google.FindDefaultCredentials() checks multiple sources:
    1. GOOGLE_APPLICATION_CREDENTIALS environment variable (file path)
    2. ~/.config/gcloud/application_default_credentials.json (standard gcloud ADC location)
    3. Metadata server (if running on GCP)
  - Scope: https://www.googleapis.com/auth/cloud-platform (full access)
  - Project ID detection: From credentials JSON → quota_project_id fallback → gcloud config fallback
  - User email detection: From ADC JSON client_email field, gcloud config, or "Unknown"
  - Code: `internal/core/auth.go`

**Per-Service Scopes:**
- GCE/IAM/Cloud SQL: Specific scope enforcement (e.g., compute.ComputeScope, iam.CloudPlatformScope)
- Logging: logging.LoggingReadScope
- Most other services: cloud-platform scope (default)

## Monitoring & Observability

**Error Tracking:**
- Not detected - Errors logged to local debug.log only

**Logs:**
- Debug logging approach:
  - File: ~/.tgcp/debug.log (when --debug flag enabled)
  - Log function: internal/utils.Log() - Simple file-based logging
  - Runtime enablement: via --debug CLI flag
  - No external log aggregation service

**Rate Limiting & Retries:**
- Implementation: `internal/core/client.go`
- Token bucket rate limiter: 10 requests/sec, 20 burst capacity
- Retry transport: Exponential backoff, max 3 retries
  - Backoff: 100ms, 200ms, 400ms (2^i * 100ms)
  - Retries on: network errors, HTTP 429, HTTP 5xx
  - Non-retryable: HTTP 4xx (except 429), context cancellation

## CI/CD & Deployment

**Hosting:**
- GitHub Releases - Official distribution
- Homebrew (macOS) - Package management via homebrew-tgcp tap
- Direct binary download - Linux tarball distribution

**CI Pipeline:**
- GoReleaser - Automated cross-platform builds and releases
  - Triggers: Git tags (automatic via GitHub Actions, inferred from .goreleaser.yaml)
  - Builds: macOS (amd64, arm64), Linux (amd64, arm64)
  - Artifacts: tar.gz archives with README and LICENSE
  - Homebrew formula auto-generation
  - Checksums and release notes generation

**Build Prerequisites:**
- go mod tidy - Dependency management
- CGO_ENABLED=0 - Static linking for portability
- Version injection via ldflags during build

## Environment Configuration

**Required env vars:**
- GOOGLE_APPLICATION_CREDENTIALS - Optional, path to service account JSON or user credentials JSON (alternative to standard ADC path)
- GOOGLE_CLOUD_PROJECT - Optional but recommended (fallback project ID if not in ADC)

**Configuration file:**
- ~/.tgcprc - User configuration (YAML format)
  - Fields: project, region, zone, ui (sidebar_visible, refresh_interval, default_view), features (enable_gce, enable_cloudsql)

**Secrets location:**
- Secrets stored in Google Cloud Secret Manager - Viewed via `internal/services/secrets/api.go`
- ADC credentials: Standard Google Cloud credential locations (no local secrets committed)
- No hardcoded credentials or API keys in codebase

## Webhooks & Callbacks

**Incoming:**
- None detected - TGCP is read-only/observational UI

**Outgoing:**
- None detected - No webhook callbacks or event publishing to external systems
- Action confirmations: Displayed in local UI only (start/stop instances, etc.)

---

*Integration audit: 2026-03-03*
