import { useState, type ReactNode } from 'react';
import { TGCP_SERVICES } from './services';
import { TgcpMark } from './Mark';

function Section({
  id,
  label,
  title,
  children,
  className = '',
}: {
  id: string;
  label: string;
  title: ReactNode;
  children: ReactNode;
  className?: string;
}) {
  return (
    <section id={id} className={`section ${className}`}>
      <div className="section-inner">
        <div className="section-head">
          <span className="section-label">{label}</span>
          <h2 className="section-title">{title}</h2>
        </div>
        {children}
      </div>
    </section>
  );
}

export function Features() {
  const items = [
    {
      n: '01',
      title: '21 GCP services',
      body: 'Compute, databases, storage, analytics, networking, security, observability, DevOps — one binary, one view.',
      tags: ['GCE', 'GKE', 'Run', 'SQL', 'GCS', 'BigQuery', 'Pub/Sub', '+14'],
    },
    {
      n: '02',
      title: 'Keyboard-first',
      body: 'Vim-style keybindings, fuzzy filter, command palette. Everything is one keystroke away.',
      tags: ['↑↓ / jk', '/', 'Enter', '?', 'gg', 'G'],
    },
    {
      n: '03',
      title: 'Zero config',
      body: 'Uses your existing gcloud credentials. No API keys, no setup files — brew install and go.',
      tags: ['gcloud auth', 'no YAML', 'no tokens'],
    },
  ];
  return (
    <Section
      id="features"
      label="WHY TGCP"
      title={
        <>
          Built for the terminal. <span className="dim">Built for speed.</span>
        </>
      }
    >
      <div className="feature-grid">
        {items.map((it) => (
          <div key={it.n} className="feature-card">
            <div className="feature-n">{it.n}</div>
            <h3 className="feature-title">{it.title}</h3>
            <p className="feature-body">{it.body}</p>
            <div className="feature-tags">
              {it.tags.map((t) => (
                <span key={t} className="chip">
                  {t}
                </span>
              ))}
            </div>
          </div>
        ))}
      </div>
    </Section>
  );
}

export function ServiceMatrix() {
  const groups: Record<string, typeof TGCP_SERVICES> = {};
  TGCP_SERVICES.forEach((s) => {
    const g = s.group || 'Overview';
    (groups[g] = groups[g] || []).push(s);
  });
  const order = ['Compute', 'Storage', 'Databases', 'Networking', 'Analytics', 'Observe', 'DevOps'];
  const visible = order.filter((g) => groups[g]);

  return (
    <Section
      id="services"
      label="COVERAGE"
      title={
        <>
          Every service, <span className="dim">one keystroke away.</span>
        </>
      }
    >
      <div className="matrix">
        {visible.map((g) => (
          <div key={g} className="matrix-col">
            <div className="matrix-head">
              <span className="matrix-dot" />
              <span>{g}</span>
              <span className="matrix-count">{groups[g].length}</span>
            </div>
            <ul className="matrix-list">
              {groups[g].map((s) => (
                <li key={s.name}>
                  <span className="matrix-icon">{s.icon}</span>
                  <span className="matrix-name">{s.name}</span>
                  <span className="matrix-desc">{s.desc}</span>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
    </Section>
  );
}

type Line = { p: '❯' | ' '; text: string; cls?: string };

export function Install() {
  const tabs: { id: string; label: string; sub: string; lines: Line[] }[] = [
    {
      id: 'macos',
      label: 'macOS',
      sub: 'Homebrew',
      lines: [
        { p: '❯', text: 'brew tap yogirk/tgcp' },
        { p: '❯', text: 'brew install tgcp' },
        { p: ' ', text: '🍺  tgcp 0.4.1 installed to /opt/homebrew/bin/tgcp', cls: 'ok' },
      ],
    },
    {
      id: 'linux',
      label: 'Linux',
      sub: 'curl',
      lines: [
        { p: '❯', text: 'curl -L https://github.com/yogirk/tgcp/releases/latest/\\' },
        { p: ' ', text: '    download/tgcp_0.4.1_linux_amd64.tar.gz | tar -xz' },
        { p: '❯', text: 'sudo mv tgcp /usr/local/bin/' },
        { p: ' ', text: '✓  tgcp ready', cls: 'ok' },
      ],
    },
    {
      id: 'go',
      label: 'Go install',
      sub: 'go 1.21+',
      lines: [
        { p: '❯', text: 'go install github.com/yogirk/tgcp/cmd/tgcp@latest' },
        { p: '❯', text: 'tgcp --version' },
        { p: ' ', text: 'tgcp 0.4.1', cls: 'ok' },
      ],
    },
    {
      id: 'source',
      label: 'From source',
      sub: 'go build',
      lines: [
        { p: '❯', text: 'git clone https://github.com/yogirk/tgcp.git' },
        { p: '❯', text: 'cd tgcp && go build -o tgcp ./cmd/tgcp' },
        { p: '❯', text: 'sudo mv tgcp /usr/local/bin/' },
      ],
    },
  ];
  const [active, setActive] = useState('macos');
  const [copied, setCopied] = useState(false);
  const current = tabs.find((t) => t.id === active) || tabs[0];

  const copy = () => {
    const text = current.lines
      .filter((l) => l.p === '❯')
      .map((l) => l.text)
      .join('\n');
    navigator.clipboard
      ?.writeText(text)
      .then(() => {
        setCopied(true);
        setTimeout(() => setCopied(false), 1600);
      })
      .catch(() => {});
  };

  return (
    <Section
      id="install"
      label="INSTALL"
      title={
        <>
          Running in <span className="dim">under ten seconds.</span>
        </>
      }
    >
      <div className="install-wrap">
        <div className="install-tabs">
          {tabs.map((t) => (
            <button
              key={t.id}
              className={`install-tab${active === t.id ? ' is-active' : ''}`}
              onClick={() => setActive(t.id)}
            >
              <span className="install-tab-label">{t.label}</span>
              <span className="install-tab-sub">{t.sub}</span>
            </button>
          ))}
        </div>
        <div className="install-window">
          <div className="tui-titlebar">
            <span className="tui-dots">
              <i style={{ background: '#ff5f56' }} />
              <i style={{ background: '#ffbd2e' }} />
              <i style={{ background: '#27c93f' }} />
            </span>
            <span className="tui-title">terminal — {current.sub}</span>
            <button className="copy-btn" onClick={copy}>
              {copied ? '✓ copied' : 'copy'}
            </button>
          </div>
          <div className="install-body">
            {current.lines.map((l, i) => (
              <div
                key={i}
                className={`typed-line ${l.p === '❯' ? 'is-cmd' : 'is-out'} ${l.cls || ''}`}
              >
                {l.p === '❯' ? (
                  <span className="typed-prompt">❯</span>
                ) : (
                  <span className="typed-prompt is-blank">{l.p}</span>
                )}
                <span>{l.text}</span>
              </div>
            ))}
          </div>
        </div>
        <p className="install-note">
          Requires <span className="mono">gcloud</span> CLI authenticated with your account. tgcp reads
          Application Default Credentials from your active config — no extra setup.
        </p>
      </div>
    </Section>
  );
}

export function Footer() {
  return (
    <footer className="footer">
      <div className="footer-inner">
        <div className="footer-brand">
          <TgcpMark />
          <span>tgcp</span>
        </div>
        <div className="footer-links">
          <a href="#features">Features</a>
          <a href="#services">Services</a>
          <a href="#install">Install</a>
          <a href="https://github.com/yogirk/tgcp" target="_blank" rel="noreferrer">
            GitHub ↗
          </a>
        </div>
        <div className="footer-meta">
          <span>MIT License</span>
          <span className="sep">·</span>
          <span>
            Built by{' '}
            <a href="https://github.com/yogirk" target="_blank" rel="noreferrer">
              yogirk
            </a>
          </span>
          <span className="sep">·</span>
          <span>Bubble Tea · Lipgloss</span>
        </div>
      </div>
    </footer>
  );
}
