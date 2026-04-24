export type Service = {
  group: string | null;
  icon: string;
  name: string;
  desc: string;
};

// FAKE_RESOURCES now lives in fixtures.ts, which adapts canonical JSON
// shipped inside the Go binary. Re-export so existing imports keep working.
export { FAKE_RESOURCES, type FakeResource } from './fixtures';

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

