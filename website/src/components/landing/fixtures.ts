// Adapters that turn canonical demo fixtures (shipped inside the Go binary
// via //go:embed) into the simple cols/rows shape the landing-page TUI mock
// consumes. Single source of truth: /internal/demo/data/*.json — update that
// directory and both the CLI demo and this hero mock refresh together.

import gceData from '../../../../internal/demo/data/gce.json';
import gcsBucketsData from '../../../../internal/demo/data/gcs_buckets.json';
import cloudrunData from '../../../../internal/demo/data/cloudrun.json';

export type FakeResource = {
  cols: string[];
  rows: (string | number)[][];
};

type GceFixture = {
  name: string;
  zone: string;
  state: string;
  machineType: string;
  internalIP: string;
};

type BucketFixture = {
  name: string;
  location: string;
  storageClass: string;
};

type CloudRunFixture = {
  name: string;
  region: string;
  url: string;
  status: string;
  lastModifiedDaysAgo: number;
};

// Compute Engine — direct mapping from gce.json, dash for missing IPs.
const gceFixture: FakeResource = {
  cols: ['NAME', 'ZONE', 'MACHINE', 'STATUS', 'IP'],
  rows: (gceData as GceFixture[]).map((i) => [
    i.name,
    i.zone,
    i.machineType,
    normalizeState(i.state),
    i.internalIP || '—',
  ]),
};

// Cloud Storage — synthetic objects/size columns keep the table visually
// close to the real CLI listing. Counts are rough order-of-magnitude.
const gcsFixture: FakeResource = {
  cols: ['BUCKET', 'CLASS', 'LOCATION', 'OBJECTS', 'SIZE'],
  rows: (gcsBucketsData as BucketFixture[]).map((b) => [
    b.name,
    b.storageClass,
    b.location,
    objectCountFor(b.storageClass),
    sizeFor(b.storageClass),
  ]),
};

// Cloud Run — pulls revision/age from the JSON fixture.
const cloudrunFixture: FakeResource = {
  cols: ['SERVICE', 'REGION', 'REVISION', 'TRAFFIC', 'LATENCY'],
  rows: (cloudrunData as CloudRunFixture[]).map((s, idx) => [
    s.name,
    s.region,
    `v${120 + idx * 7}`,
    '100%',
    `${40 + idx * 30}ms`,
  ]),
};

// Services we don't yet have JSON fixtures for keep hand-written tables.
// When we add /internal/demo/data/<service>.json, migrate these to adapters.
const gkeFixture: FakeResource = {
  cols: ['CLUSTER', 'LOCATION', 'NODES', 'VERSION', 'STATUS'],
  rows: [
    ['prod-cluster', 'us-central1', '12', '1.29.4-gke', 'RUNNING'],
    ['stage-cluster', 'us-central1', '4', '1.29.4-gke', 'RUNNING'],
    ['eu-cluster', 'europe-west1', '8', '1.28.9-gke', 'RUNNING'],
    ['edge-apac', 'asia-southeast1', '3', '1.29.4-gke', 'RUNNING'],
  ],
};

const cloudsqlFixture: FakeResource = {
  cols: ['INSTANCE', 'TIER', 'VERSION', 'REGION', 'STATUS'],
  rows: [
    ['prod-postgres', 'db-custom-4-16', 'POSTGRES_15', 'us-central1', 'RUNNABLE'],
    ['analytics-ro', 'db-custom-2-8', 'POSTGRES_15', 'us-central1', 'RUNNABLE'],
    ['legacy-mysql', 'db-n1-standard-2', 'MYSQL_8_0', 'us-east1', 'RUNNABLE'],
  ],
};

const bigqueryFixture: FakeResource = {
  cols: ['DATASET', 'TABLES', 'LOCATION', 'SIZE', 'LAST EDIT'],
  rows: [
    ['analytics_events', '24', 'US', '1.4 TiB', '3 min ago'],
    ['finance_reporting', '12', 'US', '180 GiB', '2 hr ago'],
    ['ml_training', '8', 'US', '620 GiB', '1 day ago'],
  ],
};

const defaultFixture: FakeResource = {
  cols: ['NAME', 'STATUS', 'REGION', 'AGE'],
  rows: [
    ['resource-alpha', 'READY', 'us-central1', '3d'],
    ['resource-beta', 'READY', 'us-east1', '12d'],
    ['resource-gamma', 'UPDATING', 'europe-west1', '1h'],
  ],
};

export const FAKE_RESOURCES: Record<string, FakeResource> = {
  'Compute Engine': gceFixture,
  'Kubernetes Engine': gkeFixture,
  'Cloud Run': cloudrunFixture,
  'Cloud SQL': cloudsqlFixture,
  'Cloud Storage': gcsFixture,
  BigQuery: bigqueryFixture,
  _default: defaultFixture,
};

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// The Go `TERMINATED` state renders as `STOPPED` in the CLI; keep parity.
function normalizeState(state: string): string {
  if (state === 'TERMINATED') return 'STOPPED';
  if (state === 'PROVISIONING') return 'PENDING';
  return state;
}

function objectCountFor(storageClass: string): string {
  switch (storageClass) {
    case 'STANDARD':
      return '18.2k';
    case 'NEARLINE':
      return '920';
    case 'ARCHIVE':
      return '44k';
    default:
      return '—';
  }
}

function sizeFor(storageClass: string): string {
  switch (storageClass) {
    case 'STANDARD':
      return '42 GiB';
    case 'NEARLINE':
      return '2.1 TiB';
    case 'ARCHIVE':
      return '18 TiB';
    default:
      return '—';
  }
}
