export interface Service {
  name: string;
  description: string;
}

export interface ServiceCategory {
  name: string;
  color: string;
  services: Service[];
}

export const categories: ServiceCategory[] = [
  {
    name: 'Compute',
    color: 'blue',
    services: [
      { name: 'GCE Instances', description: 'Manage power state, SSH' },
      { name: 'GKE Clusters', description: 'Launch k9s integration' },
      { name: 'Cloud Run', description: 'Services & revisions' },
      { name: 'Cloud Functions', description: 'Serverless functions' },
    ],
  },
  {
    name: 'Data',
    color: 'red',
    services: [
      { name: 'Cloud SQL', description: 'Managed databases' },
      { name: 'BigQuery', description: 'Data warehouse' },
      { name: 'Bigtable', description: 'NoSQL wide-column' },
      { name: 'Spanner', description: 'Global relational DB' },
      { name: 'Firestore', description: 'Document database' },
      { name: 'Redis', description: 'In-memory cache' },
    ],
  },
  {
    name: 'Storage',
    color: 'yellow',
    services: [
      { name: 'GCS Buckets', description: 'Object storage browser' },
      { name: 'Persistent Disks', description: 'Block storage' },
    ],
  },
  {
    name: 'Security',
    color: 'green',
    services: [
      { name: 'IAM', description: 'Identity & access' },
      { name: 'Secret Manager', description: 'Secrets vault' },
    ],
  },
  {
    name: 'Networking',
    color: 'blue',
    services: [
      { name: 'VPCs', description: 'Virtual networks' },
      { name: 'Subnets', description: 'Network partitions' },
      { name: 'Firewalls', description: 'Security rules' },
    ],
  },
  {
    name: 'Analytics & Observability',
    color: 'yellow',
    services: [
      { name: 'Pub/Sub', description: 'Messaging & events' },
      { name: 'Dataflow', description: 'Stream processing' },
      { name: 'Dataproc', description: 'Managed Spark/Hadoop' },
      { name: 'Cloud Logging', description: 'Log explorer' },
    ],
  },
  {
    name: 'DevOps',
    color: 'red',
    services: [
      { name: 'Cloud Build', description: 'CI/CD pipelines' },
      { name: 'Artifact Registry', description: 'Container & package registry' },
    ],
  },
];
