"use client";

import { format, isSameDay } from "date-fns";
import { ClockIcon, ExternalLinkIcon, StoryIcon } from "icons";
import { Box, Button, Dialog, Flex, Text } from "ui";
import { useWorkspacePath } from "@/hooks";
import type { CalendarScheduleBlock } from "@/lib/queries/calendar/types";
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
  const mayaLabel = block ? getMayaCalendarBlockLabel(block) : null;
  const mayaReason = block ? getMayaCalendarBlockReason(block) : null;
  const storyHref =
    block?.storyId && block.blockType === "work"
      ? withWorkspace(getStoryPath({ id: block.storyCode || block.storyId }))
      : null;
  const storyReference = block?.storyCode || block?.teamCode;

  return (
    <Dialog onOpenChange={onOpenChange} open={Boolean(block)}>
      <Dialog.Content
        className="border-border/70 shadow-shadow bg-surface-elevated mr-6 mb-8 max-w-148 rounded-3xl border-[0.5px] shadow-2xl outline-none md:mt-auto"
        overlayClassName="justify-end bg-black/10"
      >
        <Dialog.Header className="border-border flex min-h-16 items-center border-b-[0.5px] px-6 pr-16">
          <Dialog.Title className="text-primary min-w-0 truncate text-lg font-semibold">
            {block?.title ?? "Scheduled story"}
          </Dialog.Title>
        </Dialog.Header>
        <Dialog.Description className="sr-only">
          Details for the selected scheduled story.
        </Dialog.Description>

        <Dialog.Body className="h-[calc(100dvh-8rem)] max-h-[calc(100dvh-8rem)] px-6 py-6">
          {block ? (
            <Box className="space-y-6">
              <Box>
                <Text
                  as="h2"
                  className="text-primary text-2xl"
                  fontWeight="semibold"
                >
                  {block.storyTitle ?? block.title}
                </Text>
                {storyReference ? (
                  <Text className="mt-2" color="muted" fontSize="md">
                    {storyReference}
                  </Text>
                ) : null}
              </Box>

              <Flex align="start" gap={3}>
                <ClockIcon className="text-text-muted mt-0.5 h-5 w-auto shrink-0" />
                <Text fontSize="md">
                  {getCalendarScheduleBlockTimeLabel(block)}
                </Text>
              </Flex>

              {mayaLabel ? (
                <Flex align="start" gap={3}>
                  <StoryIcon className="text-text-muted mt-0.5 h-5 w-auto shrink-0" />
                  <Box>
                    <Text fontSize="md" fontWeight="medium">
                      {mayaLabel}
                    </Text>
                    {mayaReason ? (
                      <Text className="mt-1" color="muted" fontSize="md">
                        {mayaReason}
                      </Text>
                    ) : null}
                  </Box>
                </Flex>
              ) : null}

              {block.hasConflict ? (
                <Box className="border-danger/40 bg-danger/[0.08] rounded-xl border px-4 py-3">
                  <Text color="danger" fontSize="md" fontWeight="medium">
                    This scheduled work overlaps a meeting.
                  </Text>
                </Box>
              ) : null}

              <Flex align="center" className="flex-wrap" gap={2}>
                {storyHref ? (
                  <Button
                    color="invert"
                    href={storyHref}
                    leftIcon={<ExternalLinkIcon className="h-4 w-auto" />}
                  >
                    Open story
                  </Button>
                ) : null}
                {isEditable ? (
                  <Button
                    color="tertiary"
                    onClick={() => {
                      onEdit(block);
                    }}
                    variant="outline"
                  >
                    Edit schedule
                  </Button>
                ) : null}
              </Flex>
            </Box>
          ) : null}
        </Dialog.Body>
      </Dialog.Content>
    </Dialog>
  );
};
