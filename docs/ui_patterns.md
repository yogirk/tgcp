# TGCP UI Patterns

This document defines the UI patterns to use for list and detail views so new
services stay consistent.

## Detail Layout Decision Tree

1. **Single entity with <15 fields** → use `DetailCard`.
2. **Entity + secondary lists/sections** → use `DetailCard` + `DetailSection`.
3. **Primarily list/table view** → keep table view; use breadcrumb + filter bar.
4. **Schema/table drilldowns** → keep table; use breadcrumb only.

## Required Elements

- **Breadcrumbs** on all list and detail views via `components.Breadcrumb`.
- **Filter bar** on list views via `filter.View()` (with counts).
- **Detail card** on detail views, unless explicitly table-driven.

## Component Overview

### DetailHeader (breadcrumb)
Use `components.Breadcrumb("Project ...", service, ...)` at the top of the view.

### DetailCard
Use for single-entity details.

### DetailSection
Use for secondary blocks inside a detail view (e.g., clusters, node pools).

## Existing Service Patterns (Reference)

- **DetailCard only:** GCE, Cloud SQL, Redis, Spanner, Firestore, Pub/Sub, IAM
- **DetailCard + Section:** GKE, Bigtable, Disks, Dataproc, Dataflow
- **Table-driven details:** BigQuery schema, GCS object list

If a new service doesn't fit, add a short note here explaining the choice.

## UI Consistency Checklist

When adding a new service, ensure:

- ✅ Add breadcrumb at top of list and detail views using `components.Breadcrumb`
- ✅ Keep the filter bar visible on list views with counts
- ✅ Choose detail layout per the decision tree above:
  - Use `DetailCard` for single-entity details (<15 fields)
  - Use `DetailSection` for secondary blocks (clusters, node pools, attachments)
  - Use table view only when primarily showing lists/tables
- ✅ Follow shared component patterns:
  - Use `StandardTable` for all tables
  - Use `components.RenderError()` for errors
  - Use `components.RenderSpinner()` for loading states
  - Use `components.RenderConfirmation()` for confirmations
- ✅ Implement proper caching with TTL
- ✅ Handle errors gracefully with recovery
- ✅ Keep UI responsive using `tea.Cmd` for async operations
