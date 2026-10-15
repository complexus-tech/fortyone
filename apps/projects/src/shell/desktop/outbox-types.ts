export type DesktopOutboxScope = { accountId: string; workspaceId: string };
export type DesktopOperationKind =
  | "story.create"
  | "story.update"
  | "comment.create"
  | "attachment.upload";
export type DesktopOperation = {
  id: string;
  kind: DesktopOperationKind;
  resourceId: string;
  payload: Record<string, unknown>;
  baseVersion?: string | null;
  status: "pending" | "conflict" | "blocked";
  error?: string | null;
  serverVersion?: string | null;
  serverValue?: Record<string, unknown> | null;
  attempted: boolean;
};
export type DesktopOperationAck = {
  operationId: string;
  kind: DesktopOperationKind;
  provisionalId: string;
  resourceId: string;
  operationVersion?: string | null;
  storyId?: string;
  resource?: Record<string, unknown>;
};
export type DesktopOutboxSnapshot = {
  operations: DesktopOperation[];
  completed: DesktopOperationAck[];
};
export type DesktopEnqueueOperation = DesktopOutboxScope & {
  id: string;
  kind: DesktopOperationKind;
  resourceId: string;
  payload: Record<string, unknown>;
  baseVersion?: string | null;
  consumedDraftKey?: string;
};
export type DesktopEnqueueResult = {
  operationId: string;
  resourceId: string;
  pending: boolean;
  resource?: Record<string, unknown>;
};
export type DesktopResolveOperation = DesktopOutboxScope & {
  operationId: string;
  resolution: "retry" | "discard";
  payload?: Record<string, unknown>;
  baseVersion?: string;
};

export const DESKTOP_OUTBOX_CHANGED = "fortyone:desktop-outbox-changed";
