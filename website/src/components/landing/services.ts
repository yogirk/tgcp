export type Service = {
  group: string | null;
  icon: string;
  name: string;
  desc: string;
};

// 21 implemented services + Overview, matching internal/services/* directories.
// Categories match docs/CODEMAPS (8 groups incl Overview).
export const TGCP_SERVICES: Service[] = [
  { group: null,         icon: '◉', name: 'Overview',           desc: 'Command Center' },

  { group: 'Compute',    icon: '⚙', name: 'Compute Engine',     desc: 'VMs, SSH, power' },
  { group: 'Compute',    icon: '☸', name: 'Kubernetes Engine',  desc: 'GKE · k9s launch' },
  { group: 'Compute',    icon: '▷', name: 'Cloud Run',          desc: 'Services & revisions' },

  { group: 'Storage',    icon: '▤', name: 'Cloud Storage',      desc: 'GCS buckets' },
  { group: 'Storage',    icon: '◔', name: 'Persistent Disks',   desc: 'Block storage' },

  { group: 'Databases',  icon: '⛁', name: 'Cloud SQL',          desc: 'Managed SQL' },
  { group: 'Databases',  icon: '⬡', name: 'Spanner',            desc: 'Global relational' },
  { group: 'Databases',  icon: '▦', name: 'Bigtable',           desc: 'NoSQL wide-column' },
  { group: 'Databases',  icon: '◈', name: 'Firestore',          desc: 'Document DB' },
  { group: 'Databases',  icon: '◐', name: 'Memorystore',        desc: 'Redis / in-memory' },

  { group: 'Networking', icon: '⚿', name: 'IAM',                desc: 'Identity & access' },
  { group: 'Networking', icon: '◆', name: 'Secret Manager',     desc: 'Secrets & versions' },
  { group: 'Networking', icon: '⬢', name: 'VPC',                desc: 'Virtual networks' },
  { group: 'Networking', icon: '▥', name: 'Subnets',            desc: 'Network partitions' },
  { group: 'Networking', icon: '🛡', name: 'Firewalls',          desc: 'Security rules' },

  { group: 'Analytics',  icon: '◇', name: 'BigQuery',           desc: 'Data warehouse' },
  { group: 'Analytics',  icon: '◎', name: 'Pub/Sub',            desc: 'Messaging' },
  { group: 'Analytics',  icon: '≋', name: 'Dataflow',           desc: 'Stream processing' },
  { group: 'Analytics',  icon: '≈', name: 'Dataproc',           desc: 'Spark / Hadoop' },

  { group: 'Observe',    icon: '☲', name: 'Cloud Logging',      desc: 'Log explorer' },

  { group: 'DevOps',     icon: '⟳', name: 'Cloud Build',        desc: 'CI/CD pipelines' },
  { group: 'DevOps',     icon: '◱', name: 'Artifact Registry',  desc: 'Packages & images' },
];

export type FakeResource = {
  cols: string[];
  rows: (string | number)[][];
};

export const FAKE_RESOURCES: Record<string, FakeResource> = {
  'Compute Engine': {
    cols: ['NAME', 'ZONE', 'MACHINE', 'STATUS', 'IP'],
    rows: [
      ['prod-api-01',    'us-central1-a', 'n2-standard-4', 'RUNNING', '10.0.1.14'],
      ['prod-api-02',    'us-central1-b', 'n2-standard-4', 'RUNNING', '10.0.1.15'],
      ['worker-queue-a', 'us-east1-b',    'n2-standard-2', 'RUNNING', '10.0.2.31'],
      ['worker-queue-b', 'us-east1-c',    'n2-standard-2', 'RUNNING', '10.0.2.32'],
      ['batch-nightly',  'us-central1-f', 'n2-highmem-8',  'STOPPED', '—'],
      ['dev-sandbox',    'us-west1-a',    'e2-medium',     'RUNNING', '10.0.9.4'],
    ],
  },
  'Kubernetes Engine': {
    cols: ['CLUSTER', 'LOCATION', 'NODES', 'VERSION', 'STATUS'],
    rows: [
      ['prod-cluster',  'us-central1',     '12', '1.29.4-gke', 'RUNNING'],
      ['stage-cluster', 'us-central1',     '4',  '1.29.4-gke', 'RUNNING'],
      ['eu-cluster',    'europe-west1',    '8',  '1.28.9-gke', 'RUNNING'],
      ['edge-apac',     'asia-southeast1', '3',  '1.29.4-gke', 'RUNNING'],
    ],
  },
  'Cloud Run': {
    cols: ['SERVICE', 'REGION', 'REVISION', 'TRAFFIC', 'LATENCY'],
    rows: [
      ['api-gateway',   'us-central1',  'v147', '100%', '42ms'],
      ['image-resizer', 'us-central1',  'v89',  '100%', '120ms'],
      ['webhook-sink',  'us-east1',     'v23',  '100%', '18ms'],
      ['pdf-render',    'europe-west1', 'v12',  '100%', '380ms'],
    ],
  },
  'Cloud SQL': {
    cols: ['INSTANCE', 'TIER', 'VERSION', 'REGION', 'STATUS'],
    rows: [
      ['prod-postgres', 'db-custom-4-16',   'POSTGRES_15', 'us-central1', 'RUNNABLE'],
      ['analytics-ro',  'db-custom-2-8',    'POSTGRES_15', 'us-central1', 'RUNNABLE'],
      ['legacy-mysql',  'db-n1-standard-2', 'MYSQL_8_0',   'us-east1',    'RUNNABLE'],
    ],
  },
  'Cloud Storage': {
    cols: ['BUCKET', 'CLASS', 'LOCATION', 'OBJECTS', 'SIZE'],
    rows: [
      ['tgcp-artifacts', 'STANDARD', 'US',          '18.2k', '42 GiB'],
      ['user-uploads',   'STANDARD', 'US-CENTRAL1', '1.2M',  '380 GiB'],
      ['backup-nightly', 'NEARLINE', 'US',          '920',   '2.1 TiB'],
      ['cold-archive',   'ARCHIVE',  'US',          '44k',   '18 TiB'],
    ],
  },
  'BigQuery': {
    cols: ['DATASET', 'TABLES', 'LOCATION', 'SIZE', 'LAST EDIT'],
    rows: [
      ['analytics_events',  '24', 'US', '1.4 TiB', '3 min ago'],
      ['finance_reporting', '12', 'US', '180 GiB', '2 hr ago'],
      ['ml_training',       '8',  'US', '620 GiB', '1 day ago'],
    ],
  },
  _default: {
    cols: ['NAME', 'STATUS', 'REGION', 'AGE'],
    rows: [
      ['resource-alpha', 'READY',    'us-central1',  '3d'],
      ['resource-beta',  'READY',    'us-east1',     '12d'],
      ['resource-gamma', 'UPDATING', 'europe-west1', '1h'],
    ],
  },
};
