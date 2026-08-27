import type { MayaToolName } from "./tool-names";

/**
 * Canonical action schemas for tools that combine reads and mutations. Tool
 * schemas and approval policy both consume this registry so a newly added
 * action cannot silently bypass classification tests.
 */
export const MAYA_TOOL_ACTIONS = {
  comments: ["list-comments", "add-comment", "reply-to-comment"],
  labels: ["list-labels", "create-label", "edit-label", "delete-label"],
  links: ["list-links", "add-link", "update-link", "delete-link"],
  notifications: [
    "list-notifications",
    "get-unread-count",
    "mark-as-read",
    "mark-all-as-read",
    "mark-as-unread",
    "delete-notification",
    "delete-all-notifications",
    "delete-read-notifications",
    "filter-notifications",
    "update-notification-preferences",
  ],
  objectiveStatuses: [
    "list-objective-statuses",
    "get-objective-status-details",
    "create-objective-status",
    "update-objective-status",
    "delete-objective-status",
    "set-default-objective-status",
  ],
  statuses: [
    "list-all-statuses",
    "list-team-statuses",
    "get-status-details",
    "create-status",
    "update-status",
    "delete-status",
    "set-default-status",
  ],
  storyLabels: [
    "get-story-labels",
    "set-story-labels",
    "add-labels-to-story",
    "remove-labels-from-story",
  ],
} as const satisfies Readonly<Partial<Record<MayaToolName, readonly string[]>>>;

export type MayaActionToolName = keyof typeof MAYA_TOOL_ACTIONS;
