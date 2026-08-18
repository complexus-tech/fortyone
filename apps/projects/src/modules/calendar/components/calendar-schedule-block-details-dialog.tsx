"use client";

import { format, isSameDay } from "date-fns";
import dynamic from "next/dynamic";
import { ClockIcon, CopyIcon, EditIcon, ExternalLinkIcon } from "icons";
import { Box, Button, Dialog, Flex, Skeleton, Text, Tooltip } from "ui";
import { toast } from "sonner";
import { useCopyToClipboard, useWorkspacePath } from "@/hooks";
import { useIsAdminOrOwner } from "@/hooks/owner";
import type { CalendarScheduleBlock } from "@/lib/queries/calendar/types";
import { StoryActionsMenu } from "@/modules/story/components/story-actions-menu";
import { useStoryById } from "@/modules/story/hooks/story";
import {
  getStoryPath,
  getStoryReference,
} from "@/modules/story/utils/story-url";
import { isCalendarScheduleBlockEditable } from "./calendar-block";

const CalendarStoryMainDetails = dynamic(
  () =>
    import("@/modules/story/components/main-details").then(
      (module) => module.MainDetails,
    ),
  {
    loading: () => (
      <Box
        aria-label="Loading story details"
        className="space-y-5 px-6"
        role="status"
      >
        <Skeleton className="h-8 w-4/5" />
        <Skeleton className="h-20 w-full" />
      </Box>
    ),
    ssr: false,
  },
);

export const getCalendarScheduleBlockTimeLabel = (
  block: Pick<CalendarScheduleBlock, "startAt" | "endAt">,
) => {
  const start = new Date(block.startAt);
  const end = new Date(block.endAt);
  if (isSameDay(start, end)) {
    return `${format(start, "EEEE, MMMM d")} · ${format(start, "h:mm a")} – ${format(end, "h:mm a")}`;
  }
  return `${format(start, "MMM d, h:mm a")} – ${format(end, "MMM d, h:mm a")}`;
};

const ScheduledStoryDetails = ({
  block,
}: {
  block: CalendarScheduleBlock & { storyId: string };
}) => {
  const { data: story } = useStoryById(block.storyId);

  if (!story) {
    return (
      <Box
        aria-label="Loading story details"
        className="space-y-5 px-6"
        role="status"
      >
        <Skeleton className="h-4 w-20" />
        <Skeleton className="h-8 w-4/5" />
        <Skeleton className="h-20 w-full" />
      </Box>
    );
  }

  return (
    <CalendarStoryMainDetails
      inlineProperties
      isDialog
      isNotifications={false}
      storyId={story.id}
    />
  );
};

const ScheduledStoryHeader = ({
  block,
  onEdit,
}: {
  block: CalendarScheduleBlock & { storyId: string };
  onEdit: (block: CalendarScheduleBlock) => void;
}) => {
  const { data: story } = useStoryById(block.storyId);
  const { isAdminOrOwner } = useIsAdminOrOwner(story?.reporterId);
  const { withWorkspace } = useWorkspacePath();
  const [, copyText] = useCopyToClipboard();
  const isEditable = isCalendarScheduleBlockEditable(block);
  const storyReference = story
    ? getStoryReference(story)
    : block.storyCode || block.teamCode || "Story";
  const storyHref = withWorkspace(
    getStoryPath({
      id: story?.id ?? block.storyId,
      sequenceId: story?.sequenceId,
      teamCode: story?.teamCode ?? block.teamCode,
    }),
  );

  const copyStoryLink = () => {
    void copyText(`${window.location.origin}${storyHref}`).then((copied) => {
      if (copied) {
        toast.success("Link copied to clipboard");
      }
    });
  };

  return (
    <Flex align="center" className="w-full" justify="between">
      <Text
        className="text-text-muted tracking-wide"
        fontSize="md"
        fontWeight="semibold"
        transform="uppercase"
      >
        {storyReference}
      </Text>
      <Flex align="center" gap={1}>
        {isEditable ? (
          <Tooltip side="bottom" title="Edit schedule">
            <Button
              asIcon
              color="tertiary"
              leftIcon={<EditIcon className="h-5 w-auto" />}
              onClick={() => {
                onEdit(block);
              }}
              variant="naked"
            >
              <span className="sr-only">Edit schedule</span>
            </Button>
          </Tooltip>
        ) : null}
        <Tooltip side="bottom" title="Copy story link">
          <Button
            asIcon
            color="tertiary"
            leftIcon={<CopyIcon className="h-5 w-auto" />}
            onClick={copyStoryLink}
            variant="naked"
          >
            <span className="sr-only">Copy story link</span>
          </Button>
        </Tooltip>
        <Tooltip side="bottom" title="Open story in a new tab">
          <Button
            asIcon
            color="tertiary"
            href={storyHref}
            leftIcon={<ExternalLinkIcon className="h-5 w-auto" />}
            rel="noreferrer"
            target="_blank"
            variant="naked"
          >
            <span className="sr-only">Open story in a new tab</span>
          </Button>
        </Tooltip>
        <StoryActionsMenu
          align="end"
          isAdminOrOwner={isAdminOrOwner}
          storyId={block.storyId}
        />
      </Flex>
    </Flex>
  );
};

export const CalendarScheduleBlockDetailsDialog = ({
  block,
  onEdit,
  onOpenChange,
}: {
  block: CalendarScheduleBlock | null;
  onEdit: (block: CalendarScheduleBlock) => void;
  onOpenChange: (open: boolean) => void;
}) => {
  const isScheduledStory = Boolean(
    block?.blockType === "work" && block.storyId,
  );
  let bodyContent = null;
  if (block) {
    if (isScheduledStory && block.storyId) {
      bodyContent = (
        <ScheduledStoryDetails block={{ ...block, storyId: block.storyId }} />
      );
    } else {
      bodyContent = (
        <Box className="space-y-4 px-6 py-3">
          <Text as="h2" className="text-2xl leading-8" fontWeight="semibold">
            {block.title}
          </Text>
          <Flex align="start" gap={3}>
            <ClockIcon className="text-text-muted mt-0.5 h-5 w-auto shrink-0" />
            <Text fontSize="md">
              {getCalendarScheduleBlockTimeLabel(block)}
            </Text>
          </Flex>
        </Box>
      );
    }
  }

  return (
    <Dialog onOpenChange={onOpenChange} open={Boolean(block)}>
      <Dialog.Content
        className="border-border/70 shadow-shadow bg-surface-elevated mr-6 mb-8 max-w-148 rounded-3xl border-[0.5px] shadow-2xl outline-none md:mt-auto"
        overlayClassName="justify-end bg-black/10"
      >
        <Dialog.Header className="flex min-h-16 items-center px-6 pr-16">
          <Dialog.Title className="sr-only">
            Scheduled story details
          </Dialog.Title>
          {isScheduledStory && block?.storyId ? (
            <ScheduledStoryHeader
              block={{ ...block, storyId: block.storyId }}
              onEdit={onEdit}
            />
          ) : null}
        </Dialog.Header>
        <Dialog.Description className="sr-only">
          Details for the selected scheduled story.
        </Dialog.Description>

        <Dialog.Body className="h-[calc(100dvh-8rem)] max-h-[calc(100dvh-8rem)] overflow-hidden px-0 py-0">
          {bodyContent}
        </Dialog.Body>
      </Dialog.Content>
    </Dialog>
  );
};
