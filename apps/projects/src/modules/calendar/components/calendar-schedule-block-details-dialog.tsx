"use client";

import { format, isSameDay } from "date-fns";
import { ClockIcon, EditIcon, ExternalLinkIcon, StoryIcon } from "icons";
import { Box, Button, Dialog, Flex, Skeleton, Text, Tooltip } from "ui";
import { useWorkspacePath } from "@/hooks";
import type { CalendarScheduleBlock } from "@/lib/queries/calendar/types";
import { Options } from "@/modules/story/components/options";
import { useStoryById } from "@/modules/story/hooks/story";
import { getStoryPath } from "@/modules/story/utils/story-url";
import {
  getMayaCalendarBlockLabel,
  getMayaCalendarBlockReason,
  isCalendarScheduleBlockEditable,
} from "./calendar-block";

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
  const mayaLabel = getMayaCalendarBlockLabel(block);
  const mayaReason = getMayaCalendarBlockReason(block);

  if (!story) {
    return (
      <Box
        aria-label="Loading story details"
        className="space-y-5"
        role="status"
      >
        <Skeleton className="h-4 w-20" />
        <Skeleton className="h-8 w-4/5" />
        <Skeleton className="h-20 w-full" />
      </Box>
    );
  }

  const storyReference = story.teamCode
    ? `${story.teamCode}-${story.sequenceId}`
    : `#${story.sequenceId}`;
  const description = story.description.trim();

  return (
    <Box className="space-y-7">
      <Box>
        <Text
          className="text-text-muted tracking-wide"
          fontSize="sm"
          fontWeight="semibold"
          transform="uppercase"
        >
          {storyReference}
        </Text>
        <Text as="h2" className="mt-2 text-2xl leading-8" fontWeight="semibold">
          {story.title}
        </Text>
        {description ? (
          <Text className="text-text-muted mt-3 text-[0.9375rem] leading-6 whitespace-pre-wrap">
            {description}
          </Text>
        ) : null}
      </Box>

      <Box className="space-y-3">
        <Flex align="start" gap={3}>
          <ClockIcon className="text-text-muted mt-0.5 h-5 w-auto shrink-0" />
          <Text fontSize="md">{getCalendarScheduleBlockTimeLabel(block)}</Text>
        </Flex>
        {mayaLabel ? (
          <Flex align="start" gap={3}>
            <StoryIcon className="text-text-muted mt-0.5 h-5 w-auto shrink-0" />
            <Box>
              <Text fontSize="md" fontWeight="medium">
                {mayaLabel}
              </Text>
              {mayaReason ? (
                <Text className="mt-1 leading-5" color="muted" fontSize="md">
                  {mayaReason}
                </Text>
              ) : null}
            </Box>
          </Flex>
        ) : null}
      </Box>

      {block.hasConflict ? (
        <Box className="border-danger/40 bg-danger/[0.08] rounded-xl border px-4 py-3">
          <Text color="danger" fontSize="md" fontWeight="medium">
            This scheduled work overlaps a meeting.
          </Text>
        </Box>
      ) : null}

      <Box>
        <Text fontSize="md" fontWeight="semibold">
          Properties
        </Text>
        <Box className="mt-3">
          <Options
            isNotifications={false}
            storyId={story.id}
            variant="inline"
          />
        </Box>
      </Box>
    </Box>
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
  const { withWorkspace } = useWorkspacePath();
  const isEditable = Boolean(block && isCalendarScheduleBlockEditable(block));
  const storyHref =
    block?.storyId && block.blockType === "work"
      ? withWorkspace(getStoryPath({ id: block.storyCode || block.storyId }))
      : null;

  return (
    <Dialog onOpenChange={onOpenChange} open={Boolean(block)}>
      <Dialog.Content
        className="border-border/70 shadow-shadow bg-surface-elevated mr-6 mb-8 max-w-148 rounded-3xl border-[0.5px] shadow-2xl outline-none md:mt-auto"
        overlayClassName="justify-end bg-black/10"
      >
        <Dialog.Header className="flex min-h-16 items-center justify-end px-6 pr-16">
          <Dialog.Title className="sr-only">
            Scheduled story details
          </Dialog.Title>
          <Flex align="center" gap={1}>
            {isEditable && block ? (
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
            {storyHref ? (
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
            ) : null}
          </Flex>
        </Dialog.Header>
        <Dialog.Description className="sr-only">
          Details for the selected scheduled story.
        </Dialog.Description>

        <Dialog.Body className="h-[calc(100dvh-8rem)] max-h-[calc(100dvh-8rem)] px-6 pt-3 pb-8">
          {block ? (
            block.blockType === "work" && block.storyId ? (
              <ScheduledStoryDetails
                block={{ ...block, storyId: block.storyId }}
              />
            ) : (
              <Box className="space-y-4">
                <Text
                  as="h2"
                  className="text-2xl leading-8"
                  fontWeight="semibold"
                >
                  {block.title}
                </Text>
                <Flex align="start" gap={3}>
                  <ClockIcon className="text-text-muted mt-0.5 h-5 w-auto shrink-0" />
                  <Text fontSize="md">
                    {getCalendarScheduleBlockTimeLabel(block)}
                  </Text>
                </Flex>
              </Box>
            )
          ) : null}
        </Dialog.Body>
      </Dialog.Content>
    </Dialog>
  );
};
