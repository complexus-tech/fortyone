"use client";

import type { Editor } from "@tiptap/react";
import { useEffect, useId, useRef, useState } from "react";
import { Avatar, Button } from "ui";
import { ArrowDown2Icon, CheckIcon, QuoteIcon } from "icons";
import { MayaAvatar } from "@/components/ui/maya-avatar";
import { resolveCommentHighlight } from "@/lib/tiptap/comment-highlights";
import { useDocumentCommentMutations } from "./hooks";
import { DocumentCommentContent } from "./document-comment-content";
import { DocumentCommentReplyComposer } from "./document-comment-composer";
import type { DocumentComment, DocumentCommentThread } from "./types";
import type { DocumentMayaRun } from "./document-comment-maya";
import { isDocumentMayaActive } from "./document-comment-maya";
import { DocumentMayaPending } from "./document-maya-pending";
import { useDocumentMayaSources } from "./use-document-maya-sources";
import { useDocumentMentionRecipients } from "./use-document-mentions";

const formatTime = (value: string) => {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const minutes = Math.floor(Math.max(0, Date.now() - date.getTime()) / 60_000);
  if (minutes < 1) return "now";
  if (minutes < 60) return `${minutes}m`;
  if (minutes < 1_440) return `${Math.floor(minutes / 60)}h`;
  return new Intl.DateTimeFormat(undefined, {
    day: "numeric",
    month: "short",
  }).format(date);
};

function CommentMessage({
  comment,
  mayaId,
}: {
  comment: DocumentComment;
  mayaId?: string;
}) {
  const AuthorAvatar = comment.createdBy === mayaId ? MayaAvatar : Avatar;
  const sending = comment.id.startsWith("optimistic-");
  return (
    <div aria-busy={sending} className="flex min-w-0 items-start gap-2.5">
      <AuthorAvatar
        className="shrink-0"
        name={comment.authorName || "Teammate"}
        size="sm"
        src={comment.authorAvatar}
      />
      <div className="min-w-0 flex-1">
        <div className="mb-1.5 flex min-w-0 flex-wrap items-baseline gap-x-2 gap-y-0.5">
          <span className="text-foreground min-w-0 truncate text-base font-medium">
            {comment.authorName || "Teammate"}
          </span>
          <time
            className="text-text-muted shrink-0 text-base"
            dateTime={comment.createdAt}
            title={new Date(comment.createdAt).toLocaleString()}
          >
            {sending ? "Sending…" : formatTime(comment.createdAt)}
          </time>
        </div>
        <DocumentCommentContent body={comment.body} />
      </div>
    </div>
  );
}

export function DocumentCommentThreadCard({
  editor,
  thread,
  canComment,
  active,
  onThreadSelect,
  mayaRuns = [],
  mayaUnavailable = false,
  onCheckMaya = () => undefined,
}: {
  editor: Editor | null;
  thread: DocumentCommentThread;
  canComment: boolean;
  active: boolean;
  onThreadSelect?: (thread: DocumentCommentThread) => void;
  mayaRuns?: DocumentMayaRun[];
  mayaUnavailable?: boolean;
  onCheckMaya?: () => void;
}) {
  const { resolve, isThreadPending } = useDocumentCommentMutations(
    thread.documentId,
  );
  const { data: recipients } = useDocumentMentionRecipients(thread.documentId);
  const [expanded, setExpanded] = useState(false);
  const [replying, setReplying] = useState(false);
  const [replyToCommentId, setReplyToCommentId] = useState<string>();
  const [actionError, setActionError] = useState<string | null>(null);
  const changingResolution = useRef(false);
  const element = useRef<HTMLElement>(null);
  const repliesId = useId();
  const root = thread.comments.at(0);
  const replies = thread.comments.slice(1);
  const participants = [
    ...new Map(replies.map((comment) => [comment.createdBy, comment])).values(),
  ].slice(0, 3);
  const resolved = Boolean(thread.resolvedAt);
  const pending =
    thread.id.startsWith("optimistic-") || isThreadPending(thread.id);
  const canMutate = canComment && !pending;
  const pendingRuns = mayaRuns.filter(
    (run) =>
      thread.comments.some((comment) => comment.id === run.sourceId) &&
      (Boolean(run.replyCommentId) ||
        isDocumentMayaActive(run) ||
        run.status === "failed") &&
      !thread.comments.some((comment) => comment.id === run.replyCommentId),
  );
  const optimisticMayaSources = useDocumentMayaSources(
    thread.comments,
    mayaRuns,
  );
  const hasPendingMaya =
    pendingRuns.length > 0 || optimisticMayaSources.length > 0;

  useEffect(() => {
    if (!active) return;
    setExpanded(true);
    element.current?.scrollIntoView({ block: "nearest", behavior: "instant" });
  }, [active]);

  const locate = () => {
    onThreadSelect?.(thread);
    if (!editor) return;
    const anchor = resolveCommentHighlight(editor, {
      id: thread.id,
      from: thread.anchorStart,
      to: thread.anchorEnd,
      quote: thread.quote,
    });
    if (!anchor) {
      setActionError(
        "This passage has changed. The original quote is preserved above.",
      );
      return;
    }
    editor
      .chain()
      .setTextSelection({ from: anchor.from, to: anchor.to })
      .scrollIntoView()
      .run();
  };
  const toggleResolution = () => {
    if (!canMutate || changingResolution.current) return;
    changingResolution.current = true;
    setActionError(null);
    void resolve
      .mutateAsync({ threadId: thread.id, resolved: !resolved })
      .catch(() => {
        setActionError(
          `Could not ${resolved ? "reopen" : "resolve"} this thread. Try again.`,
        );
      })
      .finally(() => {
        changingResolution.current = false;
      });
  };

  return (
    <article
      className="border-border/70 border-b py-5 first:pt-0 last:border-b-0"
      ref={element}
    >
      <button
        className="text-text-muted hover:bg-foreground/5 focus-visible:ring-ring mb-4 flex w-full items-start gap-2 rounded-lg p-2 text-left text-base leading-6 transition-colors focus-visible:ring-2 focus-visible:outline-none"
        disabled={!editor && !onThreadSelect}
        onClick={locate}
        title="View quoted passage in document"
        type="button"
      >
        <QuoteIcon
          className="text-text-muted mt-1 size-4 shrink-0"
          strokeWidth={1.6}
        />
        <span className="border-primary/50 line-clamp-2 max-h-12 min-w-0 flex-1 border-l-2 pl-2.5 wrap-break-word whitespace-normal">
          {thread.quote}
        </span>
      </button>
      {root ? (
        <CommentMessage comment={root} mayaId={recipients?.mayaId} />
      ) : (
        <p className="text-text-muted text-base">
          This thread has no comments.
        </p>
      )}
      <div className="mt-3 ml-10 flex flex-wrap items-center gap-x-2 gap-y-1">
        {canComment && !resolved && !replying ? (
          <Button
            className="text-text-muted -ml-2 px-2 text-base"
            color="tertiary"
            disabled={!canMutate}
            onClick={() => {
              setReplyToCommentId(
                thread.comments.findLast(
                  (comment) => !comment.id.startsWith("optimistic-"),
                )?.id,
              );
              setExpanded(true);
              setReplying(true);
            }}
            size="sm"
            type="button"
            variant="naked"
          >
            Reply
          </Button>
        ) : null}
        {replies.length ? (
          <Button
            aria-controls={repliesId}
            aria-expanded={expanded}
            className="text-text-muted -ml-1 gap-2 px-2 text-base"
            color="tertiary"
            onClick={() => {
              setExpanded(!expanded);
              if (expanded) setReplying(false);
            }}
            size="sm"
            type="button"
            variant="naked"
          >
            {expanded ? (
              <ArrowDown2Icon className="size-3.5 rotate-180" strokeWidth={2} />
            ) : null}
            {!expanded ? (
              <span aria-hidden className="flex -space-x-1.5">
                {participants.map((person) => {
                  const ParticipantAvatar =
                    person.createdBy === recipients?.mayaId
                      ? MayaAvatar
                      : Avatar;
                  return (
                    <ParticipantAvatar
                      className="ring-surface ring-2"
                      key={person.createdBy}
                      name={person.authorName || "Teammate"}
                      size="xs"
                      src={person.authorAvatar}
                    />
                  );
                })}
              </span>
            ) : null}
            <span>
              {expanded
                ? "Hide replies"
                : `${replies.length} ${replies.length === 1 ? "reply" : "replies"}`}
            </span>
            {!expanded ? (
              <ArrowDown2Icon className="size-3.5" strokeWidth={2} />
            ) : null}
          </Button>
        ) : null}
        {resolved ? (
          <span className="text-text-muted flex items-center gap-1 text-base">
            <CheckIcon className="size-4" />
            Resolved
          </span>
        ) : null}
        {canComment ? (
          <Button
            className="text-text-muted ml-auto text-base"
            color="tertiary"
            disabled={!canMutate || resolve.isPending}
            leftIcon={resolved ? undefined : <CheckIcon className="size-4" />}
            onClick={() => {
              toggleResolution();
            }}
            size="sm"
            type="button"
            variant="naked"
          >
            {resolved ? "Reopen" : "Resolve"}
          </Button>
        ) : null}
      </div>
      {expanded && (replies.length > 0 || replying) ? (
        <div
          className="border-border/80 mt-4 ml-3.5 space-y-5 border-l pl-5"
          id={repliesId}
        >
          {replies.map((comment) => (
            <CommentMessage
              comment={comment}
              key={comment.id}
              mayaId={recipients?.mayaId}
            />
          ))}
          {replying && canComment && !resolved ? (
            <DocumentCommentReplyComposer
              canComment={canComment}
              documentId={thread.documentId}
              onCancel={() => {
                setReplying(false);
              }}
              onCreated={() => {
                setReplying(false);
              }}
              pending={pending}
              replyToCommentId={replyToCommentId}
              threadId={thread.id}
            />
          ) : null}
        </div>
      ) : null}
      {hasPendingMaya ? (
        <div className="border-border/80 mt-4 ml-3.5 space-y-5 border-l pl-5">
          {pendingRuns.map((run) => (
            <DocumentMayaPending
              documentId={thread.documentId}
              key={run.id}
              offline={mayaUnavailable}
              onCheck={onCheckMaya}
              run={run}
              since={run.updatedAt}
            />
          ))}
          {optimisticMayaSources.map((comment) => (
            <DocumentMayaPending
              documentId={thread.documentId}
              key={comment.id}
              offline={mayaUnavailable}
              onCheck={onCheckMaya}
              since={comment.createdAt}
            />
          ))}
        </div>
      ) : null}
      {actionError ? (
        <p className="text-danger mt-3 text-base leading-6" role="alert">
          {actionError}
        </p>
      ) : null}
    </article>
  );
}
