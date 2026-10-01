# Roadmap capability boundaries

Custom fields, work presets and team automations retain their own request,
validation, query and mutation implementations. Cross-feature consumers use
explicit capability entrypoints rather than importing those implementations.

| Owner              | Public entrypoint            | Consumer contract                                                                             |
| ------------------ | ---------------------------- | --------------------------------------------------------------------------------------------- |
| `custom-fields`    | `public/types.ts`            | Field definitions and value payload types for templates and migration.                        |
| `custom-fields`    | `public/schemas.ts`          | Runtime validation of imported icons and template values.                                     |
| `custom-fields`    | `public/client.ts`           | The owned task-value query used when saving a template.                                       |
| `custom-fields`    | `public/creation.ts`         | Creation properties and their draft preparation workflow.                                     |
| `custom-fields`    | `public/properties.ts`       | Editable task-detail properties, including their mutation policy.                             |
| `custom-fields`    | `public/report.ts`           | The custom-field reporting workflow composed into Analytics.                                  |
| `custom-fields`    | `public/settings.ts`         | Team field-definition management.                                                             |
| `custom-fields`    | `public/display.ts`          | Property loading, badges and field selection composed into boards and lists by the shell.     |
| `custom-fields`    | `public/import.ts`           | Live definition matching, creation, option reconciliation and value validation for migration. |
| `work-presets`     | `public/types.ts`            | Versioned view and template payload types.                                                    |
| `work-presets`     | `public/schemas.ts`          | The task-template validator used by recurrence.                                               |
| `work-presets`     | `public/template-picker.ts`  | The owned template-selection workflow.                                                        |
| `work-presets`     | `public/template-actions.ts` | Saving an existing task as a template.                                                        |
| `work-presets`     | `public/views.ts`            | The owned saved-view selection and management workflow.                                       |
| `team-automations` | `public/settings.ts`         | Rule, recurrence and run-history management for a team.                                       |
| `auth`             | `public/sso-status.ts`       | The anonymous, validated workspace SSO availability probe.                                    |

UI entrypoints declare `use client`. Type and schema boundaries contain no
session or server credentials. Import and SSO adapters use the existing HTTP
clients and keep their response validation with the resource owner; these
entrypoints do not create a second query or mutation implementation.

Team selection consumes `teams/public/client.ts`. Shared settings presentation
uses `components/ui/section-header.tsx`, avoiding a dependency from field or
automation ownership back into workspace settings.

Board and list consumers depend on the typed component contract in
`shared/story/board-property-slots.tsx`. The shell supplies the custom-field
provider, badges and display picker, plus the story owner's workflow-count
provider. Feature queries stay inside their owning components. The shared
contract contains no module import or injectable hook implementation.

Without shell composition, the default property provider preserves native task
children, optional property UI renders nothing, and workflow counts are an
empty read-only map. Missing counts remain unknown. Stable component references
preserve interaction state when the shell's slot object changes.

Task creation uses `shared/story/creation-property-slots.tsx`. The custom-field
component owns its draft and exposes a typed ref controller for preparing
values, resetting and applying template values. The template picker receives a
transport-neutral template configuration. Default slots expose an empty
controller and hide the picker, preserving native task creation without
installing feature hooks into the shared component.

`shell/work-feature-slots.tsx`, mounted by `app/providers.tsx`, installs stable
component references from the owning public boundaries. Components remain
feature-owned and run their own hooks inside normal React component lifecycles.

Saved Views compose at the team-story client route adapter. The adapter supplies
the preset owner's `SavedViews` component through the team owner's explicit
`SavedViewsAction` contract. Teams own applying the view to filters, layout and
view options. `shared/story/view-configuration.ts` owns the transport shape, so
team views do not depend on preset persistence or request implementation.

Terminology reads the hydrated canonical team list at
`teamKeys.lists(workspaceSlug)` through the feature-neutral
`shared/query-cache/use-cached-query-data.ts` subscription. The Teams owner
continues to own fetching and updating the resource. Team-specific names can
react to authoritative cache updates without adding a second request owner or
depending on private team hooks.

The public paths remain part of the production dependency graph. A public
entrypoint does not waive cycle, layer, server-safety or authorization checks.
The architecture baseline is unchanged.
