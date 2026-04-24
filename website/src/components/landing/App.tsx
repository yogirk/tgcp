import { useEffect, useState } from 'react';
import { TGCPTui } from './Tui';
import { Features, ServiceMatrix, Install, Footer } from './Sections';
import { TgcpMark } from './Mark';
import { TGCP_VERSION, RELEASE_TAGLINE } from './constants';

type Theme = 'dark' | 'light';

export function App() {
  const [theme, setTheme] = useState<Theme>('dark');
  const [heroCopied, setHeroCopied] = useState(false);

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
  }, [theme]);

  const heroCmd = 'brew install yogirk/tgcp/tgcp';
  const copyHero = () => {
    navigator.clipboard
      ?.writeText(heroCmd)
      .then(() => {
        setHeroCopied(true);
        setTimeout(() => setHeroCopied(false), 1400);
      })
      .catch(() => {});
  };

  return (
    <>
      <nav className="nav">
        <div className="nav-inner">
          <div className="nav-brand">
            <TgcpMark />
            tgcp
            <span className="nav-version">v{TGCP_VERSION}</span>
          </div>
          <div className="nav-links">
            <a href="#features">Features</a>
            <a href="#services">Services</a>
            <a href="#install">Install</a>
          </div>
          <div className="nav-actions">
            <button
              className="theme-btn"
              onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
              aria-label="Toggle theme"
              title="Toggle theme"
            >
              {theme === 'dark' ? '☾ dark' : '☀ light'}
            </button>
            <a
              className="gh-btn"
              href="https://github.com/yogirk/tgcp"
              target="_blank"
              rel="noreferrer"
            >
              <svg viewBox="0 0 16 16" width="13" height="13" aria-hidden="true" fill="currentColor">
                <path
                  fillRule="evenodd"
                  d="M8 0C3.58 0 0 3.58 0 8a8.02 8.02 0 005.47 7.59c.4.07.55-.17.55-.38 0-.19-.01-.69-.01-1.36-2.23.49-2.7-1.07-2.7-1.07-.36-.93-.89-1.17-.89-1.17-.73-.5.05-.49.05-.49.8.06 1.22.82 1.22.82.72 1.23 1.88.88 2.34.67.07-.52.28-.87.51-1.07-1.78-.2-3.65-.89-3.65-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.13 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27s1.36.09 2 .27c1.53-1.04 2.2-.82 2.2-.82.44 1.11.16 1.93.08 2.13.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.66 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.01 8.01 0 0016 8c0-4.42-3.58-8-8-8z"
                />
              </svg>
              GitHub
            </a>
          </div>
        </div>
      </nav>

      <div className="hero">
        <div className="hero-grid-bg" />
        <div className="hero-top">
          <div>
            <div className="hero-eyebrow">
              <span className="ping" />
              <span>{RELEASE_TAGLINE}</span>
            </div>
            <h1 className="hero-title">
              Your GCP,
              <br />
              <span className="accent">a keystroke away.</span>
            </h1>
            <p className="hero-sub">
              A fast, keyboard-driven <strong>TUI</strong> for observing and managing Google Cloud —
              twenty-one services, vim bindings, zero config. Inspired by k9s.
            </p>
            <div className="hero-cta">
              <button
                className={`hero-install${heroCopied ? ' is-copied' : ''}`}
                onClick={copyHero}
                title="Copy to clipboard"
              >
                <span className="prompt">$</span>
                <span>{heroCmd}</span>
                <span className="copy-state">{heroCopied ? 'copied' : 'copy'}</span>
              </button>
              <a className="btn-ghost" href="#install">
                More install options →
              </a>
            </div>
            <div className="hero-stats">
              <div className="hero-stat">
                <span className="hero-stat-n">21</span>
                <span className="hero-stat-l">Services</span>
              </div>
              <div className="hero-stat">
                <span className="hero-stat-n">vim</span>
                <span className="hero-stat-l">Keybindings</span>
              </div>
              <div className="hero-stat">
                <span className="hero-stat-n">ADC</span>
                <span className="hero-stat-l">Auth, zero config</span>
              </div>
              <div className="hero-stat">
                <span className="hero-stat-n">MIT</span>
                <span className="hero-stat-l">Open source</span>
              </div>
            </div>
          </div>

          <div className="tui-wrap">
            <TGCPTui autoplay compact />
            <div
              style={{
                marginTop: 14,
                display: 'flex',
                gap: 10,
                alignItems: 'center',
                fontFamily: 'var(--font-mono)',
                fontSize: 12,
                color: 'var(--fg-3)',
                flexWrap: 'wrap',
              }}
            >
              <span>Try it —</span>
              <kbd>↑</kbd>
              <kbd>↓</kbd>
              <span>navigate</span>
              <kbd>/</kbd>
              <span>filter</span>
              <kbd>⏎</kbd>
              <span>open</span>
              <kbd>Esc</kbd>
              <span>back</span>
            </div>
          </div>
        </div>
      </div>

      <Features />
      <ServiceMatrix />
      <Install />
      <Footer />
    </>
  );
}
