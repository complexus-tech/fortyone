"use client";

import type { Editor } from "@tiptap/react";
import { useEffect, useRef, useState } from "react";
import { Button } from "ui";
import { CloseIcon, Comment01Icon } from "icons";
import { cn } from "lib";
import { useDocumentComments } from "./hooks";
import { DocumentCommentComposer } from "./document-comment-composer";
import { DocumentCommentThreadCard } from "./document-comment-thread";
import type { DocumentCommentSelection, DocumentCommentThread } from "./types";
import {
  useDocumentMayaRefresh,
  useDocumentMayaRuns,
} from "./document-comment-maya";

export { DocumentCommentComposer } from "./document-comment-composer";

export function DocumentCommentsPanel({
  draft,
  documentId,
  editor,
  onClose,
  onDraftChange,
  canComment = true,
  activeThreadId,
  threadScope,
  onThreadSelect,
}: {
  draft: DocumentCommentSelection | null;
  documentId: string;
  editor: Editor | null;
  onClose: () => void;
  onDraftChange: (draft: DocumentCommentSelection | null) => void;
  canComment?: boolean;
  activeThreadId?: string;
  threadScope?: readonly string[];
  onThreadSelect?: (thread: DocumentCommentThread) => void;
}) {
  const {
    data: allThreads = [],
    isPending,
    isError,
    isFetching,
    refetch,
  } = useDocumentComments(documentId);
  const scopedThreadIds = threadScope ? new Set(threadScope) : undefined;
  const threads = scopedThreadIds
    ? allThreads.filter((thread) => scopedThreadIds.has(thread.id))
    : allThreads;
  const maya = useDocumentMayaRuns(documentId, true);
  useDocumentMayaRefresh(documentId, maya.data?.runs);
  const [filter, setFilter] = useState<"open" | "resolved">("open");
  const openCount = threads.filter((thread) => !thread.resolvedAt).length;
  const visible = threads.filter((thread) =>
    filter === "open" ? !thread.resolvedAt : Boolean(thread.resolvedAt),
  );
  const activeThread = threads.find((thread) => thread.id === activeThreadId);
  let activeFilter: "open" | "resolved" | undefined;
  if (activeThread)
    activeFilter = activeThread.resolvedAt ? "resolved" : "open";
  let emptyDescription = "Comments on this document will appear here.";
  if (filter === "resolved")
    emptyDescription =
      "Resolved discussions will appear here. You can reopen them anytime.";
  else if (canComment)
    emptyDescription =
      "Select a passage in the document to start a conversation.";
  const selectedThread = useRef<string | undefined>(undefined);
  const selectedScope = useRef(threadScope);
  useEffect(() => {
    if (!activeThreadId) selectedThread.current = undefined;
    if (
      activeFilter &&
      (selectedThread.current !== activeThreadId ||
        selectedScope.current !== threadScope)
    ) {
      selectedThread.current = activeThreadId;
      selectedScope.current = threadScope;
      setFilter(activeFilter);
    }
  }, [activeFilter, activeThreadId, threadScope]);

  return (
    <aside
      aria-labelledby="document-comments-title"
      className="bg-surface flex h-full min-h-0 min-w-0 flex-col text-base"
    >
      <header className="border-border shrink-0 border-b px-5 pt-4">
        <div className="flex items-center justify-between gap-3">
          <h2
            className="text-foreground flex items-center gap-2.5 text-lg font-semibold"
            id="document-comments-title"
          >
            <Comment01Icon className="size-5" strokeWidth={1.7} />
            Comments
          </h2>
          <Button
            aria-label="Close comments"
            asIcon
            color="tertiary"
            onClick={onClose}
            size="sm"
            type="button"
            variant="naked"
          >
            <CloseIcon className="size-5" />
          </Button>
        </div>
        <div
          aria-label="Comment status"
          className="mt-4 flex gap-5"
          role="group"
        >
          {(
            [
              { label: "Open", value: "open", count: openCount },
              {
                label: "Resolved",
                value: "resolved",
                count: threads.length - openCount,
              },
            ] as const
          ).map((item) => (
            <button
              aria-pressed={filter === item.value}
              className={cn(
                "text-text-muted relative flex items-center gap-2 pb-3 text-base font-medium transition-colors",
                filter === item.value &&
                  "text-foreground after:bg-primary after:absolute after:inset-x-0 after:-bottom-px after:h-0.5 after:rounded-full",
              )}
              key={item.value}
              onClick={() => {
                setFilter(item.value);
              }}
              type="button"
            >
              {item.label}
              <span className="text-text-muted font-normal">{item.count}</span>
            </button>
          ))}
        </div>
      </header>
      <div className="min-h-0 flex-1 overflow-y-auto overscroll-contain px-5 py-5">
        {isError ? (
          <div
            className="border-border bg-surface-muted/40 mb-4 rounded-xl border p-4"
            role="alert"
          >
            <p className="text-foreground text-base">
              Comments could not be loaded.
            </p>
            <p className="text-text-muted mt-1 text-base leading-6">
              Try again to see the latest discussion.
            </p>
            <Button
              className="mt-3 text-base"
              color="tertiary"
              disabled={isFetching}
              onClick={() => {
                void refetch();
              }}
              size="sm"
              type="button"
              variant="outline"
            >
              {isFetching ? "Retrying…" : "Retry"}
            </Button>
          </div>
        ) : null}
        {isPending ? (
          <p className="text-text-muted py-5 text-base" role="status">
            Loading comments…
          </p>
        ) : null}
        {!isPending &&
        !isError &&
        visible.length === 0 &&
        !(draft && canComment) ? (
          <div className="flex min-h-60 flex-col items-center justify-center px-2 text-center">
            <Comment01Icon
              className="text-text-muted mb-4 size-8"
              strokeWidth={1.4}
            />
            <p className="text-foreground text-base font-medium">
              {filter === "open" ? "No open comments" : "No resolved comments"}
            </p>
            <p className="text-text-muted mt-2 max-w-72 text-base leading-6">
              {emptyDescription}
            </p>
          </div>
        ) : null}
        {visible.map((thread) => (
          <DocumentCommentThreadCard
            active={thread.id === activeThreadId}
            canComment={canComment}
            editor={editor}
            key={thread.id}
            mayaRuns={maya.data?.runs}
            mayaUnavailable={maya.isError}
            onCheckMaya={() => {
              void maya.refetch();
              void refetch();
            }}
            onThreadSelect={onThreadSelect}
            thread={thread}
          />
        ))}
      </div>
      {draft && canComment ? (
        <div className="border-border shrink-0 border-t p-4">
          <DocumentCommentComposer
            documentId={documentId}
            draft={draft}
            onCancel={() => {
              onDraftChange(null);
            }}
            onCreated={(thread) => {
              onDraftChange(null);
              setFilter("open");
              onThreadSelect?.(thread);
            }}
          />
        </div>
      ) : null}
    </aside>
  );
}
