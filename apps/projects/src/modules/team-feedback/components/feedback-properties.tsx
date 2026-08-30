import type { ReactNode } from "react";
import Link from "next/link";
import {
  ClockIcon,
  CommentIcon,
  StoryIcon,
  ThumbsDownIcon,
  ThumbsUpIcon,
} from "icons";
import { Avatar, Box, Button, Container, Flex, Text, TimeAgo } from "ui";
import { Dot } from "@/components/ui/dot";
import { PropertyOption } from "@/components/ui/property-option";
import { useTerminology } from "@/hooks/use-terminology-display";
import { FeedbackStatus } from "../status";
import type { TeamFeedbackItem, TeamFeedbackPrivateAuthor } from "../types";

const LINKED_STORY_TITLE_MAX_LENGTH = 16;

const formatLinkedStoryTitle = (
  title: string | null | undefined,
  fallback: string,
) => {
  if (!title) return fallback;
  if (title.length <= LINKED_STORY_TITLE_MAX_LENGTH) return title;

  return `${title.slice(0, LINKED_STORY_TITLE_MAX_LENGTH)}...`;
};

const MetadataValue = ({ children }: { children: ReactNode }) => (
  <Flex align="center" className="min-w-0" gap={2}>
    {children}
  </Flex>
);

export const FeedbackProperties = ({
  authorProfileHref,
  feedback,
  linkedStoryHref,
  privateAuthor,
  variant = "sidebar",
}: {
  authorProfileHref?: string;
  feedback: TeamFeedbackItem;
  linkedStoryHref?: string;
  privateAuthor?: TeamFeedbackPrivateAuthor;
  variant?: "inline" | "sidebar";
}) => {
  const { getTermDisplay } = useTerminology();
  const storyTerm = getTermDisplay("storyTerm");
  const linkedStory = feedback.storyLinks.find((link) => link.isPrimary);
  const isInline = variant === "inline";
  const authorName = privateAuthor?.displayName || feedback.authorName;
  const authorAvatar = privateAuthor?.avatarUrl ?? feedback.authorAvatar;
  const showPrivateIdentity = Boolean(privateAuthor?.publicMasked);

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
        <PropertyOption
          className="my-5 md:my-6"
          isCompact={isInline}
          isNotifications={isInline}
          label="Author"
          value={
            <MetadataValue>
              <Avatar
                name={authorName}
                size="xs"
                src={authorAvatar ?? undefined}
              />
              {authorProfileHref && !showPrivateIdentity ? (
                <Link
                  className="text-foreground hover:text-primary min-w-0 transition-colors"
                  href={authorProfileHref}
                >
                  <Text as="span" className="line-clamp-1">
                    {authorName}
                  </Text>
                </Link>
              ) : (
                <Box className="min-w-0">
                  <Text className="line-clamp-1">{authorName}</Text>
                  {showPrivateIdentity ? (
                    <Text className="text-xs" color="muted">
                      Hidden publicly
                    </Text>
                  ) : null}
                </Box>
              )}
            </MetadataValue>
          }
        />
        {privateAuthor?.email &&
        (privateAuthor.kind === "verified_guest" ||
          privateAuthor.kind === "external") ? (
          <PropertyOption
            className="my-5 md:my-6"
            isCompact={isInline}
            isNotifications={isInline}
            label="Contact"
            value={
              <a
                className="text-foreground hover:text-primary min-w-0 truncate transition-colors"
                href={`mailto:${privateAuthor.email}`}
              >
                {privateAuthor.email}
              </a>
            }
          />
        ) : null}
        <PropertyOption
          className="my-5 md:my-6"
          isCompact={isInline}
          isNotifications={isInline}
          label="Status"
          value={<FeedbackStatus status={feedback.status} />}
        />
        <PropertyOption
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
        <PropertyOption
          className="my-5 md:my-6"
          isCompact={isInline}
          isNotifications={isInline}
          label="Upvotes"
          value={
            <MetadataValue>
              <ThumbsUpIcon className="h-4" />
              <Text>
                {feedback.upvoteCount}{" "}
                {feedback.upvoteCount === 1 ? "upvote" : "upvotes"}
              </Text>
            </MetadataValue>
          }
        />
        <PropertyOption
          className="my-5 md:my-6"
          isCompact={isInline}
          isNotifications={isInline}
          label="Downvotes"
          value={
            <MetadataValue>
              <ThumbsDownIcon className="h-4" />
              <Text>
                {feedback.downvoteCount}{" "}
                {feedback.downvoteCount === 1 ? "downvote" : "downvotes"}
              </Text>
            </MetadataValue>
          }
        />
        <PropertyOption
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
        <PropertyOption
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
        <PropertyOption
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
        <PropertyOption
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
