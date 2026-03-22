"use client";

import { useEditor } from "@tiptap/react";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
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
import { LinkIcon, NewTabIcon } from "icons";
import { useSession } from "@/lib/auth/client";
import {
  useStoryGitHubComments,
  usePostGitHubComment,
} from "@/lib/hooks/github";
import type { GitHubComment } from "@/modules/settings/workspace/integrations/github/types";
import {
  getStoryCommentEditorExtensions,
  serializeStoryCommentToGitHubMarkdown,
} from "./story-comment-editor";

const CommentRow = ({ comment }: { comment: GitHubComment }) => (
  <Box className="relative pb-4">
    <Flex align="center" gap={1}>
      <Box className="bg-surface relative top-px flex aspect-square items-center rounded-full p-[0.3rem]">
        <Avatar
          name={comment.userLogin}
          size="xs"
          className="relative top-0.5"
          src={comment.userAvatar || undefined}
        />
      </Box>
      <Text className="relative top-0.5 ml-1 text-black dark:text-white">
        {comment.userLogin}
      </Text>
      <Text className="mx-0.5 text-[0.95rem]" color="muted">
        ·
      </Text>
      <Text className="text-[0.95rem]" color="muted">
        <TimeAgo timestamp={comment.createdAt} />
      </Text>
      <a
        className="text-muted hover:text-foreground rounded-md p-0.5 transition"
        href={comment.htmlUrl}
        rel="noopener noreferrer"
        target="_blank"
        title="View on GitHub"
      >
        <LinkIcon className="h-4" />
      </a>
    </Flex>
    <Box className="prose prose-stone dark:prose-invert prose-headings:font-semibold prose-a:text-primary prose-pre:bg-surface-muted prose-pre:text-foreground mt-0.5 ml-9 max-w-full leading-6">
      <Markdown remarkPlugins={[remarkGfm]}>{comment.body}</Markdown>
    </Box>
  </Box>
);

const CommentsSkeleton = () => (
  <Box className="space-y-4">
    {Array.from({ length: 2 }).map((_, i) => (
      <Box className="flex gap-3" key={i}>
        <Skeleton className="h-8 w-8 rounded-full" />
        <Box className="flex-1 space-y-2">
          <Skeleton className="h-4 w-32" />
          <Skeleton className="h-12 w-full" />
        </Box>
      </Box>
    ))}
  </Box>
);

const GitHubCommentInput = ({ storyId }: { storyId: string }) => {
  const { data: session } = useSession();
  const { mutate: postComment, isPending } = usePostGitHubComment();
  const editor = useEditor({
    content: "",
    editable: !isPending,
    extensions: getStoryCommentEditorExtensions({
      placeholder: "Leave a comment...",
    }),
    immediatelyRender: false,
  });

  const handleSubmit = () => {
    if (!editor || editor.isEmpty) {
      return;
    }

    const body = serializeStoryCommentToGitHubMarkdown(editor.getJSON());

    if (!body) {
      return;
    }

    postComment(
      { storyId, body },
      {
        onSuccess: () => {
          editor.commands.clearContent();
        },
      },
    );
  };

  return (
    <Flex align="start" className="mb-3">
      <Box className="bg-surface z-1 flex aspect-square items-center rounded-full p-[0.3rem]">
        <Avatar
          name={session?.user?.name ?? undefined}
          size="xs"
          src={session?.user?.image ?? undefined}
        />
      </Box>
      <Flex
        className="border-border/40 bg-surface-muted/40 ml-1 min-h-24 w-full rounded-2xl border px-4 pb-4"
        direction="column"
        gap={2}
        justify="between"
      >
        <TextEditor
          className="prose-base prose-a:text-foreground leading-6 antialiased"
          editor={editor}
        />
        <Flex gap={2} justify="end">
          <Button
            className="px-3"
            color="tertiary"
            disabled={isPending}
            onClick={handleSubmit}
            size="sm"
            variant="outline"
          >
            Comment
          </Button>
        </Flex>
      </Flex>
    </Flex>
  );
};

export const GitHubCommentsPanel = ({
  storyId,
  hasLinks,
}: {
  storyId: string;
  hasLinks: boolean;
}) => {
  const { data: comments, isLoading } = useStoryGitHubComments(
    storyId,
    hasLinks,
  );

  return (
    <Box>
      <GitHubCommentInput storyId={storyId} />
      {isLoading ? (
        <CommentsSkeleton />
      ) : comments && comments.length > 0 ? (
        <Box>
          {comments.map((comment) => (
            <CommentRow comment={comment} key={comment.id} />
          ))}
        </Box>
      ) : null}
    </Box>
  );
};
