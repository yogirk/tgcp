# TGCP Features

TGCP (Terminal GCP Explorer) provides a comprehensive terminal interface for managing Google Cloud Platform resources. Below is a detailed breakdown of the supported features for each service.

## Core Capabilities

-   **Fast Navigation**: Global keyboard shortcuts (`Shift+Left/Right` to switch tabs/services, `j/k` for list navigation).
-   **Mouse Support**: Click to select items in the sidebar and service lists. Hold Shift while dragging to select text for copying.
-   **Service Sidebar**: Quick access to all supported GCP services.
-   **Command Palette**: Access any resource or command instantly with `:`.
-   **Smart Caching**: Minimizes API calls for a responsive experience.
-   **ADC Authentication**: Seamless integration with your existing `gcloud` credentials.
-   **Version Updates**: Automatic update checking with notifications when new versions are available.

## Supported Services

### Compute Engine (GCE)
-   **List Instances**: View all instances across zones.
-   **Instance Details**: Deep dive into instance metadata, IPs, machine types, and status.
-   **Power Management**: Start and Stop instances safely.
-   **Smart SSH**: SSH into instances directly. If using Tmux, opens a new pane automatically.

### Cloud SQL
-   **Instance Monitoring**: View database instances, versions, and states.
-   **Health Checks**: Quickly verify if your database is runnable or stopped.

### Kubernetes Engine (GKE)
-   **Cluster Overview**: View clusters and node pools.
-   **K9s Integration**: Launch `k9s` for a selected cluster instantly with the `K` key.

### Cloud Run
-   **Service Listing**: View deployed Cloud Run services.
-   **Revision Management**: Inspect revisions (drill-down support).
-   **Cloud Functions**: Dedicated tab for viewing Cloud Functions.

### Cloud Storage (GCS)
-   **Bucket Browser**: List all storage buckets.
-   **Object Navigation**: Navigate through folders and files within buckets.

### Identity & Access Management (IAM)
-   **Service Accounts**: Audit service accounts and their keys.
-   **Policy Review**: View IAM policies (read-only in MVP).

### Secret Manager
-   **Secret Listing**: View all secrets in your project.
-   **Version Management**: Browse secret versions and metadata.
-   **Safe by Default**: Secret values are never displayed to prevent accidental exposure.

### Networking
-   **VPC Networks**: List Virtual Private Clouds.
-   **Subnets**: View subnetworks and their IP ranges.

### Data & Analytics
-   **BigQuery**: Browse datasets and tables.
-   **Pub/Sub**: Monitor topics and subscriptions. View backlog statistics.
-   **Dataflow**: Monitor stream and batch jobs.
-   **Dataproc**: View Spark/Hadoop cluster status.
-   **Spanner**: View Spanner instances and configurations.
-   **Bigtable**: View Bigtable instances.
-   **Firestore**: View Firestore databases.
-   **Memorystore (Redis)**: View Redis instances.

### Storage
-   **Disks**: Manage Persistent Disks.
-   **Orphaned Disk Detection**: Easily identify unused disks to save costs.

### Observability
-   **Cloud Logging**: Browse and filter logs across your project.
-   **Log Entries**: View structured log entries with severity levels.
-   **Resource Filtering**: Filter logs by resource type and labels.

## Future Roadmap

-   **Billing Integration**: View daily cost estimates.
-   **Live Log Tailing**: Real-time streaming of logs for resources.
-   **Resource Graph**: Visualize dependencies between resources.
