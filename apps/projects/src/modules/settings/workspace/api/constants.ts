import type { DeveloperScope, WebhookEventType } from "./types";

export const DEVELOPER_SCOPES: {
  description: string;
  label: string;
  value: DeveloperScope;
}[] = [
  {
    value: "workspaces:read",
    label: "Workspace",
    description: "Read workspace metadata.",
  },
  {
    value: "teams:read",
    label: "Teams",
    description: "Read teams available to the credential.",
  },
  {
    value: "stories:read",
    label: "Read stories",
    description: "Read stories and their public API fields.",
  },
  {
    value: "stories:write",
    label: "Write stories",
    description: "Create and update supported story fields.",
  },
  {
    value: "comments:read",
    label: "Read comments",
    description: "Read story comments.",
  },
  {
    value: "comments:write",
    label: "Write comments",
    description: "Create and update supported comments.",
  },
  {
    value: "labels:read",
    label: "Read labels",
    description: "Read workspace and team labels.",
  },
  {
    value: "labels:write",
    label: "Write labels",
    description: "Manage supported label operations.",
  },
  {
    value: "sprints:read",
    label: "Sprints",
    description: "Read sprint metadata.",
  },
  {
    value: "objectives:read",
    label: "Read objectives",
    description: "Read objectives and key results.",
  },
  {
    value: "objectives:write",
    label: "Write objectives",
    description: "Manage supported objective operations.",
  },
  {
    value: "webhooks:manage",
    label: "Manage webhooks",
    description: "Create, rotate, update, and disable webhooks.",
  },
];

export const WEBHOOK_EVENTS: {
  label: string;
  value: WebhookEventType;
}[] = [
  { value: "story.created", label: "Story created" },
  { value: "story.updated", label: "Story updated" },
  { value: "story.deleted", label: "Story deleted" },
  { value: "comment.created", label: "Comment created" },
  { value: "comment.updated", label: "Comment updated" },
  { value: "comment.deleted", label: "Comment deleted" },
];

export const expiryFromNow = (days: number) =>
  new Date(Date.now() + days * 24 * 60 * 60 * 1000).toISOString();

export const formatDeveloperDate = (value?: string) => {
  if (!value) return "Never";
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
};
