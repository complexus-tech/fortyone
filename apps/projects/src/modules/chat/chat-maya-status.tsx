"use client";

import { Button } from "ui";
import { useSession } from "@/lib/auth/client";
import { ChatTaskResults } from "./chat-task-results";
import { useChatMayaAction, useChatMayaRuns } from "./use-chat-maya-runs";

export function ChatMayaStatus({
  conversationId,
  sourceMessageId,
}: {
  conversationId: string;
  sourceMessageId: string;
}) {
  const { data } = useChatMayaRuns(conversationId);
  const { data: session } = useSession();
  const action = useChatMayaAction(conversationId);
  const run = data?.runs.find(
    (candidate) =>
      candidate.sourceId === sourceMessageId ||
      candidate.replyMessageId === sourceMessageId,
  );
  if (!run) return null;
  const source = run.sourceId === sourceMessageId;
  const working = run.status === "queued" || run.status === "running";
  const awaiting = run.status === "awaiting_approval";
  const failed = run.status === "failed";
  const canControl = run.requesterId === session?.user.id;
  if (!source)
    return run.tasks?.length ? <ChatTaskResults tasks={run.tasks} /> : null;
  if (working || (!awaiting && !failed)) return null;
  let status = run.summary || "Maya couldn’t finish this request.";
  if (awaiting) status = "Maya has a proposal for you to review.";
  return (
    <div className="border-border bg-surface-elevated mt-2 max-w-lg rounded-xl border p-3">
      <p aria-live="polite" className="text-text-muted" role="status">
        {status}
      </p>
      {awaiting && run.proposalReview ? (
        <p className="mt-2 whitespace-pre-wrap">{run.proposalReview}</p>
      ) : null}
      {canControl ? (
        <div className="mt-2 flex gap-2">
          {awaiting ? (
            <Button
              disabled={action.isPending}
              onClick={() => {
                action.mutate({ run, action: "approve" });
              }}
              size="sm"
            >
              Approve
            </Button>
          ) : null}
          {failed ? (
            <Button
              disabled={action.isPending}
              onClick={() => {
                action.mutate({ run, action: "retry" });
              }}
              size="sm"
            >
              Retry
            </Button>
          ) : null}
          {awaiting ? (
            <Button
              color="tertiary"
              disabled={action.isPending}
              onClick={() => {
                action.mutate({ run, action: "cancel" });
              }}
              size="sm"
              variant="naked"
            >
              Cancel
            </Button>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}
