"use client";

import { useState } from "react";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { Avatar, Box, Button, Flex, Skeleton, Text, TimeAgo } from "ui";
import { NewTabIcon } from "icons";
import { useSession } from "@/lib/auth/client";
import { useStoryGitHubComments, usePostGitHubComment } from "@/lib/hooks/github";
import type { GitHubComment } from "@/modules/settings/workspace/integrations/github/types";

const CommentRow = ({ comment }: { comment: GitHubComment }) => (
  <Box className="relative pb-4">
    <Flex align="center" gap={1}>
      <Box className="bg-surface relative top-px flex aspect-square items-center rounded-full p-[0.3rem]">
        <Avatar
          name={comment.userLogin}
          size="xs"
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
        className="text-muted hover:text-foreground ml-1 rounded-md p-0.5 transition"
        href={comment.htmlUrl}
        rel="noopener noreferrer"
        target="_blank"
        title="View on GitHub"
      >
        <NewTabIcon className="h-3.5 w-3.5" />
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
  const [value, setValue] = useState("");

  const handleSubmit = () => {
    const trimmed = value.trim();
    if (!trimmed) return;
    postComment(
      { storyId, body: trimmed },
      { onSuccess: () => setValue("") },
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
        <textarea
          className="bg-transparent mt-3 w-full resize-none leading-6 outline-none placeholder:text-gray-400"
          disabled={isPending}
          onChange={(e) => setValue(e.target.value)}
          placeholder="Comment on GitHub..."
          rows={3}
          value={value}
        />
        <Flex gap={2} justify="end">
          <Button
            className="px-3"
            color="tertiary"
            disabled={isPending || !value.trim()}
            onClick={handleSubmit}
            size="sm"
            variant="outline"
          >
            {isPending ? "Posting..." : "Comment"}
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
      ) : (
        <Text color="muted">No GitHub comments yet</Text>
      )}
    </Box>
  );
};
