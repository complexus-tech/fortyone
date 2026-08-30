"use client";

import { useRef } from "react";
import { useEditor } from "@tiptap/react";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { toast } from "sonner";
import { cn } from "lib";
import {
  Avatar,
  Box,
  Button,
  Flex,
  Skeleton,
  Text,
  TextEditor,
  TimeAgo,
} from "ui";
import { useSession } from "@/lib/auth/client";
import { getCommentEditorExtensions } from "@/lib/tiptap/comment-editor";
import { serializeCommentToGitHubMarkdown } from "@/lib/tiptap/comment-markdown";
import { useCreateSlackAccountLinkSession } from "@/lib/hooks/slack";
import { getCommentDeliveryLabel } from "./delivery-status";
import { usePostIntegrationRequestComment } from "./hooks/use-post-request-comment";
import { useIntegrationRequestThread } from "./hooks/use-request-thread";

const SlackCommentComposer = ({
  compact,
  requestId,
}: {
  compact: boolean;
  requestId: string;
}) => {
  const { data: session } = useSession();
  const postComment = usePostIntegrationRequestComment(requestId);
  const accountLinkSession = useCreateSlackAccountLinkSession();
  const retryRef = useRef<{ body: string; idempotencyKey: string } | null>(
    null,
  );
  const editor = useEditor({
    content: "",
    editable: !postComment.isPending,
    extensions: getCommentEditorExtensions({
      placeholder: "Reply to the Slack thread...",
    }),
    immediatelyRender: false,
  });

  const handleSubmit = async () => {
    if (!editor || editor.isEmpty) return;

    const body = serializeCommentToGitHubMarkdown(editor.getJSON());
    if (!body) return;

    const previousAttempt = retryRef.current;
    const idempotencyKey =
      previousAttempt?.body === body
        ? previousAttempt.idempotencyKey
        : globalThis.crypto.randomUUID();
    retryRef.current = { body, idempotencyKey };

    const linkResponse = await accountLinkSession.mutateAsync(
      window.location.href,
    );
    if (linkResponse.error?.message) {
      toast.error("Slack", { description: linkResponse.error.message });
      return;
    }
    if (!linkResponse.data?.linked && linkResponse.data?.canLink) {
      const installUrl = linkResponse.data.installUrl;
      if (!installUrl) {
        toast.error("Slack", {
          description: "FortyOne could not start Slack account linking.",
        });
        return;
      }
      toast.info("Connect your Slack account, then post again.");
      window.location.assign(installUrl);
      return;
    }

    editor.commands.clearContent();
    postComment.mutate(
      { body, idempotencyKey },
      {
        onSuccess: () => {
          retryRef.current = null;
        },
        onError: () => {
          if (editor.isEmpty) editor.commands.setContent(body);
        },
      },
    );
  };

  return (
    <Flex align="start" className={cn("gap-2", compact ? "mb-4" : "mb-6")}>
      <Box className="bg-surface flex aspect-square shrink-0 items-center rounded-full p-[0.3rem]">
        <Avatar
          name={session?.user.name ?? undefined}
          size="xs"
          src={session?.user.image ?? undefined}
        />
      </Box>
      <Flex
        className={cn(
          "border-border/40 bg-surface-muted/40 min-w-0 flex-1 border",
          compact
            ? "min-h-20 rounded-xl px-3 pb-3"
            : "min-h-24 rounded-2xl px-4 pb-4",
        )}
        direction="column"
        gap={2}
        justify="between"
      >
        <TextEditor
          aria-label="Reply to Slack"
          className="prose-base prose-a:text-foreground leading-6 antialiased"
          editor={editor}
        />
        <Flex justify="end">
          <Button
            color="tertiary"
            disabled={postComment.isPending || accountLinkSession.isPending}
            loading={postComment.isPending || accountLinkSession.isPending}
            loadingText={
              accountLinkSession.isPending ? "Connecting..." : "Posting..."
            }
            onClick={handleSubmit}
            size="sm"
            variant="outline"
          >
            Post to Slack
          </Button>
        </Flex>
      </Flex>
    </Flex>
  );
};

export const IntegrationRequestThreadActivity = ({
  className,
  compact = false,
  requestId,
}: {
  className?: string;
  compact?: boolean;
  requestId: string;
}) => {
  const { data, isError, isLoading } = useIntegrationRequestThread(requestId);

  if (isLoading) {
    return <Skeleton className="h-28 w-full rounded-2xl" />;
  }
  if (isError || !data) {
    return (
      <Text color="muted">
        This older request is not connected to a Slack thread.
      </Text>
    );
  }

  return (
    <Box className={className}>
      <SlackCommentComposer compact={compact} requestId={requestId} />
      {data.comments.length > 0 ? (
        <Box
          aria-label="Slack thread history"
          className={cn("space-y-5", {
            "max-h-80 overflow-y-auto pr-1": compact,
          })}
        >
          {data.comments.map((comment) => (
            <Box key={comment.id}>
              <Flex align="center" gap={1}>
                <Avatar
                  name={comment.authorName}
                  size="xs"
                  src={comment.authorAvatar}
                />
                <Text className="ml-1" fontWeight="medium">
                  {comment.authorName}
                </Text>
                <Text color="muted">·</Text>
                <Text color="muted">
                  <TimeAgo timestamp={comment.createdAt} />
                </Text>
                {comment.direction === "outbound" &&
                getCommentDeliveryLabel(comment.deliveryStatus) ? (
                  <Text
                    color={
                      comment.deliveryStatus === "failed" ? "danger" : "muted"
                    }
                  >
                    · {getCommentDeliveryLabel(comment.deliveryStatus)}
                  </Text>
                ) : null}
              </Flex>
              <Box className="prose prose-stone dark:prose-invert prose-a:text-primary mt-1 ml-9 max-w-full leading-6">
                <Markdown remarkPlugins={[remarkGfm]}>{comment.body}</Markdown>
              </Box>
            </Box>
          ))}
        </Box>
      ) : (
        <Text className="ml-10" color="muted">
          No Slack replies yet.
        </Text>
      )}
    </Box>
  );
};
