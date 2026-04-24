import { useEffect, useMemo, useRef, useState } from 'react';
import { TGCP_SERVICES, FAKE_RESOURCES, type Service } from './services';
import { TGCP_VERSION, DEMO_USER, DEMO_PROJECT, DEMO_REGION } from './constants';

type ListItem =
  | { type: 'header'; label: string; _i: string }
  | ({ type: 'item'; _i: number } & Service);

type Props = {
  autoplay?: boolean;
  compact?: boolean;
};

export function TGCPTui({ autoplay = false, compact = false }: Props) {
  const items: ListItem[] = useMemo(() => {
    const out: ListItem[] = [];
    let lastGroup: string | null | undefined = undefined;
    TGCP_SERVICES.forEach((s, i) => {
      if (s.group !== lastGroup) {
        out.push({ type: 'header', label: s.group || 'Overview', _i: `h${i}` });
        lastGroup = s.group;
      }
      out.push({ type: 'item', _i: i, ...s });
    });
    return out;
  }, []);

  const selectableIndices = useMemo(
    () =>
      items
        .map((it, i) => (it.type === 'item' ? i : null))
        .filter((x): x is number => x !== null),
    [items],
  );

  const [sel, setSel] = useState<number>(selectableIndices[0] ?? 0);
  const [filter, setFilter] = useState('');
  const [filterMode, setFilterMode] = useState(false);
  const [view, setView] = useState<'list' | 'detail'>('list');
  const [activeService, setActiveService] = useState<Service | null>(null);
  const [focused, setFocused] = useState(false);
  const rootRef = useRef<HTMLDivElement | null>(null);

  const visibleItems = useMemo<ListItem[]>(() => {
    if (!filter.trim()) return items;
    const f = filter.toLowerCase();
    return items.filter((it) => {
      if (it.type === 'header') {
        return items.some(
          (x) => x.type === 'item' && x.group === it.label && x.name.toLowerCase().includes(f),
        );
      }
      return (
        it.name.toLowerCase().includes(f) ||
        (it.desc || '').toLowerCase().includes(f)
      );
    });
  }, [items, filter]);

  const visibleSelectable = useMemo(
    () =>
      visibleItems
        .map((it, i) => (it.type === 'item' ? i : null))
        .filter((x): x is number => x !== null),
    [visibleItems],
  );

  useEffect(() => {
    if (visibleSelectable.length === 0) return;
    if (!visibleSelectable.includes(sel)) {
      setSel(visibleSelectable[0]);
    }
  }, [visibleSelectable, sel]);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (!focused) return;
      const t = e.target as HTMLElement | null;
      if (t && t.tagName === 'INPUT' && !t.closest('[data-tgcp-tui]')) return;

      if (view === 'detail') {
        if (
          e.key === 'Escape' ||
          e.key === 'Backspace' ||
          e.key === 'h' ||
          e.key === 'ArrowLeft'
        ) {
          e.preventDefault();
          setView('list');
        }
        return;
      }

      if (filterMode) {
        if (e.key === 'Escape') {
          e.preventDefault();
          setFilter('');
          setFilterMode(false);
          return;
        }
        if (e.key === 'Enter') {
          e.preventDefault();
          setFilterMode(false);
          return;
        }
        if (e.key === 'Backspace') {
          e.preventDefault();
          setFilter((f) => f.slice(0, -1));
          return;
        }
        if (e.key.length === 1) {
          e.preventDefault();
          setFilter((f) => f + e.key);
          return;
        }
      }

      if (e.key === '/') {
        e.preventDefault();
        setFilterMode(true);
        return;
      }
      if (e.key === 'Escape') {
        e.preventDefault();
        setFilter('');
        return;
      }

      if (e.key === 'ArrowDown' || e.key === 'j') {
        e.preventDefault();
        const idx = visibleSelectable.indexOf(sel);
        const next = visibleSelectable[Math.min(visibleSelectable.length - 1, idx + 1)];
        if (next != null) setSel(next);
      } else if (e.key === 'ArrowUp' || e.key === 'k') {
        e.preventDefault();
        const idx = visibleSelectable.indexOf(sel);
        const next = visibleSelectable[Math.max(0, idx - 1)];
        if (next != null) setSel(next);
      } else if (e.key === 'Enter' || e.key === 'l' || e.key === 'ArrowRight') {
        e.preventDefault();
        const it = visibleItems[sel];
        if (it && it.type === 'item') {
          setActiveService(it);
          setView('detail');
        }
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [focused, sel, visibleSelectable, visibleItems, filterMode, view]);

  useEffect(() => {
    if (!autoplay) return;
    let cancelled = false;
    const steps: Array<() => void> = [
      () => setSel(visibleSelectable[0]),
      () => setSel(visibleSelectable[2]),
      () => setSel(visibleSelectable[3]),
      () => setSel(visibleSelectable[4]),
      () => {
        const it = visibleItems[visibleSelectable[4]];
        if (it && it.type === 'item') {
          setActiveService(it);
          setView('detail');
        }
      },
      () => {},
      () => setView('list'),
      () => setSel(visibleSelectable[7] ?? visibleSelectable[0]),
      () => setSel(visibleSelectable[11] ?? visibleSelectable[0]),
      () => {
        const it = visibleItems[visibleSelectable[11] ?? visibleSelectable[0]];
        if (it && it.type === 'item') {
          setActiveService(it);
          setView('detail');
        }
      },
      () => {},
      () => setView('list'),
    ];
    let i = 0;
    const tick = () => {
      if (cancelled) return;
      steps[i % steps.length]?.();
      i += 1;
      setTimeout(tick, 1400);
    };
    const t = setTimeout(tick, 900);
    return () => {
      cancelled = true;
      clearTimeout(t);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [autoplay]);

  useEffect(() => {
    const onDown = (e: MouseEvent) => {
      if (!rootRef.current) return;
      setFocused(rootRef.current.contains(e.target as Node));
    };
    document.addEventListener('mousedown', onDown);
    return () => document.removeEventListener('mousedown', onDown);
  }, []);

  const statusLine = filterMode
    ? `Filter: ${filter}_ (Enter to confirm, Esc to cancel)`
    : filter
    ? `Filter: ${filter} — ${visibleSelectable.length} match · Press / to edit · Esc to clear`
    : focused
    ? '↑/↓ navigate · / filter · Enter open · Esc back'
    : 'Click to focus · then ↑/↓ to navigate';

  return (
    <div
      ref={rootRef}
      data-tgcp-tui
      tabIndex={0}
      onFocus={() => setFocused(true)}
      className={`tui-root${focused ? ' is-focused' : ''}${compact ? ' is-compact' : ''}`}
    >
      <div className="tui-titlebar">
        <span className="tui-dots">
          <i style={{ background: '#ff5f56' }} />
          <i style={{ background: '#ffbd2e' }} />
          <i style={{ background: '#27c93f' }} />
        </span>
        <span className="tui-title">
          tgcp —{' '}
          {view === 'detail' && activeService
            ? activeService.name.toLowerCase().replace(/\s+/g, '-')
            : 'services'}
        </span>
        <span className="tui-cell">{TGCP_VERSION}</span>
      </div>

      <div className="tui-meta">
        <span>
          <span className="tui-muted">user</span> {DEMO_USER}
        </span>
        <span>
          <span className="tui-muted">project</span> {DEMO_PROJECT}
        </span>
        <span>
          <span className="tui-muted">region</span> {DEMO_REGION}
        </span>
      </div>

      {view === 'list' ? (
        <div className="tui-body">
          <div className="tui-header-row">
            <span className="tui-section-label">SERVICES</span>
            <span className="tui-muted">{visibleSelectable.length} items</span>
          </div>

          <div className="tui-list">
            {visibleItems.map((it, i) => {
              if (it.type === 'header') {
                return (
                  <div key={it._i} className="tui-group">
                    {it.label}
                  </div>
                );
              }
              const isSel = i === sel;
              return (
                <div
                  key={it._i}
                  className={`tui-item${isSel ? ' is-sel' : ''}`}
                  onMouseEnter={() => setSel(i)}
                  onClick={() => {
                    setActiveService(it);
                    setView('detail');
                    setFocused(true);
                  }}
                >
                  <span className="tui-chev">{isSel ? '▸' : ' '}</span>
                  <span className="tui-icon">{it.icon}</span>
                  <span className="tui-name">{it.name}</span>
                  <span className="tui-desc">{it.desc}</span>
                </div>
              );
            })}
            {visibleSelectable.length === 0 && (
              <div className="tui-empty">no matches for “{filter}”</div>
            )}
          </div>
        </div>
      ) : (
        <DetailView service={activeService} />
      )}

      <div className="tui-status">
        <span className={filter ? 'tui-accent' : ''}>{statusLine}</span>
        <span className="tui-muted">? help</span>
      </div>
    </div>
  );
}

function DetailView({ service }: { service: Service | null }) {
  const data = (service && FAKE_RESOURCES[service.name]) || FAKE_RESOURCES._default;
  return (
    <div className="tui-body">
      <div className="tui-header-row">
        <span className="tui-section-label">
          <span className="tui-accent">▸</span> {service?.name?.toUpperCase()}
        </span>
        <span className="tui-muted">
          {data.rows.length} resources · press ← to go back
        </span>
      </div>

      <div className="tui-table">
        <div className="tui-tr tui-th">
          {data.cols.map((c) => (
            <div key={c} className="tui-td">
              {c}
            </div>
          ))}
        </div>
        {data.rows.map((row, ri) => (
          <div key={ri} className={`tui-tr${ri === 0 ? ' is-sel' : ''}`}>
            {row.map((cell, ci) => (
              <div key={ci} className="tui-td">
                {ci === 0 ? <span className="tui-accent-text">{cell}</span> : cell}
                {typeof cell === 'string' &&
                  ['RUNNING', 'RUNNABLE', 'READY'].includes(cell) && (
                    <span className="tui-pip tui-pip-ok" />
                  )}
                {typeof cell === 'string' && cell === 'STOPPED' && (
                  <span className="tui-pip tui-pip-off" />
                )}
              </div>
            ))}
          </div>
        ))}
      </div>
    </div>
  );
}
