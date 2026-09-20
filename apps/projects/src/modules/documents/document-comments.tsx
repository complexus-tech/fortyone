"use client";

import { useMemo, useState } from "react";
import type { Editor } from "@tiptap/react";
import { Avatar, Box, Button, Flex, Text, TextArea } from "ui";
import { CheckIcon, CloseIcon, Comment01Icon } from "icons";
import { cn } from "lib";
import { useSession } from "@/lib/auth/client";
import { useDocumentCommentMutations, useDocumentComments } from "./hooks";
import type { DocumentCommentSelection, DocumentCommentThread } from "./types";

const formatCommentTime = (value: string) => {
  const elapsed = Math.max(0, Date.now() - new Date(value).getTime());
  const minutes = Math.floor(elapsed / 60_000);
  if (minutes < 1) return "now";
  if (minutes < 60) return `${minutes}m`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days}d`;
  return new Intl.DateTimeFormat(undefined, {
    day: "numeric",
    month: "short",
  }).format(new Date(value));
};

const CommentComposer = ({
  authorAvatar,
  authorName,
  isPending,
  onCancel,
  onSubmit,
}: {
  authorAvatar?: string | null;
  authorName: string;
  isPending: boolean;
  onCancel: () => void;
  onSubmit: (body: string) => void;
}) => {
  const [body, setBody] = useState("");
  return (
    <Flex align="start" gap={3}>
      <Avatar
        className="ring-border-strong ring-1"
        name={authorName}
        size="md"
        src={authorAvatar}
      />
      <Box className="border-border bg-surface min-w-0 flex-1 rounded-lg border-[0.5px] p-3 shadow-sm">
        <TextArea
          aria-label="Comment"
          autoFocus
          className="min-h-20 resize-none rounded-none border-0 bg-transparent p-0 text-base leading-6 focus-visible:ring-0 dark:bg-transparent"
          onChange={(event) => {
            setBody(event.target.value);
          }}
          placeholder="Comment or type ‘/’ for commands…"
          value={body}
        />
        <Flex className="mt-2" gap={2} justify="end">
          <Button color="tertiary" onClick={onCancel} size="sm" variant="naked">
            Cancel
          </Button>
          <Button
            color="primary"
            disabled={body.trim() === "" || isPending}
            onClick={() => {
              onSubmit(body.trim());
            }}
            size="sm"
          >
            {isPending ? "Commenting…" : "Comment"}
          </Button>
        </Flex>
      </Box>
    </Flex>
  );
};

const CommentThread = ({
  editor,
  thread,
}: {
  editor: Editor | null;
  thread: DocumentCommentThread;
}) => {
  const { reply, resolve } = useDocumentCommentMutations(thread.documentId);
  const [isReplying, setIsReplying] = useState(false);
  const [replyBody, setReplyBody] = useState("");
  const resolved = Boolean(thread.resolvedAt);
  const locate = () => {
    if (!editor) return;
    const maximumPosition = editor.state.doc.content.size;
    if (
      thread.anchorStart < 1 ||
      thread.anchorEnd <= thread.anchorStart ||
      thread.anchorEnd > maximumPosition
    )
      return;
    editor
      .chain()
      .focus()
      .setTextSelection({ from: thread.anchorStart, to: thread.anchorEnd })
      .scrollIntoView()
      .run();
  };

  return (
    <Box
      className={cn(
        "border-border/70 border-b-[0.5px] py-4 first:pt-0 last:border-b-0 last:pb-0",
        resolved && "opacity-70",
      )}
    >
      <button
        className="text-warning hover:bg-warning/10 mb-3 flex items-center gap-1.5 rounded-md px-2 py-1.5 text-sm font-medium transition-colors"
        onClick={locate}
        type="button"
      >
        <Comment01Icon className="size-4 text-current" strokeWidth={2} />
        View in document
      </button>
      <Box className="space-y-4">
        {thread.comments.map((comment) => (
          <Flex align="start" gap={3} key={comment.id}>
            <Avatar
              className="ring-border-strong ring-1"
              name={comment.authorName || "Teammate"}
              size="md"
              src={comment.authorAvatar}
            />
            <Box className="min-w-0 flex-1">
              <Flex align="center" justify="between">
                <Text className="truncate" fontSize="md" fontWeight="medium">
                  {comment.authorName || "Teammate"}
                </Text>
                <Text
                  className="shrink-0"
                  color="muted"
                  fontSize="sm"
                  title={new Date(comment.createdAt).toLocaleString()}
                >
                  {formatCommentTime(comment.createdAt)}
                </Text>
              </Flex>
              <Text className="mt-1 text-base leading-6 whitespace-pre-wrap">
                {comment.body}
              </Text>
            </Box>
          </Flex>
        ))}
      </Box>
      {isReplying ? (
        <Box className="mt-4">
          <TextArea
            aria-label="Reply"
            className="border-border bg-surface min-h-20 resize-none rounded-lg border-[0.5px] px-3 py-2 text-base leading-6"
            onChange={(event) => {
              setReplyBody(event.target.value);
            }}
            placeholder="Reply…"
            value={replyBody}
          />
          <Flex className="mt-2" gap={2} justify="end">
            <Button
              color="tertiary"
              onClick={() => {
                setIsReplying(false);
                setReplyBody("");
              }}
              size="sm"
              variant="naked"
            >
              Cancel
            </Button>
            <Button
              disabled={replyBody.trim() === "" || reply.isPending}
              onClick={() => {
                reply.mutate({ body: replyBody.trim(), threadId: thread.id });
                setReplyBody("");
                setIsReplying(false);
              }}
              size="sm"
            >
              Reply
            </Button>
          </Flex>
        </Box>
      ) : (
        <Flex className="mt-4" gap={1}>
          {!resolved ? (
            <Button
              color="tertiary"
              onClick={() => {
                setIsReplying(true);
              }}
              size="sm"
              variant="naked"
            >
              Reply
            </Button>
          ) : null}
          <Button
            color="primary"
            leftIcon={resolved ? undefined : <CheckIcon className="size-4" />}
            onClick={() => {
              resolve.mutate({ resolved: !resolved, threadId: thread.id });
            }}
            size="sm"
            variant="naked"
          >
            {resolved ? "Reopen" : "Resolve"}
          </Button>
        </Flex>
      )}
    </Box>
  );
};

export const DocumentCommentsPanel = ({
  draft,
  documentId,
  editor,
  onClose,
  onDraftChange,
}: {
  draft: DocumentCommentSelection | null;
  documentId: string;
  editor: Editor | null;
  onClose: () => void;
  onDraftChange: (draft: DocumentCommentSelection | null) => void;
}) => {
  const { data: threads = [], isPending } = useDocumentComments(documentId);
  const { data: session } = useSession();
  const { create } = useDocumentCommentMutations(documentId);
  const [filter, setFilter] = useState<"open" | "resolved">("open");
  const openCount = threads.filter((thread) => !thread.resolvedAt).length;
  const resolvedCount = threads.length - openCount;
  const visibleThreads = useMemo(
    () =>
      threads.filter((thread) =>
        filter === "open" ? !thread.resolvedAt : Boolean(thread.resolvedAt),
      ),
    [filter, threads],
  );

  return (
    <Box
      aria-labelledby="document-comments-title"
      className="bg-surface/80 dark:bg-surface/80 flex h-full min-h-0 flex-col backdrop-blur-xl"
      role="complementary"
    >
      <Box className="border-border shrink-0 border-b-[0.5px] px-5 py-4">
        <Flex align="center" justify="between">
          <Flex align="center" gap={2}>
            <Comment01Icon className="size-4" strokeWidth={2} />
            <Text
              fontSize="xl"
              fontWeight="semibold"
              id="document-comments-title"
            >
              Comments
            </Text>
            {threads.length > 0 ? (
              <Text color="muted" fontSize="sm">
                {threads.length}
              </Text>
            ) : null}
          </Flex>
          <Button
            aria-label="Close comments"
            asIcon
            color="tertiary"
            onClick={onClose}
            size="sm"
            variant="naked"
          >
            <CloseIcon className="size-5" />
          </Button>
        </Flex>
        <Flex className="mt-3" gap={4}>
          {(
            [
              { count: openCount, label: "Open", value: "open" },
              { count: resolvedCount, label: "Resolved", value: "resolved" },
            ] as const
          ).map((item) => (
            <button
              className={cn(
                "text-text-muted relative pb-1.5 text-sm font-medium transition-colors",
                filter === item.value &&
                  "text-foreground after:bg-primary after:absolute after:inset-x-0 after:-bottom-[1px] after:h-0.5 after:rounded-full",
              )}
              key={item.value}
              onClick={() => {
                setFilter(item.value);
              }}
              type="button"
            >
              {item.label} {item.count}
            </button>
          ))}
        </Flex>
      </Box>
      <Box className="min-h-0 flex-1 overflow-y-auto px-5 py-4">
        {draft ? (
          <Box className="mb-4">
            <CommentComposer
              authorAvatar={session?.user.image}
              authorName={session?.user.name ?? "You"}
              isPending={create.isPending}
              onCancel={() => {
                onDraftChange(null);
              }}
              onSubmit={(body) => {
                create.mutate({
                  anchorEnd: draft.to,
                  anchorStart: draft.from,
                  body,
                  quote: draft.text,
                });
                onDraftChange(null);
              }}
            />
          </Box>
        ) : null}
        {isPending ? (
          <Text color="muted" fontSize="sm">
            Loading comments…
          </Text>
        ) : null}
        {!isPending && visibleThreads.length === 0 && !draft ? (
          <Flex
            align="center"
            className="min-h-64 text-center"
            direction="column"
            justify="center"
          >
            <Comment01Icon
              className="text-text-muted mb-3 size-8"
              strokeWidth={1.5}
            />
            <Text fontSize="md" fontWeight="medium">
              {filter === "open" ? "No open comments" : "No resolved comments"}
            </Text>
            {filter === "open" ? (
              <Text className="mt-1 max-w-64" color="muted" fontSize="md">
                Select text in the document and choose Comment from the toolbar.
              </Text>
            ) : null}
          </Flex>
        ) : null}
        <Box>
          {visibleThreads.map((thread) => (
            <CommentThread editor={editor} key={thread.id} thread={thread} />
          ))}
        </Box>
      </Box>
    </Box>
  );
};
