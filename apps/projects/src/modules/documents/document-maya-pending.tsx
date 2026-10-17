import { useEffect, useState } from "react";
import { Button } from "ui";
import { MayaAvatar } from "@/components/ui/maya-avatar";
import { useSession } from "@/lib/auth/client";
import { useDocumentMayaAction } from "./document-comment-maya";
import type { DocumentMayaRun } from "./document-comment-maya";

export function DocumentMayaPending({
  documentId,
  run,
  since,
  offline,
  onCheck,
}: {
  documentId: string;
  run?: DocumentMayaRun;
  since: string;
  offline?: boolean;
  onCheck: () => void;
}) {
  const { data: session } = useSession();
  const action = useDocumentMayaAction(documentId);
  const [now, setNow] = useState(Date.now);
  useEffect(() => {
    const timer = window.setInterval(() => {
      setNow(Date.now());
    }, 10_000);
    return () => {
      window.clearInterval(timer);
    };
  }, []);
  const age = now - Date.parse(since);
  const failed = run?.status === "failed" && !run.replyCommentId;
  const delayed = age > 180_000;
  const active = !failed && !offline && !delayed;
  let label = age > 30_000 ? "Still working on this" : "Thinking";
  if (run?.replyCommentId) label = "Replying";
  if (delayed) label = "This is taking longer than usual.";
  if (offline) label = "Couldn’t check for a response.";
  if (failed) label = "Couldn’t finish this.";
  return (
    <div className="flex items-start gap-2.5">
      <MayaAvatar size="sm" />
      <div className="min-w-0 flex-1">
        <p className="text-foreground mb-1.5 text-base font-medium">Maya</p>
        <span
          aria-live="polite"
          className="text-text-muted inline-flex items-baseline gap-[3px] text-base leading-6"
          role="status"
        >
          {label}
          {active ? (
            <span
              aria-hidden
              className="comment-maya-dots inline-flex items-end gap-[2px]"
            >
              <span />
              <span />
              <span />
            </span>
          ) : null}
        </span>
        {failed && run.requesterId === session?.user.id ? (
          <Button
            className="ml-1 text-[0.95rem]"
            color="tertiary"
            disabled={action.isPending}
            onClick={() => {
              action.mutate({ run, action: "retry" });
            }}
            size="sm"
            variant="naked"
          >
            Retry
          </Button>
        ) : null}
        {!failed && (offline || delayed) ? (
          <Button
            className="ml-1 text-[0.95rem]"
            color="tertiary"
            onClick={onCheck}
            size="sm"
            variant="naked"
          >
            Check again
          </Button>
        ) : null}
      </div>
    </div>
  );
}
