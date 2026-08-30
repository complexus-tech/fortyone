"use client";

import { useEffect, useRef, useState } from "react";
import { useParams, useRouter, useSearchParams } from "next/navigation";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { ArrowLeftIcon, ClockIcon } from "icons";
import { Box, Button, Container, Divider, Flex, Skeleton, Text } from "ui";
import { BodyContainer } from "@/components/shared";
import { ConfirmDialog } from "@/components/ui";
import { useWorkspacePath } from "@/hooks";
import { useUserRole } from "@/hooks/role";
import {
  getAuthorPathByPortalSlug,
  getRequestPathBySlugs,
} from "@/modules/public-portal/utils";
import { useFeedbackPortals } from "@/modules/settings/workspace/feedback/hooks";
import { getStoryPath } from "@/shared/routing/story";
import { openDialogAfterMenuClose } from "@/utils/menu-dialog-state";
import { CloseTeamFeedbackDialog } from "./close-dialog";
import { FeedbackBanner } from "./components/feedback-banner";
import {
  FeedbackCommentComposer,
  FeedbackCommentThreads,
} from "./components/feedback-comments";
import { FeedbackProperties } from "./components/feedback-properties";
import {
  useTeamFeedbackItem,
  useTeamFeedbackPrivateAuthor,
} from "./hooks/use-feedback";
import { usePlanTeamFeedback } from "./hooks/use-plan-feedback";
import { useSetTeamFeedbackReadState } from "./hooks/use-read-state";
import { useTrashTeamFeedback } from "./hooks/use-trash";
import { useUpdateTeamFeedbackStatus } from "./hooks/use-update-status";
import { LinkFeedbackStoryDialog } from "./link-story-dialog";
import { MergeTeamFeedbackDialog } from "./merge-dialog";

const FeedbackDetailsSkeleton = () => (
  <Box className="h-full min-h-0">
    <Box className="notification-story-container flex h-full min-h-0 overflow-hidden">
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

const FeedbackDetailsError = ({ onRetry }: { onRetry: () => void }) => (
  <Box className="flex h-full items-center justify-center px-6">
    <Box>
      <Text align="center" className="mb-2" fontSize="xl">
        Couldn&apos;t load feedback
      </Text>
      <Text align="center" className="mb-4" color="muted">
        Check your connection and try again.
      </Text>
      <Flex justify="center">
        <Button color="tertiary" onClick={onRetry} size="sm" variant="outline">
          Try again
        </Button>
      </Flex>
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
  const { userRole } = useUserRole();
  const {
    data: feedback,
    isError,
    isPending,
    refetch,
  } = useTeamFeedbackItem(feedbackId);
  const { data: privateAuthor } = useTeamFeedbackPrivateAuthor(
    feedbackId,
    userRole === "admin",
  );
  const { data: feedbackPortals = [] } = useFeedbackPortals();
  const planFeedback = usePlanTeamFeedback();
  const { mutate: setReadState } = useSetTeamFeedbackReadState();
  const trashFeedback = useTrashTeamFeedback();
  const updateStatus = useUpdateTeamFeedbackStatus();
  const lastAutoReadFeedbackId = useRef<string | null>(null);
  const [isClosing, setIsClosing] = useState(false);
  const [isLinking, setIsLinking] = useState(false);
  const [isMerging, setIsMerging] = useState(false);
  const [isTrashing, setIsTrashing] = useState(false);

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
      <FeedbackDetailsError
        onRetry={() => {
          void refetch();
        }}
      />
    );
  }

  const feedbackTeamId = feedback.board.teamId || teamId;
  const linkedStory = feedback.storyLinks.find((link) => link.isPrimary);
  const linkedStoryHref = linkedStory
    ? withWorkspace(getStoryPath({ id: linkedStory.storyId }))
    : undefined;
  const feedbackPortal = feedbackPortals.find(
    (portal) => portal.id === feedback.portalId,
  );
  const portalHref =
    feedbackPortal?.isPublic === true
      ? getRequestPathBySlugs(feedbackPortal.slug, feedback.slug)
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
    router.push(withWorkspace(getStoryPath({ id: storyId })));
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
    <Box className="h-full min-h-0">
      <Box className="notification-story-container flex h-full min-h-0 overflow-hidden">
        <Box className="min-h-0 min-w-0 flex-1">
          <BodyContainer className="h-full min-h-0 overflow-y-auto pb-8">
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
                canManageTrash={userRole === "admin"}
                feedback={feedback}
                isPlanning={planFeedback.isPending}
                isTrashing={trashFeedback.isPending}
                onClose={() => {
                  openDialogAfterMenuClose(setIsClosing);
                }}
                onLink={() => {
                  openDialogAfterMenuClose(setIsLinking);
                }}
                onMerge={() => {
                  openDialogAfterMenuClose(setIsMerging);
                }}
                onOpenStory={() => {
                  if (linkedStory) openStory(linkedStory.storyId);
                }}
                onPlan={handlePlan}
                onReview={handleReview}
                onTrash={() => {
                  openDialogAfterMenuClose(setIsTrashing);
                }}
                portalHref={portalHref}
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
                  privateAuthor={privateAuthor}
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
                <FeedbackCommentThreads
                  comments={feedback.comments}
                  feedbackId={feedback.id}
                />
              </Box>
            </Container>
          </BodyContainer>
        </Box>
        <Box className="notification-story-sidebar from-sidebar/70 to-sidebar/40 border-border h-full min-h-0 w-(--story-sidebar-width) shrink-0 overflow-y-auto border-l-[0.5px] bg-linear-to-br pb-6">
          <FeedbackProperties
            authorProfileHref={authorProfileHref}
            feedback={feedback}
            linkedStoryHref={linkedStoryHref}
            privateAuthor={privateAuthor}
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
      <MergeTeamFeedbackDialog
        onClose={() => {
          setIsMerging(false);
        }}
        onMerged={(target) => {
          setIsMerging(false);
          router.push(
            withWorkspace(
              `/teams/${target.board.teamId}/feedback/${target.id}`,
            ),
          );
        }}
        open={isMerging}
        source={feedback}
      />
      <ConfirmDialog
        confirmText="Move to trash"
        description="This feedback will be hidden from team lists and the public portal. You can restore it from Trash for 30 days."
        isLoading={trashFeedback.isPending}
        isOpen={isTrashing}
        loadingText="Moving..."
        onCancel={() => {
          setIsTrashing(false);
        }}
        onClose={() => {
          setIsTrashing(false);
        }}
        onConfirm={() => {
          trashFeedback.mutate(feedback.id, {
            onSuccess: () => {
              setIsTrashing(false);
              router.push(listHref);
            },
          });
        }}
        title="Move this feedback to trash?"
      />
      {isClosing ? (
        <CloseTeamFeedbackDialog
          isLoading={updateStatus.isPending}
          onCancel={() => {
            setIsClosing(false);
          }}
          onConfirm={(publicExplanation) => {
            updateStatus.mutate(
              {
                feedbackId: feedback.id,
                payload: {
                  roadmapSummary: publicExplanation,
                  status: "closed",
                },
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
        />
      ) : null}
    </Box>
  );
};
