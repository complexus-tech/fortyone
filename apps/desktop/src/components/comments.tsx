import { useState } from "react";
import type { Comment, OperationInput, WorkspaceData } from "../types";
import { useDraft } from "../hooks/use-draft";
import { errorMessage } from "../native";
import { formatTime } from "../model";
import { ErrorNotice } from "./primitives";

export function Comments({
  storyId,
  data,
  save,
}: {
  storyId: string;
  data: WorkspaceData;
  save: (input: OperationInput) => Promise<void>;
}) {
  const draft = useDraft(data.workspaceId, storyId, "comment", { comment: "" });
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const submit = async () => {
    setBusy(true);
    setError(null);
    try {
      await save({
        workspaceId: data.workspaceId,
        resourceId: crypto.randomUUID(),
        kind: "comment.create",
        baseVersion: null,
        payload: { storyId, comment: draft.value.comment },
      });
      draft.update({ comment: "" });
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setBusy(false);
    }
  };
  return (
    <section className="comments">
      <h2>Activity</h2>
      {!data.commentsCompleteFor?.includes(storyId) &&
      !data.pending.some(
        (operation) =>
          operation.kind === "story.create" && operation.resourceId === storyId,
      ) ? (
        <p className="footnote" role="status">
          Only downloaded comments are shown. Open this story while connected to
          load the full conversation.
        </p>
      ) : null}
      {data.comments
        .filter((comment) => comment.storyId === storyId)
        .map((comment) => (
          <CommentThread key={comment.id} comment={comment} />
        ))}
      <ErrorNotice message={error ?? draft.error} />
      <label className="sr-only" htmlFor="comment">
        Write a comment
      </label>
      <textarea
        id="comment"
        rows={3}
        placeholder="Write a comment…"
        value={draft.value.comment}
        disabled={!draft.ready || busy}
        onChange={(event) => draft.update({ comment: event.target.value })}
      />
      <div className="comment-actions">
        <span className="save-status" role="status">
          {draft.status}
        </span>
        <button
          disabled={!draft.ready || busy || !draft.value.comment.trim()}
          onClick={() => {
            void submit();
          }}
        >
          Post comment
        </button>
      </div>
    </section>
  );
}

function CommentThread({ comment }: { comment: Comment }) {
  return (
    <article>
      <div>
        <strong>{comment.authorName ?? "Team member"}</strong>
        <time>{formatTime(comment.createdAt)}</time>
      </div>
      <p>{comment.content ?? comment.body ?? comment.comment ?? ""}</p>
      {comment.subComments?.length ? (
        <div className="comment-replies">
          {comment.subComments.map((reply) => (
            <CommentThread key={reply.id} comment={reply} />
          ))}
        </div>
      ) : null}
    </article>
  );
}
