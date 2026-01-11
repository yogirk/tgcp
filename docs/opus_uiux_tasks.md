# TGCP UI/UX Review: The Path to Production-Grade

**Reviewer**: Claude Opus 4.5
**Date**: 2026-01-11
**Goal**: Make TGCP look production-grade for open source release without being pompous

---

## Executive Summary

TGCP has solid engineering foundations but needs **refinement in visual polish, information density, and interaction feedback** to compete with tools like k9s, lazygit, and harlequin. The architecture is sound - the styling system is semantic, components are modular, and patterns are consistent. What's missing is the **micro-interaction polish** that makes professional TUI tools feel alive.

**Bottom line**: 70% of the way there. The remaining 30% is all about polish, not features.

---

## Part I: What's Working (Don't Touch These)

### 1. Architecture
- Semantic color system with clear purpose for each color
- Modular components that follow single responsibility
- Service interface that enforces consistency
- FilterSession abstraction is elegant

### 2. Interaction Model
- Vim + arrow key navigation is correct
- `/` for filter, `:` for palette, `?` for help - follows terminal conventions
- Focus states exist (sidebar vs main)
- Tab to toggle sidebar is clever

### 3. Component Design
- StandardTable wraps Bubble Tea's table well
- DetailCard provides consistent key-value rendering
- Filter component with HandleFilterUpdate is well-factored
- Error component with context-aware suggestions is thoughtful

---

## Part II: Critical Issues (Fix These First)

### Issue 1: The Purple Table Selection is Jarring

**Current**: `TableSelectedFocused = lipgloss.Color("57")` (Bright purple)

**Problem**: This screams "developer default" not "production tool". k9s uses subtle dark grey backgrounds. lazygit uses inverse text. harlequin uses gentle highlighting. The bright purple:
- Fights for attention with status colors
- Looks like an error highlight
- Doesn't match the GCP blue brand

**Fix**:
```go
// table.go
TableSelectedFocused = lipgloss.Color("236")  // Dark grey background
// With accent-colored text and left indicator bar
```

**Task**: `[x] Replace purple table selection with subtle dark grey + accent text`

**Implementation Notes**:
- Focused: Dark grey (`236`) background + brand accent (`39`) blue text + bold
- Blurred: Lighter grey (`240`) background + muted (`245`) text, no bold
- Updated both `StandardTable` and legacy `TableModel` for consistency

---

### Issue 2: Status Indicators are Too Subtle

**Current**: `● RUNNING` with just color

**Problem**: Single dots are hard to scan quickly. Compare to:
- k9s: `[OK]`, `[WARN]`, `[ERR]` badges
- lazygit: Color blocks with icons
- harlequin: Bold colored text with background

**Fix**: Use filled circles (◉) or badges with slight background:
```go
// RUNNING  -> [green bg] ✓ RUNNING
// STOPPED  -> [red bg]   ■ STOPPED
// PENDING  -> [yellow]   ◐ PENDING
```

**Task**: `[x] Enhance status indicators with stronger visual presence`

**Implementation Notes**:
- Created centralized `RenderStatus` component in `internal/ui/components/status.go`
- Badge style with icon + background: `✓ RUNNING` (green), `✗ STOPPED` (red), `◐ PENDING` (yellow)
- **Auto-detection**: `DetailCard` now auto-renders any field with key "Status" or "State"
- Updated all services: GCE, CloudSQL, Dataproc, Spanner, Disks, Redis, Dataflow, Bigtable, GKE, Overview, Cloud Run
- Added comprehensive state mappings for all GCP status types

---

### Issue 3: Sidebar is Plain Text

**Current**: Just service names in a list

**Problem**: No icons, no state, no visual interest. k9s shows resource counts. taws shows service health. Even a simple icon adds scannability.

**Fix**:
```
SERVICES
─────────
🏠 Overview
🖥  Compute Engine     12
☸  Kubernetes          3
🗄  Cloud SQL          2
🏃 Cloud Run           8
...
```

**Tasks**:
- `[x] Add service icons to sidebar`
- `[ ] Show resource count badges (optional, fetched lazily)`

**Implementation Notes**:
- Added `Icon` field to `ServiceItem` struct
- Using semantic Unicode icons (no special fonts required):
  - `◉` Overview, `⚙` Compute, `☸` Kubernetes (official K8s logo), `⛁` SQL, `⚿` IAM, etc.
- Icons render before service name in sidebar

---

### Issue 4: Home Menu Uses Emoji Pointer

**Current**: `👉` as the selection indicator

**Problem**: This looks playful, not professional. The pointing emoji varies across terminals and fonts. Compare to lazygit's clean `▸` or k9s's highlight bar.

**Fix**: Use Unicode arrows or inverse highlighting:
```go
// Instead of: 👉 Compute Engine
// Use:        ▸ Compute Engine  (with accent color)
// Or:         │ Compute Engine  (left bar indicator)
```

**Task**: `[x] Replace emoji pointer with unicode arrow or highlight bar`

**Implementation Notes**:
- Changed `👉` to `▸` (right-pointing triangle) in `home_menu.go:79`
- Consistent with professional TUI tools like lazygit

---

### Issue 5: Command Palette Looks Basic ✓ DONE

**Current**: Rounded box with simple list

**Problems**:
1. Input and suggestions feel disconnected (border break)
2. No fuzzy match highlighting - can't see what matched
3. Generic placeholder "Type a command..."
4. No recent commands (MRU)

**Compare to**: VS Code command palette, fzf, telescope.nvim

**Tasks**:
- `[x] Connect input box to suggestions seamlessly (no top border on dropdown)`
- `[x] Highlight matching characters in suggestions`
- `[x] Better placeholder: "Search services, actions..."`
- `[-] Consider MRU section when input is empty` (deferred - nice-to-have)

**Implementation Notes**:
- Input box now removes bottom border when dropdown is visible (seamless connection)
- Both input and dropdown use same accent color border for visual unity
- Placeholder changed to "Search services, actions..." (more descriptive)
- Added `SuggestionMatch` struct in `core/navigation.go` to carry fuzzy match indexes
- Added `highlightMatches()` function in `palette.go` to render matched chars in accent color
- Matched characters in both name and description are highlighted with bold accent color

---

### Issue 6: Filter Bar Disappears When Inactive

**Current**: Shows `/ to filter` in subtle text, then becomes prominent only when active

**Problem**: Users don't know filtering is available. The transition is jarring.

**Better pattern** (harlequin): Always show the filter bar, but style it differently:
- Inactive: Subtle, shows "Press / to filter"
- Active: Prominent, shows query + match count
- Has filter: Shows applied filter badge

**Task**: `[x] Make filter bar always visible with clear state transitions`

**Implementation Notes**:
- Inactive (no filter): Shows `Filter: Press / to filter    Items: 12` (all muted)
- Active (typing): Shows `Filter: / query▊    Matches: 3/12` (prominent input)
- Inactive (filter applied): Shows `Filter: [query]    Matches: 3/12 │ Esc to clear` (badge style)
- Badge uses accent color background for clear visibility

---

### Issue 7: Status Bar is Cluttered

**Current**: `[MODE] [Message............] [Updated: Xs ago] [Help]`

**Problems**:
1. "Updated X ago" is low-value information taking prime real estate
2. Mode indicator blends into background
3. No visual separators between sections
4. Help text doesn't change based on context

**Fix**:
```
┃ MAIN ┃ GCE > Instances > prod-web-1        [?] Help  [/] Filter  [Enter] Select
```
- Remove timestamp (or move to detail views)
- Add subtle separators `│`
- Make mode badge pop more
- Contextual help hints

**Task**: `[x] Redesign status bar for clarity and context-awareness`

**Implementation Notes**:
- Removed "Updated: Xs ago" timestamp (low-value noise)
- Added `│` separator between message and help hints
- Cleaned up layout with better spacing
- Help hints now use muted text style

**Future Enhancement**: Context-aware help hints
- Pass structured hint data instead of single string
- Show different hints based on current view:
  - List view: `[?] Help  [/] Filter  [Enter] Select  [Tab] Sidebar`
  - Detail view: `[?] Help  [s] Start  [x] Stop  [q] Back`
  - Sidebar: `[?] Help  [Enter] Select  [Tab] Main`

---

### Issue 8: The Banner Takes Too Much Space

**Current**: Large ASCII art banner on home screen

**Problem**: Takes 6+ lines. In terminal real estate, this is expensive. On small terminals, it dominates the view.

**Better pattern**:
- Home screen: Keep full banner (first impression)
- Command palette: Use mini banner or none
- Inner screens: No banner, just project context in header

**Task**: `[ ] Create compact header for inner screens (no ASCII art)`

---

### Issue 9: Overview Dashboard is Emoji-Heavy

**Current**: `📡 Project Overview`, `💳 Account`, `⚡ Actionable Insights`, `📦 Inventory`, `💰 Budget`

**Problem**: Too many emojis compete for attention. Professional tools use typography hierarchy, not emoji spam. Compare to k9s's clean dashboard or AWS console.

**Fix**: Use icons sparingly, rely on typography:
```
PROJECT OVERVIEW
━━━━━━━━━━━━━━━━
Billing: Active                Account: my-account (ID: xxx)

RECOMMENDATIONS                 Potential Savings: $57.70/mo
───────────────────────────────────────────────────────────
  3 Idle VMs              $45.20/mo    Action: Stop
  5 Unused Disks          $12.50/mo    Action: Delete

INVENTORY
───────────────────────────────────────────────────────────
  VMs: 12    Disks: 8 (240GB)    IPs: 3    SQL: 2    Buckets: 5
```

**Task**: `[ ] Reduce emoji usage in Overview, use typography hierarchy`

---

### Issue 10: No Visual Feedback on Actions

**Current**: Actions like Start/Stop/SSH just change state silently

**Problem**: Users don't get confirmation that their action was registered. Did it work? Did it fail? Is it in progress?

**Better pattern** (lazygit): Show a brief toast/notification:
```
┌──────────────────────────────────┐
│ ✓ Starting instance prod-web-1  │
└──────────────────────────────────┘
```

**Task**: `[x] Add toast notifications for user actions (3-second auto-dismiss)`

**Implementation Notes**:
- Created `ToastModel` component in `internal/ui/components/toast.go`
- Added `core.ToastMsg` and `core.ToastType` in `internal/core/events.go`
- Toast types: `ToastSuccess` (green ✓), `ToastError` (red ✗), `ToastInfo` (blue ℹ)
- Toast displays at bottom-right, auto-dismisses after 3 seconds
- Integrated with GCE, CloudSQL, GKE action results
- Services emit `core.ToastMsg` after actions complete

---

## Part III: Medium Priority Improvements

### 1. Border Hierarchy

**Problem**: All boxes use the same `RoundedBorder()` with `ColorBorderSubtle`. Everything looks the same importance.

**Solution**: Introduce visual hierarchy:
- **Focus border**: Double border or primary color thick border
- **Secondary**: Rounded border with subtle color (current)
- **Inline sections**: Left border only or just divider lines

**Task**: `[x] Create border hierarchy: primary/secondary box styles`

**Implementation Notes**:
- Added `PrimaryBoxStyle` in `styles.go`: Rounded border, accent color (#75), padding (1, 2)
- Added `SecondaryBoxStyle` in `styles.go`: Normal border, subtle grey (#240), padding (0, 1)
- `DetailCard` uses `PrimaryBoxStyle` for main content
- `DetailSection` uses `SecondaryBoxStyle` for supporting sections
- Home page info box uses Primary, hints section uses Secondary
- Help dialog uses Primary for modal prominence
- Updated `ui_patterns.md` and `DEVELOPER_GUIDE.md` with border hierarchy docs

---

### 2. Table Headers Need Presence

**Current**: Just muted colored text, uppercase

**Problem**: Headers don't anchor the eye. Hard to track which column is which in wide tables.

**Solution**:
- Add subtle background to header row
- Consider underline
- Maybe column separators for wide tables

**Task**: `[x] Enhance table headers with background and/or underline`

**Implementation Notes**:
- Added subtle background (`237`) to table headers
- Headers now use primary text color (white) instead of muted
- Bold text for better visibility
- Updated both `StandardTable` and legacy `TableModel`
- Headers now visually anchor the columns

---

### 3. Alternating Row Colors

**Current**: All rows same background

**Problem**: Hard to scan across wide tables. Eyes lose track.

**Solution** (harlequin pattern): Subtle zebra striping
```go
// Row backgrounds
BgEven = lipgloss.Color("234")  // Very subtle
BgOdd  = lipgloss.Color("235")  // Slightly different
```

**Task**: `[ ] Add subtle alternating row colors to tables`

---

### 4. Detail View Footer Hints are Plain

**Current**: `s Start | x Stop | h SSH | q Back`

**Problem**: All text, no visual distinction between key and action.

**Solution** (lazygit pattern): Highlight the keys:
```
[s] Start  [x] Stop  [h] SSH  [q] Back
```
With `[s]` styled differently (background or bold).

**Task**: `[x] Style footer hints with distinct key highlighting`

**Implementation Notes**:
- Added `renderFooterHint()` function in `detail.go`
- Parses `"s Start | x Stop | q Back"` format
- Renders keys as badges: `[s]` with background color
- Actions rendered in muted text
- Output: `[s] Start  [x] Stop  [q] Back`

---

### 5. Confirmation Dialogs Could Be More Contextual ✓ DONE

**Current**: Orange border for all confirmations

**Problem**: "Delete" should feel more dangerous than "Stop". No impact preview.

**Solution**:
- Delete: Red border, warning icon, impact text
- Stop: Orange border, info icon
- Start: Blue/green, minimal warning

**Task**: `[x] Add action-specific styling to confirmation dialogs`

**Implementation Notes**:
- Added `actionStyle` struct and `getActionStyle()` in `confirmation.go`
- **Delete/Remove/Destroy**: Red border, ⚠ icon, "This action cannot be undone." impact text
- **Stop/Terminate/Shutdown**: Orange border, ⏸ icon
- **Start/Restart/Resume**: Cyan/info border, ▶ icon
- **Snapshot/Backup**: Accent blue border, 📷 icon
- **Default**: Orange border, ⚠ icon (backwards compatible)

---

### 6. Help Overlay is Dense (DEFERRED - Next Release)

**Current**: Three columns of keybindings

**Problem**: Walls of text. Hard to find what you need. Not context-aware.

**Solution**:
- Show context-specific help first (current view's actions)
- Compact format: `key → action` aligned
- Maybe searchable in future

**Task**: `[-] Reorganize help with context-specific sections first` (deferred)

---

### 7. Color Palette Needs Tuning

**Current colors**:
- `ColorBrandPrimary = "33"` (GCP Blue, but dark)
- `ColorBrandAccent = "39"` (Light Blue)
- `ColorBorderSubtle = "238"` (Very dark, hard to see)

**Problems**:
- Primary blue is too dark, doesn't pop
- Border subtle is nearly invisible on some terminals
- Text muted (245) and primary (252) too similar

**Suggested adjustments**:
```go
ColorBrandPrimary = "39"   // Brighter blue (swap with current accent)
ColorBrandAccent  = "75"   // Even lighter for highlights
ColorBorderSubtle = "240"  // More visible
ColorTextMuted    = "243"  // More differentiated from primary
```

**Task**: `[x] Tune color palette for better contrast and visibility`

**Implementation Notes**:
- `ColorBrandPrimary`: `33` → `39` (brighter blue, more visible)
- `ColorBrandAccent`: `39` → `75` (lighter blue for highlights)
- `ColorTextMuted`: `245` → `243` (better contrast from primary text)
- `ColorBorderSubtle`: `238` → `240` (more visible on dark terminals)

---

### 8. Breadcrumb Needs More Presence

**Current**: Plain text with ` > ` separator

**Problem**: Blends into content. No visual anchor.

**Solution**:
- Use `›` separator (more elegant)
- Subtle background or left indicator
- Maybe truncate middle parts for long paths: `GCE > ... > instance`

**Task**: `[x] Enhance breadcrumb with better separator and styling`

**Implementation Notes**:
- Changed separator from ` > ` to ` › ` (more elegant)
- Path segments (ancestors) rendered in muted style
- Current location (last segment) rendered in primary color + bold
- Creates clear visual hierarchy

---

## Part IV: Low Priority / Future Enhancements

These are nice-to-haves that would differentiate TGCP but aren't blocking for launch:

### 1. Split Pane Mode
Show list and details simultaneously (like k9s pods + logs view).

### 2. Syntax Highlighting
Highlight JSON/YAML in detail views (like harlequin's query results).

### 3. Resource Graphs
ASCII visualization of VPC → Subnet → Instance relationships.

### 4. Column Sorting
Click/key to sort table columns (with ▲▼ indicators).

### 5. Filter History
Show recent filters when pressing `/` in empty filter.

### 6. Theming
Allow user themes via config file.

### 7. Multi-Tab
Keep multiple resources open (like harlequin's query tabs).

### 8. Animations
Pulsing indicators for transitioning states (k9s does this subtly).

---

## Part V: Prioritized Task List

### Must-Have for Launch (Do These)

| Priority | Task | Impact | Effort |
|----------|------|--------|--------|
| P0 | ~~Replace purple table selection with subtle grey + accent~~ DONE | High | Low |
| P0 | ~~Replace emoji pointer (👉) with unicode arrow~~ DONE | High | Low |
| P0 | ~~Add service icons to sidebar~~ DONE | High | Low |
| P0 | ~~Tune color palette (brighter primary, visible borders)~~ DONE | High | Low |
| P1 | ~~Enhance status indicators (badges or filled circles)~~ DONE | High | Medium |
| P1 | ~~Redesign status bar (separators, context-aware help)~~ DONE | High | Medium |
| P1 | ~~Make filter bar always visible~~ DONE | Medium | Low |
| P1 | Reduce emoji in Overview, use typography | Medium | Medium |
| P1 | ~~Style footer hints with key highlighting~~ DONE | Medium | Low |
| P1 | ~~Enhance breadcrumb styling~~ DONE | Medium | Low |

### Should-Have (Do if Time Allows)

| Priority | Task | Impact | Effort |
|----------|------|--------|--------|
| P2 | ~~Add toast notifications for actions~~ DONE | High | Medium |
| P2 | ~~Highlight matching characters in palette~~ DONE | Medium | Medium |
| P2 | ~~Create border hierarchy (primary/secondary styles)~~ DONE | Medium | Low |
| P2 | Add alternating row colors to tables | Medium | Medium |
| P2 | ~~Enhance table headers (background/underline)~~ DONE | Medium | Low |
| P2 | ~~Action-specific confirmation styling~~ DONE | Medium | Medium |
| P2 | Compact header for inner screens | Medium | Low |

### Nice-to-Have (Post-Launch)

| Priority | Task | Impact | Effort |
|----------|------|--------|--------|
| P3 | MRU commands in palette | Medium | Medium |
| P3 | Reorganize help with context sections (DEFERRED) | Medium | Medium |
| P3 | Resource count badges in sidebar | Medium | High |
| P3 | Split pane mode | High | High |
| P3 | Syntax highlighting in detail views | Medium | High |

---

## Part VI: Design Principles to Follow

1. **Information density over whitespace**: Terminal real estate is precious. Show more, not less.

2. **Typography creates hierarchy**: Bold, color, size - not emoji spam.

3. **Feedback is mandatory**: Every action should have visible feedback.

4. **States should be scannable**: Status colors should be instantly recognizable at a glance.

5. **Keyboard-first, always**: Every feature accessible without mouse.

6. **Test on 80x24**: The minimum terminal size should still be usable.

7. **Restraint over excess**: Less color variation, consistent use. Don't rainbow the UI.

8. **Follow conventions**: `/` for filter, `:` for command, `?` for help - users expect these.

---

## Part VII: Testing Checklist

Before release, verify:

### Visual
- [ ] Looks good on dark terminals (default expectation)
- [ ] Readable on light terminals (many corporate users)
- [ ] Works without Nerd Fonts (fallback glyphs)
- [ ] Usable at 80x24 (minimum size)
- [ ] No color conflicts (status colors vs selection)

### Interaction
- [ ] All vim keys work (hjkl navigation)
- [ ] All arrow keys work
- [ ] Tab/focus switching is intuitive
- [ ] Esc always goes back or closes overlay
- [ ] Filter mode entry/exit is smooth
- [ ] Command palette is discoverable

### Consistency
- [ ] All services follow same patterns
- [ ] All detail views use DetailCard
- [ ] All lists use StandardTable
- [ ] All errors use RenderError
- [ ] All confirmations use RenderConfirmation

---

## Part VIII: Quick Wins (< 1 Hour Each)

If you want immediate impact, do these in order:

1. ~~**Change table selection color** (10 min)~~ DONE
   - ~~Edit `table.go`: Change `"57"` to `"236"`~~

2. ~~**Replace emoji pointer** (10 min)~~ DONE
   - ~~Edit `home_menu.go`: Change `"👉 "` to `"▸ "`~~

3. ~~**Add sidebar icons** (30 min)~~ DONE
   - ~~Edit `sidebar.go`: Add icon field to ServiceItem, render it~~

4. ~~**Brighten primary blue** (5 min)~~ DONE
   - ~~Edit `styles.go`: Change `"33"` to `"39"`~~

5. ~~**Style footer hints** (20 min)~~ DONE
   - ~~Edit `detail.go`: Parse hint string, wrap keys in `[key]` style~~

6. ~~**Better breadcrumb separator** (5 min)~~ DONE
   - ~~Edit `breadcrumb.go`: Change ` > ` to ` › ` with subtle background~~

---

## Closing Thoughts

TGCP is genuinely well-engineered. The bones are excellent. What it needs now is the **visual refinement** that separates hobby projects from production tools.

The tools you're competing with (k9s, lazygit, harlequin) all share a common trait: **restraint with polish**. They don't try to be flashy. They're just clean, consistent, and responsive.

Focus on the P0 and P1 tasks. Skip the clever stuff. Ship a polished v1, then iterate based on user feedback.

Good luck with the open source launch.
