# TGCP Features Guide

TGCP (Terminal GCP Explorer) is designed to be the fastest way to navigate and observe your Google Cloud infrastructure. this guide covers the capabilities of all 17 supported services.

## 🚀 Core Features

### ⚡ Command Palette
Press `:` to open the Command Palette. This is the **fastest way to jump anywhere**.
- Type `gke` to jump to Clusters.
- Type `switch` to change GCP projects instantly.
- Type `help` to see shortcuts.
- **New**: Jump to *any* service (e.g., `pubsub`, `redis`, `spanner`) directly.

### 🔍 Unified Search / Filter
Press `/` in any list view to filter resources by name, ID, or region. This uses fuzzy matching for speed.

### 🔄 Smart Caching
TGCP caches list results for 30-120 seconds to make navigation snappy. Press `r` to force a refresh from the API.

---

## 🛠️ Compute Services

### Compute Engine (GCE)
- **View**: List instances with Status, Zone, and Machine Type.
- **Actions**:
  - `s`: Start Instance.
  - `x`: Stop Instance.
  - `h`: **Smart SSH**. Splits your terminal window (if using tmux/iterm) or opens a shell to the instance.

### Kubernetes Engine (GKE)
- **View**: List Clusters (locations, versions) and Node Pools (machine types, autoscaling).
- **Pro Feature**: Press `K` on a cluster to instantly launch **k9s** pre-configured for that cluster context.

### Cloud Run & Functions
- **View**: Unified view for Cloud Run Services and Cloud Functions (Gen 1 & 2).
- **Details**: See latest revision status, traffic splits, and URLs.

### Disks (Block Storage)
- **View**: List Persistent Disks.
- **Orphan Detection**: TGCP automatically highlights **Unattached** disks so you can identify wasted cost.
- **Actions**: Create Snapshot (coming soon).

---

## 💾 Data & Storage

### Cloud SQL
- **View**: Instances (MySQL, Postgres, SQL Server) with Version and State.
- **Actions**: Start/Stop instance (useful for dev environments).

### Cloud Storage (GCS)
- **View**: List Buckets (Location, Storage Class).
- **Object Browser**: Press `Enter` on a bucket to browse files and folders interactively.

### Spanner
- **View**: Instances with **Node Count** and **Processing Units**.
- **Focus**: Quickly see capacity allocation across regions.

### Bigtable
- **View**: Instances and Clusters.
- **Details**: See node counts and storage type (SSD/HDD).

### Firestore
- **View**: Native and Datastore-mode databases.

### Memorystore (Redis)
- **View**: Redis instances with Tier (Basic/Standard) and Capacity.
- **Details**: Connection info (Host/Port) and Authorized Network.

---

## 📡 Messaging & Analytics

### Pub/Sub
- **View**: Top-level view of **Topics**.
- **Subscriptions**: Press `s` on a topic to see its Subscriptions.
- **Health**: Monitor **Unacked Message Count** and **Backlog Age** to spot stuck consumers.

### Dataflow
- **View**: List Jobs with State (Running/Failed) and Type (Batch/Streaming).
- **Focus**: Quickly identify failed pipelines.

### Dataproc
- **View**: Hadoop/Spark Clusters.
- **Details**: Master and Worker machine types and counts.

### BigQuery
- **View**: Datasets and Tables.
- **Drill-down**: Press `Enter` to explore tables and view schema definitions.

---

## 🛡️ Management & Networking

### VPC Network
- **View**: VPC Networks and Subnets (CIDR ranges).

### IAM
- **View**: List Service Accounts.
- **Details**: View email and disabled status.

---

## 🎹 Keyboard Cheat Sheet

| Key | Global Action |
| :--- | :--- |
| `:` | **Command Palette** (Jump to anywhere) |
| `/` | Filter current list |
| `r` | Refresh data |
| `?` | Toggle Help |
| `Tab` | Toggle Sidebar |
| `q` | Go Back / Quit |

| Key | Service Action |
| :--- | :--- |
| `Enter` | View Details / Audit / Drill-down |
| `s` | Start Resource / Switch View |
| `x` | Stop Resource |
| `h` | SSH Connect |
| `K` | Launch k9s |
