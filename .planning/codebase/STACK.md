# Technology Stack

**Analysis Date:** 2026-03-03

## Languages

**Primary:**
- Go 1.25.5 - Core application language for terminal UI and GCP API interactions

## Runtime

**Environment:**
- Go 1.25.5 - Compiled to standalone binaries for macOS and Linux

**Build System:**
- GoReleaser - Multi-platform binary compilation and release automation
- Standard Go toolchain - Build and testing

## Frameworks

**UI Framework:**
- Bubble Tea (github.com/charmbracelet/bubbletea) v1.3.10 - Terminal UI framework for interactive applications
- Bubbles (github.com/charmbracelet/bubbles) v0.21.0 - Reusable UI components (tables, lists, filters)
- Lipgloss (github.com/charmbracelet/lipgloss) v1.1.0 - Terminal styling and layout

**GCP Client Libraries:**
- Google API Go Client (google.golang.org/api) v0.259.0 - REST API clients for all GCP services
- Cloud Storage (cloud.google.com/go/storage) v1.58.0 - GCS bucket operations
- BigQuery (cloud.google.com/go/bigquery) v1.72.0 - BigQuery dataset and table queries
- google.golang.org/oauth2 v0.34.0 - OAuth2 authentication and ADC (Application Default Credentials)

**Utilities:**
- fuzzy (github.com/sahilm/fuzzy) v0.1.1 - Fuzzy search/filtering in lists
- yaml.v3 (gopkg.in/yaml.v3) v3.0.1 - YAML configuration file parsing

## Key Dependencies

**Critical (Direct):**
- google.golang.org/api v0.259.0 - Core GCP API access for Compute, Cloud SQL, IAM, Bigtable, Firestore, Redis, Spanner, Cloud Run, GKE, Dataflow, Dataproc, Pub/Sub, Logging, Secret Manager, Billing
- cloud.google.com/go/storage v1.58.0 - Google Cloud Storage client library
- cloud.google.com/go/bigquery v1.72.0 - BigQuery client library
- github.com/charmbracelet/bubbletea v1.3.10 - Terminal UI rendering engine
- golang.org/x/oauth2 v0.34.0 - OAuth2 token handling and ADC credential discovery

**Infrastructure (Transitive):**
- google.golang.org/grpc v1.78.0 - gRPC transport for GCP API calls
- google.golang.org/protobuf v1.36.11 - Protocol Buffer encoding/decoding
- cloud.google.com/go v0.123.0 - Umbrella package for Cloud Go libraries
- cloud.google.com/go/auth v0.18.0 - GCP authentication infrastructure
- google.golang.org/genproto v0.0.0-20251202230838-ff82c1b0f217 - Generated API types
- go.opentelemetry.io/otel v1.38.0 - OpenTelemetry observability (used by GCP libs)

**Terminal/UI Supporting:**
- github.com/mattn/go-isatty v0.0.20 - Terminal capability detection
- github.com/muesli/termenv v0.16.0 - Terminal environment queries
- github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 - ANSI escape sequence handling
- github.com/atotto/clipboard v0.1.4 - System clipboard access

## Website (Landing Page)

**Framework:**
- Astro v5.17 - Static site generator (single-page landing site)
- Tailwind CSS v4.1 - Styling via `@tailwindcss/vite` plugin with inline theme

**Fonts:**
- Google Sans Flex - Headings and body text (self-hosted woff2)
- Google Sans Code - Terminal/code blocks (self-hosted woff2)

**Deployment:**
- GitHub Pages, served at `https://tgcp.yogirk.dev` (custom domain via `website/public/CNAME`, base path `/`)

## Configuration

**Environment:**
- Application Default Credentials (ADC) - Uses `gcloud auth application-default login` for authentication
- Location: `~/.config/gcloud/application_default_credentials.json` (standard ADC path)
- GOOGLE_APPLICATION_CREDENTIALS - Environment variable override for credential file path (checked if ADC not found)

**Configuration File:**
- `~/.tgcprc` - YAML format configuration file in user home directory
- Configuration fields: project ID, region, zone, UI settings, feature flags
- Parsing: gopkg.in/yaml.v3

**Build Configuration:**
- `.goreleaser.yaml` - Release automation for macOS (amd64, arm64) and Linux (amd64, arm64)
- CGO disabled for static binary linking
- Version injection via ldflags: main.version, main.commit, main.date

## Platform Requirements

**Development:**
- Go 1.21 or higher (1.25.5 used in production)
- Google Cloud SDK (gcloud CLI) - Required for authentication setup
- Supported: macOS, Linux

**Production:**
- Deployment target: macOS (Intel/Apple Silicon), Linux (x86_64/ARM64)
- Distribution: Homebrew (macOS), GitHub Releases (tarball)
- Single binary deployment - No runtime dependencies beyond system libraries

**Optional:**
- k9s - Kubernetes CLI (launched from GKE clusters view, not bundled)
- SSH client - For SSH actions on GCE instances

---

*Stack analysis: 2026-03-03*
