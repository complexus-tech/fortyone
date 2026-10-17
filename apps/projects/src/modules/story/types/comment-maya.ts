export type CommentMayaStatus =
  | "queued"
  | "running"
  | "needs_input"
  | "awaiting_approval"
  | "succeeded"
  | "failed"
  | "cancelled";

export type CommentMayaRun = {
  id: string;
  sourceId: string;
  requesterId: string;
  status: CommentMayaStatus;
  replyCommentId?: string | null;
  summary: string;
  proposalReview?: string;
  updatedAt: string;
  version: number;
};

export type StoryMayaRuns = {
  enabled: boolean;
  runs: CommentMayaRun[];
};

export type CommentMayaAction = "cancel" | "retry" | "approve";

export type CommentMayaPendingStatus =
  | "thinking"
  | "working"
  | "delayed"
  | "offline"
  | "failed";

export const isMayaRunActive = (run: CommentMayaRun) =>
  run.status === "queued" || run.status === "running";
