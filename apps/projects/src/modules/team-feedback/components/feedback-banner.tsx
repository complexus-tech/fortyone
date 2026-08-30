import {
  CheckIcon,
  CloseIcon,
  DeleteIcon,
  LinkIcon,
  MoreHorizontalIcon,
  RequestsIcon,
  StoryIcon,
} from "icons";
import { Box, Button, Flex, Menu, Text } from "ui";
import { useTerminology } from "@/hooks/use-terminology-display";
import type { TeamFeedbackItem, TeamFeedbackStatus } from "../types";

const getStatusBannerCopy = (status: TeamFeedbackStatus): string => {
  const copy: Record<TeamFeedbackStatus, string> = {
    closed: "Feedback is closed",
    completed: "Feedback is completed",
    in_progress: "Feedback is in progress",
    pending: "Feedback is ready for review",
    planned: "Feedback is planned",
    reviewing: "Feedback is being reviewed",
  };

  return copy[status];
};

export const FeedbackBanner = ({
  canManageTrash,
  feedback,
  isPlanning,
  isTrashing,
  onClose,
  onLink,
  onMerge,
  onOpenStory,
  onPlan,
  onReview,
  onTrash,
  portalHref,
}: {
  canManageTrash: boolean;
  feedback: TeamFeedbackItem;
  isPlanning: boolean;
  isTrashing: boolean;
  onClose: () => void;
  onLink: () => void;
  onMerge: () => void;
  onOpenStory: () => void;
  onPlan: () => void;
  onReview: () => void;
  onTrash: () => void;
  portalHref?: string;
}) => {
  const { getTermDisplay } = useTerminology();
  const storyTerm = getTermDisplay("storyTerm");
  const linkedStory = feedback.storyLinks.find((link) => link.isPrimary);
  const isLinked = Boolean(linkedStory);
  const canPlan = !isLinked && feedback.status !== "closed";
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
                {getStatusBannerCopy(feedback.status)}
              </Text>
            </Box>
          )}
        </Flex>
        <Flex align="center" className="shrink-0" gap={1}>
          {portalHref ? (
            <Button
              aria-label="Open in feedback portal"
              asIcon
              className="text-primary hover:text-primary/80"
              color="tertiary"
              href={portalHref}
              leftIcon={<LinkIcon className="h-5 text-current" />}
              size="sm"
              target="_blank"
              title="Open in feedback portal"
              variant="naked"
            >
              <span className="sr-only">Open in feedback portal</span>
            </Button>
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
                {canManageTrash ? (
                  <Menu.Item disabled={isLinked} onSelect={onMerge}>
                    <LinkIcon className="h-5 w-auto" />
                    Merge feedback
                  </Menu.Item>
                ) : null}
                <Menu.Item
                  disabled={isLinked || feedback.status === "closed"}
                  onSelect={onClose}
                >
                  <CloseIcon className="h-5 w-auto" />
                  Close feedback...
                </Menu.Item>
              </Menu.Group>
              {canManageTrash ? (
                <>
                  <Menu.Separator />
                  <Menu.Group>
                    <Menu.Item
                      className="text-danger"
                      disabled={isLinked || isTrashing}
                      onSelect={onTrash}
                    >
                      <DeleteIcon className="text-danger h-5 w-auto" />
                      Move to trash...
                    </Menu.Item>
                  </Menu.Group>
                </>
              ) : null}
            </Menu.Items>
          </Menu>
        </Flex>
      </Flex>
    </Box>
  );
};
