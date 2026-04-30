// Single source of truth for values displayed across the landing page.
// Bump TGCP_VERSION whenever a new release is tagged.

export const TGCP_VERSION = '0.5.2';

// Identity shown inside the interactive TUI mock. Matches the synthetic
// auth state the Go binary renders under `--demo`.
export const DEMO_USER = 'demo@tgcp.dev';
export const DEMO_PROJECT = 'tgcp-demo-project';
export const DEMO_REGION = 'us-central1';

// One-line release tagline shown in the hero eyebrow.
export const RELEASE_TAGLINE = `${TGCP_VERSION} — CloudSQL view fix · navigation headers · status summaries`;
