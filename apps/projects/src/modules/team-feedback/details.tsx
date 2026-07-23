"use client";

import type { ReactNode } from "react";
import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { useEditor } from "@tiptap/react";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { useParams, useRouter, useSearchParams } from "next/navigation";
import { toast } from "sonner";
import {
  CheckIcon,
  ArrowLeftIcon,
  ClockIcon,
  CloseIcon,
  CommentIcon,
  LinkIcon,
  MoreHorizontalIcon,
  ReplyIcon,
  RequestsIcon,
  StoryIcon,
  ThumbsUpIcon,
} from "icons";
import {
  Avatar,
  Box,
  Button,
  Container,
  Divider,
  Flex,
  Menu,
  Skeleton,
  Text,
  TextEditor,
  TimeAgo,
} from "ui";
import { BodyContainer } from "@/components/shared";
import { ConfirmDialog, Dot } from "@/components/ui";
import { useTerminology, useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { getAuthorPathByPortalSlug } from "@/modules/public-portal/utils";
import { openDialogAfterMenuClose } from "@/utils/menu-dialog-state";
import { Option } from "@/modules/story/components/options";
import {
  getStoryCommentEditorExtensions,
  serializeStoryCommentToGitHubMarkdown,
} from "@/modules/story/components/story-comment-editor";
import { useCreateTeamFeedbackComment } from "./hooks/use-create-comment";
import { usePlanTeamFeedback } from "./hooks/use-plan-feedback";
import { useSetTeamFeedbackReadState } from "./hooks/use-read-state";
import { useTeamFeedbackItem } from "./hooks/use-feedback";
import { useUpdateTeamFeedbackStatus } from "./hooks/use-update-status";
import { LinkFeedbackStoryDialog } from "./link-story-dialog";
import { FeedbackStatus } from "./status";
import type {
  TeamFeedbackComment,
  TeamFeedbackItem,
  TeamFeedbackStatus,
} from "./types";

const LINKED_STORY_TITLE_MAX_LENGTH = 16;

const truncateLinkedStoryTitle = (title: string) => {
  if (title.length <= LINKED_STORY_TITLE_MAX_LENGTH) return title;

  return `${title.slice(0, LINKED_STORY_TITLE_MAX_LENGTH)}...`;
};

const formatLinkedStoryTitle = (
  title: string | null | undefined,
  fallback: string,
) => (title ? truncateLinkedStoryTitle(title) : fallback);

const getStatusBannerCopy = (
  status: TeamFeedbackStatus,
  storyTerm: string,
): { primary: string; secondary: string } => {
  const copy: Record<
    TeamFeedbackStatus,
    { primary: string; secondary: string }
  > = {
    pending: {
      primary: "Feedback is ready for review",
      secondary: "Plan it when the team is ready to commit.",
    },
    reviewing: {
      primary: "Feedback is being reviewed",
      secondary: "Plan it to add the work to the roadmap.",
    },
    planned: {
      primary: "Feedback is planned",
      secondary: `Its linked ${storyTerm} now powers the roadmap status.`,
    },
    in_progress: {
      primary: "Feedback is in progress",
      secondary: `The linked ${storyTerm} is currently being worked on.`,
    },
    completed: {
      primary: "Feedback is completed",
      secondary: `The linked ${storyTerm} has been delivered.`,
    },
    closed: {
      primary: "Feedback is closed",
      secondary: "It is no longer in the team's active queue.",
    },
  };

  return copy[status];
};

const FeedbackBanner = ({
  feedback,
  isPlanning,
  onClose,
  onLink,
  onOpenStory,
  onPlan,
  onReview,
}: {
  feedback: TeamFeedbackItem;
  isPlanning: boolean;
  onClose: () => void;
  onLink: () => void;
  onOpenStory: () => void;
  onPlan: () => void;
  onReview: () => void;
}) => {
  const { getTermDisplay } = useTerminology();
  const storyTerm = getTermDisplay("storyTerm");
  const linkedStory = feedback.storyLinks.find((link) => link.isPrimary);
  const isLinked = Boolean(linkedStory);
  const canPlan = !isLinked && feedback.status !== "closed";
  const copy = getStatusBannerCopy(feedback.status, storyTerm);
  const linkedStoryTitle = linkedStory
    ? linkedStory.storyTitle || `Open linked ${storyTerm}`
    : null;

  return (
    <Box className="mb-6">
      <Flex
        align="center"
        className="border-primary/20 bg-primary/5 rounded-xl border px-4 py-3"
        gap={3}
        justify="between"
      >
        <Flex align="center" className="min-w-0 flex-1" gap={2}>
          {linkedStory ? (
            <StoryIcon className="text-primary h-5 shrink-0" />
          ) : (
            <RequestsIcon className="text-primary h-5 shrink-0" />
          )}
          {linkedStory ? (
            <button
              className="min-w-0 flex-1 text-left"
              onClick={onOpenStory}
              type="button"
            >
              <Text
                as="span"
                className="block min-w-0 truncate"
                color="primary"
                fontWeight="medium"
                title={linkedStory.storyTitle || undefined}
              >
                Linked {storyTerm}
                <span aria-hidden="true"> · </span>
                {linkedStoryTitle}
              </Text>
            </button>
          ) : (
            <Box className="min-w-0">
              <Text
                className="line-clamp-1"
                color="primary"
                fontWeight="medium"
              >
                {copy.primary}
              </Text>
              <Text className="line-clamp-1 text-[0.92rem]" color="muted">
                {copy.secondary}
              </Text>
            </Box>
          )}
        </Flex>
        <Flex align="center" className="shrink-0" gap={1}>
          {linkedStory ? (
            <button
              aria-label={`Open linked ${storyTerm}`}
              className="text-primary hover:text-primary/80 rounded-md p-1 transition"
              onClick={onOpenStory}
              type="button"
            >
              <LinkIcon className="text-current" />
            </button>
          ) : null}
          <Menu>
            <Menu.Button>
              <button
                aria-label="More feedback actions"
                className="text-primary hover:text-primary/80 rounded-md p-1 transition"
                type="button"
              >
                <MoreHorizontalIcon className="h-5 text-current" />
              </button>
            </Menu.Button>
            <Menu.Items align="end">
              <Menu.Group>
                <Menu.Item disabled={!canPlan || isPlanning} onSelect={onPlan}>
                  <CheckIcon className="text-icon h-5 w-auto" />
                  {isPlanning ? "Planning feedback..." : "Plan feedback"}
                </Menu.Item>
                {linkedStory ? (
                  <Menu.Item onSelect={onOpenStory}>
                    <LinkIcon className="h-5 w-auto" />
                    Open linked {storyTerm}
                  </Menu.Item>
                ) : null}
                <Menu.Item
                  disabled={
                    isLinked ||
                    feedback.status === "reviewing" ||
                    feedback.status === "closed"
                  }
                  onSelect={onReview}
                >
                  <RequestsIcon className="h-5 w-auto" />
                  Mark as reviewing
                </Menu.Item>
                <Menu.Item
                  disabled={isLinked || feedback.status === "closed"}
                  onSelect={onLink}
                >
                  <LinkIcon className="h-5 w-auto" />
                  Link existing {storyTerm}
                </Menu.Item>
              </Menu.Group>
              <Menu.Separator />
              <Menu.Group>
                <Menu.Item
                  className="text-danger"
                  disabled={isLinked || feedback.status === "closed"}
                  onSelect={onClose}
                >
                  <CloseIcon className="text-danger h-5 w-auto" />
                  Close feedback...
                </Menu.Item>
              </Menu.Group>
            </Menu.Items>
          </Menu>
        </Flex>
      </Flex>
    </Box>
  );
};

const compareCommentDates = (
  first: TeamFeedbackComment,
  second: TeamFeedbackComment,
) => new Date(first.createdAt).getTime() - new Date(second.createdAt).getTime();

const getCommentThreads = (comments: TeamFeedbackComment[]) => {
  const repliesByParent = new Map<string, TeamFeedbackComment[]>();
  const topLevelComments: TeamFeedbackComment[] = [];

  for (const comment of comments) {
    if (comment.parentId) {
      const replies = repliesByParent.get(comment.parentId) ?? [];
      replies.push(comment);
      repliesByParent.set(comment.parentId, replies);
    } else {
      topLevelComments.push(comment);
    }
  }

  return topLevelComments
    .sort((first, second) => compareCommentDates(second, first))
    .map((comment) => ({
      comment,
      replies: (repliesByParent.get(comment.id) ?? []).sort(
        compareCommentDates,
      ),
    }));
};

const FeedbackCommentComposer = ({
  feedbackId,
  onCancel,
  onSubmitted,
  parentId,
}: {
  feedbackId: string;
  onCancel?: () => void;
  onSubmitted?: () => void;
  parentId?: string;
}) => {
  const { data: session } = useSession();
  const createComment = useCreateTeamFeedbackComment(feedbackId);
  const editor = useEditor({
    content: "",
    editable: true,
    extensions: getStoryCommentEditorExtensions({
      placeholder: parentId ? "Reply to comment..." : "Leave a comment...",
    }),
    immediatelyRender: false,
  });

  const handleSubmit = () => {
    if (!editor || editor.isEmpty) {
      toast.error("Comment is required", {
        description: "Please enter a comment before submitting",
      });
      return;
    }

    const body = serializeStoryCommentToGitHubMarkdown(editor.getJSON());
    if (!body) return;

    editor.commands.clearContent();
    createComment.mutate(
      { body, parentId },
      {
        onError: () => {
          if (editor.isEmpty) {
            editor.commands.setContent(body);
          }
        },
        onSuccess: () => {
          onSubmitted?.();
        },
      },
    );
  };

  return (
    <Flex align="start" className={parentId ? "gap-2" : "mb-6 gap-2"}>
      <Box className="bg-surface flex aspect-square shrink-0 items-center rounded-full p-[0.3rem]">
        <Avatar
          name={session?.user.name ?? undefined}
          size="xs"
          src={session?.user.image ?? undefined}
        />
      </Box>
      <Box className="min-w-0 flex-1">
        <Flex
          className={
            parentId
              ? "border-border/40 bg-surface-muted/40 min-h-16 rounded-xl border px-3 pb-3"
              : "border-border/40 bg-surface-muted/40 min-h-24 rounded-2xl border px-4 pb-4"
          }
          direction="column"
          gap={2}
          justify="between"
        >
          <TextEditor
            aria-label={parentId ? "Reply" : "Comment"}
            className="prose-base prose-a:text-foreground leading-6 antialiased"
            editor={editor}
          />
          <Flex gap={1} justify="end">
            {onCancel ? (
              <Button
                color="tertiary"
                onClick={onCancel}
                size="sm"
                variant="naked"
              >
                Cancel
              </Button>
            ) : null}
            <Button
              color="tertiary"
              onClick={handleSubmit}
              size="sm"
              variant="outline"
            >
              {parentId ? "Reply" : "Comment"}
            </Button>
          </Flex>
        </Flex>
        {!parentId ? (
          <Text className="mt-1.5 text-[0.95rem]" color="muted">
            Comments are visible on the feedback portal.
          </Text>
        ) : null}
      </Box>
    </Flex>
  );
};

const CommentRow = ({
  comment,
  feedbackId,
  isReply = false,
  replies = [],
}: {
  comment: TeamFeedbackComment;
  feedbackId: string;
  isReply?: boolean;
  replies?: TeamFeedbackComment[];
}) => {
  const [isReplying, setIsReplying] = useState(false);

  return (
    <Box
      className={
        isReply
          ? "border-border ml-9 border-l-2 pt-1 pb-3 pl-2"
          : "relative pb-5"
      }
    >
      <Flex align="center" gap={1}>
        <Box className="bg-surface relative top-px flex aspect-square items-center rounded-full p-[0.3rem]">
          <Avatar
            className="relative top-0.5"
            name={comment.authorName}
            size="xs"
            src={comment.authorAvatar ?? undefined}
          />
        </Box>
        <Text className="ml-1 text-black dark:text-white">
          {comment.authorName}
        </Text>
        <Text className="mx-0.5 text-[0.95rem]" color="muted">
          ·
        </Text>
        <Text className="text-[0.95rem]" color="muted">
          <TimeAgo timestamp={comment.createdAt} />
        </Text>
      </Flex>
      <Box className="prose prose-stone dark:prose-invert prose-headings:font-semibold prose-a:text-primary prose-pre:bg-surface-muted prose-pre:text-foreground mt-0.5 ml-9 max-w-full leading-6">
        <Markdown remarkPlugins={[remarkGfm]}>{comment.body}</Markdown>
      </Box>
      {replies.length > 0 ? (
        <Box className="mt-2">
          {replies.map((reply) => (
            <CommentRow
              comment={reply}
              feedbackId={feedbackId}
              isReply
              key={reply.id}
            />
          ))}
        </Box>
      ) : null}
      {!isReply && !isReplying ? (
        <Button
          className="mt-2 ml-9 px-2"
          color="tertiary"
          leftIcon={<ReplyIcon className="h-4" />}
          onClick={() => {
            setIsReplying(true);
          }}
          size="sm"
          variant="naked"
        >
          Reply
        </Button>
      ) : null}
      {isReplying ? (
        <Box className="mt-3 ml-9">
          <FeedbackCommentComposer
            feedbackId={feedbackId}
            onCancel={() => {
              setIsReplying(false);
            }}
            onSubmitted={() => {
              setIsReplying(false);
            }}
            parentId={comment.id}
          />
        </Box>
      ) : null}
    </Box>
  );
};

const MetadataValue = ({ children }: { children: ReactNode }) => (
  <Flex align="center" className="min-w-0" gap={2}>
    {children}
  </Flex>
);

const FeedbackProperties = ({
  authorProfileHref,
  feedback,
  linkedStoryHref,
  variant = "sidebar",
}: {
  authorProfileHref?: string;
  feedback: TeamFeedbackItem;
  linkedStoryHref?: string;
  variant?: "inline" | "sidebar";
}) => {
  const { getTermDisplay } = useTerminology();
  const storyTerm = getTermDisplay("storyTerm");
  const linkedStory = feedback.storyLinks.find((link) => link.isPrimary);
  const isInline = variant === "inline";

  return (
    <Container
      className={
        isInline
          ? "text-text-muted px-0 pt-0 md:px-0"
          : "text-text-muted px-0.5 pt-4 md:px-6"
      }
    >
      {!isInline ? (
        <Box className="mb-0 grid grid-cols-[9rem_auto] items-center gap-3 md:mb-6">
          <Text className="hidden md:block" fontWeight="semibold">
            Properties
          </Text>
        </Box>
      ) : null}
      <Box className={isInline ? "flex flex-wrap gap-2" : undefined}>
        <Option
          className="my-5 md:my-6"
          isCompact={isInline}
          isNotifications={isInline}
          label="Author"
          value={
            <MetadataValue>
              <Avatar
                name={feedback.authorName}
                size="xs"
                src={feedback.authorAvatar ?? undefined}
              />
              {authorProfileHref ? (
                <Link
                  className="text-foreground hover:text-primary min-w-0 transition-colors"
                  href={authorProfileHref}
                >
                  <Text as="span" className="line-clamp-1">
                    {feedback.authorName}
                  </Text>
                </Link>
              ) : (
                <Text className="line-clamp-1">{feedback.authorName}</Text>
              )}
            </MetadataValue>
          }
        />
        <Option
          className="my-5 md:my-6"
          isCompact={isInline}
          isNotifications={isInline}
          label="Status"
          value={<FeedbackStatus status={feedback.status} />}
        />
        <Option
          className="my-5 md:my-6"
          isCompact={isInline}
          isNotifications={isInline}
          label="Board"
          value={
            <MetadataValue>
              <Dot className="size-3" color={feedback.board.color} />
              <Text className="line-clamp-1">{feedback.board.name}</Text>
            </MetadataValue>
          }
        />
        <Option
          className="my-5 md:my-6"
          isCompact={isInline}
          isNotifications={isInline}
          label="Votes"
          value={
            <MetadataValue>
              <ThumbsUpIcon className="h-4" />
              <Text>
                {feedback.voteCount}{" "}
                {feedback.voteCount === 1 ? "vote" : "votes"}
              </Text>
            </MetadataValue>
          }
        />
        <Option
          className="my-5 md:my-6"
          isCompact={isInline}
          isNotifications={isInline}
          label="Comments"
          value={
            <MetadataValue>
              <CommentIcon className="h-4" />
              <Text>
                {feedback.commentCount}{" "}
                {feedback.commentCount === 1 ? "comment" : "comments"}
              </Text>
            </MetadataValue>
          }
        />
        <Option
          className="my-5 md:my-6"
          isCompact={isInline}
          isNotifications={isInline}
          label="Submitted"
          value={
            <MetadataValue>
              <ClockIcon className="h-4" />
              <Text>
                <TimeAgo timestamp={feedback.createdAt} />
              </Text>
            </MetadataValue>
          }
        />
        <Option
          className="my-5 md:my-6"
          isCompact={isInline}
          isNotifications={isInline}
          label={`Linked ${storyTerm}`}
          value={
            linkedStory && linkedStoryHref ? (
              <Button
                aria-label={linkedStory.storyTitle || `Open ${storyTerm}`}
                className="max-w-full"
                color="tertiary"
                href={linkedStoryHref}
                leftIcon={<StoryIcon className="h-4 shrink-0" />}
                size="sm"
                variant="naked"
              >
                <span
                  className="truncate"
                  title={linkedStory.storyTitle || undefined}
                >
                  {formatLinkedStoryTitle(
                    linkedStory.storyTitle,
                    `Open ${storyTerm}`,
                  )}
                </span>
              </Button>
            ) : (
              <Text color="muted">Not linked</Text>
            )
          }
        />
      </Box>
      {!isInline && feedback.roadmapSummary ? (
        <Option
          className="my-5 items-start md:my-6"
          isNotifications={false}
          label="Roadmap note"
          value={
            <Text className="leading-5" color="muted">
              {feedback.roadmapSummary}
            </Text>
          }
        />
      ) : null}
    </Container>
  );
};

const FeedbackDetailsSkeleton = () => (
  <Box className="h-dvh">
    <Box className="notification-story-container flex h-full">
      <Box className="min-w-0 flex-1 px-8 py-7">
        <Skeleton className="mb-8 h-14 w-full rounded-xl" />
        <Skeleton className="mb-7 h-10 w-2/5 rounded" />
        <Skeleton className="mb-3 h-4 w-4/5 rounded" />
        <Skeleton className="h-4 w-3/5 rounded" />
      </Box>
      <Box className="border-border w-(--story-sidebar-width) shrink-0 border-l-[0.5px] px-6 py-7">
        <Skeleton className="mb-8 h-5 w-24 rounded" />
        {Array.from({ length: 5 }).map((_, index) => (
          <Skeleton className="mb-6 h-5 w-full rounded" key={index} />
        ))}
      </Box>
    </Box>
  </Box>
);

export const TeamFeedbackDetails = ({ feedbackId }: { feedbackId: string }) => {
  const { teamId, workspaceSlug } = useParams<{
    teamId: string;
    workspaceSlug: string;
  }>();
  const router = useRouter();
  const searchParams = useSearchParams();
  const { withWorkspace } = useWorkspacePath();
  const {
    data: feedback,
    isError,
    isPending,
    refetch,
  } = useTeamFeedbackItem(feedbackId);
  const planFeedback = usePlanTeamFeedback();
  const { mutate: setReadState } = useSetTeamFeedbackReadState();
  const updateStatus = useUpdateTeamFeedbackStatus();
  const lastAutoReadFeedbackId = useRef<string | null>(null);
  const [isClosing, setIsClosing] = useState(false);
  const [isLinking, setIsLinking] = useState(false);

  useEffect(() => {
    if (
      !feedback ||
      feedback.readAt ||
      lastAutoReadFeedbackId.current === feedback.id
    ) {
      return;
    }

    lastAutoReadFeedbackId.current = feedback.id;
    setReadState({ feedbackId: feedback.id, isRead: true });
  }, [feedback, setReadState]);

  if (isPending) return <FeedbackDetailsSkeleton />;

  if (isError) {
    return (
      <Box className="flex h-dvh items-center justify-center px-6">
        <Box>
          <Text align="center" className="mb-2" fontSize="xl">
            Couldn&apos;t load feedback
          </Text>
          <Text align="center" className="mb-4" color="muted">
            Check your connection and try again.
          </Text>
          <Flex justify="center">
            <Button
              color="tertiary"
              onClick={() => {
                void refetch();
              }}
              size="sm"
              variant="outline"
            >
              Try again
            </Button>
          </Flex>
        </Box>
      </Box>
    );
  }

  const feedbackTeamId = feedback.board.teamId || teamId;
  const linkedStory = feedback.storyLinks.find((link) => link.isPrimary);
  const linkedStoryHref = linkedStory
    ? withWorkspace(`/story/${linkedStory.storyId}`)
    : undefined;
  const status = searchParams.get("status");
  const search = searchParams.get("search");
  const feedbackListHref = withWorkspace(`/teams/${feedbackTeamId}/feedback`);
  const authorProfileHref =
    getAuthorPathByPortalSlug(workspaceSlug, feedback.authorId) ?? undefined;
  const listParams = new URLSearchParams();
  if (status) listParams.set("status", status);
  if (search) listParams.set("search", search);
  const listQuery = listParams.toString();
  const listHref = listQuery
    ? `${feedbackListHref}?${listQuery}`
    : feedbackListHref;

  const openStory = (storyId: string) => {
    router.push(withWorkspace(`/story/${storyId}`));
  };

  const handlePlan = () => {
    planFeedback.mutate(
      {
        feedbackId: feedback.id,
        payload: { teamId: feedbackTeamId },
      },
      {
        onSuccess: (response) => {
          if (!response.error?.message && response.data?.storyId) {
            openStory(response.data.storyId);
          }
        },
      },
    );
  };

  const handleReview = () => {
    updateStatus.mutate({
      feedbackId: feedback.id,
      payload: { status: "reviewing", roadmapSummary: null },
    });
  };

  return (
    <Box className="h-dvh">
      <Box className="notification-story-container flex h-full">
        <Box className="min-w-0 flex-1">
          <BodyContainer className="h-dvh overflow-y-auto pb-8">
            <Container className="max-w-7xl pt-7">
              <Button
                className="mb-4 md:hidden"
                color="tertiary"
                href={listHref}
                leftIcon={<ArrowLeftIcon className="h-4" />}
                size="sm"
                variant="naked"
              >
                Back to feedback
              </Button>
              <FeedbackBanner
                feedback={feedback}
                isPlanning={planFeedback.isPending}
                onClose={() => {
                  openDialogAfterMenuClose(setIsClosing);
                }}
                onLink={() => {
                  openDialogAfterMenuClose(setIsLinking);
                }}
                onOpenStory={() => {
                  if (linkedStory) openStory(linkedStory.storyId);
                }}
                onPlan={handlePlan}
                onReview={handleReview}
              />
              <Text
                as="h1"
                className="mb-7 text-3xl md:text-4xl"
                fontWeight="semibold"
              >
                {feedback.title}
              </Text>
              {feedback.description.trim() ? (
                <Box className="prose prose-stone dark:prose-invert prose-headings:font-semibold prose-a:text-primary prose-pre:bg-surface-muted prose-pre:text-foreground max-w-full text-lg leading-7">
                  <Markdown remarkPlugins={[remarkGfm]}>
                    {feedback.description}
                  </Markdown>
                </Box>
              ) : (
                <Text className="text-lg" color="muted">
                  No description was provided.
                </Text>
              )}
              <Box className="notification-story-inline-options mt-6 hidden">
                <FeedbackProperties
                  authorProfileHref={authorProfileHref}
                  feedback={feedback}
                  linkedStoryHref={linkedStoryHref}
                  variant="inline"
                />
              </Box>
              <Divider className="my-7" />
              <Box>
                <Text
                  as="h4"
                  className="mb-5 flex items-center gap-1.5"
                  fontWeight="medium"
                >
                  <ClockIcon className="relative -top-px" />
                  Activity feed
                </Text>
                <FeedbackCommentComposer feedbackId={feedback.id} />
                {getCommentThreads(feedback.comments).map(
                  ({ comment, replies }) => (
                    <CommentRow
                      comment={comment}
                      feedbackId={feedback.id}
                      key={comment.id}
                      replies={replies}
                    />
                  ),
                )}
              </Box>
            </Container>
          </BodyContainer>
        </Box>
        <Box className="notification-story-sidebar from-sidebar/70 to-sidebar/40 border-border w-(--story-sidebar-width) shrink-0 border-l-[0.5px] bg-linear-to-br md:h-dvh md:overflow-y-auto md:pb-6">
          <FeedbackProperties
            authorProfileHref={authorProfileHref}
            feedback={feedback}
            linkedStoryHref={linkedStoryHref}
          />
        </Box>
      </Box>
      <LinkFeedbackStoryDialog
        feedbackId={feedback.id}
        isOpen={isLinking}
        onLinked={openStory}
        onOpenChange={setIsLinking}
        teamId={feedbackTeamId}
      />
      <ConfirmDialog
        confirmText="Close feedback"
        description="Closing removes this item from the team's active feedback queue. It will remain available in the feedback portal."
        isLoading={updateStatus.isPending}
        isOpen={isClosing}
        loadingText="Closing..."
        onCancel={() => {
          setIsClosing(false);
        }}
        onClose={() => {
          setIsClosing(false);
        }}
        onConfirm={() => {
          updateStatus.mutate(
            {
              feedbackId: feedback.id,
              payload: { status: "closed", roadmapSummary: null },
            },
            {
              onSuccess: (response) => {
                if (!response.error?.message) {
                  setIsClosing(false);
                  router.push(listHref);
                }
              },
            },
          );
        }}
        title="Close this feedback?"
      />
    </Box>
  );
};
