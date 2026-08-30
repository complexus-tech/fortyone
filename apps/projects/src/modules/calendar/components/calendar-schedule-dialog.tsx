import { useState } from "react";
import { addDays, addHours } from "date-fns";
import {
  ArrowDown2Icon,
  CheckIcon,
  ClockIcon,
  CloseIcon,
  DeleteIcon,
  SearchIcon,
  StoryIcon,
  TimeScheduleIcon,
} from "icons";
import { cn } from "lib";
import { Box, Button, Dialog, Flex, Input, Popover, Text } from "ui";
import { SmartDateTimeRangeInput } from "@/components/ui/smart-datetime-input";
import {
  useCreateCalendarScheduleBlock,
  useDeleteCalendarScheduleBlock,
  useUpdateCalendarScheduleBlock,
} from "@/lib/hooks/calendar";
import type {
  CalendarScheduleBlock,
  CalendarScheduleBlockInput,
} from "@/lib/queries/calendar/types";
import {
  CALENDAR_HISTORY_DAYS,
  CALENDAR_LOOKAHEAD_DAYS,
  getStoryCode,
  roundToNextHalfHour,
  toDateTimeInputValue,
} from "./calendar-presentation";
import type { CalendarStoryOption } from "./calendar-types";

const CalendarStoryPicker = ({
  onSelect,
  selectedStoryId,
  stories,
  storyTerm,
  storyTermPlural,
}: {
  onSelect: (storyId: string) => void;
  selectedStoryId: string;
  stories: CalendarStoryOption[];
  storyTerm: string;
  storyTermPlural: string;
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [query, setQuery] = useState("");
  const selectedStory = stories.find((story) => story.id === selectedStoryId);
  const normalizedQuery = query.trim().toLowerCase();
  const filteredStories = normalizedQuery
    ? stories.filter((story) => {
        const code = getStoryCode(story).toLowerCase();
        return (
          code.includes(normalizedQuery) ||
          story.title.toLowerCase().includes(normalizedQuery)
        );
      })
    : stories;

  const changeOpen = (open: boolean) => {
    setIsOpen(open);
    if (!open) {
      setQuery("");
    }
  };

  return (
    <Popover onOpenChange={changeOpen} open={isOpen}>
      <Popover.Trigger asChild>
        <Button
          align="between"
          aria-expanded={isOpen}
          className="h-[2.8rem] min-w-0 rounded-lg px-4 text-base md:h-[2.8rem]"
          color="tertiary"
          fullWidth
          rightIcon={<ArrowDown2Icon className="h-4 shrink-0" />}
          size="md"
          variant="outline"
        >
          {selectedStory ? (
            <span className="min-w-0 flex-1 truncate text-left">
              <span className="text-text-muted mr-2">
                {getStoryCode(selectedStory)}
              </span>
              {selectedStory.title}
            </span>
          ) : (
            <span className="text-text-muted">Select {storyTerm}</span>
          )}
        </Button>
      </Popover.Trigger>
      <Popover.Content
        align="start"
        className="z-[100] w-[var(--radix-popover-trigger-width)] max-w-[34rem] min-w-[22rem] overflow-hidden p-0"
        sideOffset={8}
      >
        <Box className="p-4">
          <Flex align="center" className="mb-4 min-w-0" justify="between">
            <Text
              className="min-w-0 flex-1 pr-3"
              color="muted"
              fontWeight="semibold"
            >
              Select {storyTerm}
            </Text>
            <Button
              aria-label={`Close ${storyTerm.toLowerCase()} picker`}
              asIcon
              color="tertiary"
              onClick={() => {
                changeOpen(false);
              }}
              rounded="full"
              size="xs"
              variant="naked"
            >
              <CloseIcon className="h-4" />
            </Button>
          </Flex>
          <Input
            aria-label={`Search ${storyTermPlural}`}
            autoFocus
            onChange={(event) => {
              setQuery(event.target.value);
            }}
            placeholder={`Search ${storyTermPlural}...`}
            rightIcon={<SearchIcon className="text-icon h-5" />}
            value={query}
          />
          <Box className="mt-3 max-h-72 overflow-y-auto pr-1">
            {filteredStories.length === 0 ? (
              <Text className="px-2 py-3" color="muted" fontSize="md">
                No matching {storyTermPlural}.
              </Text>
            ) : (
              filteredStories.map((story) => {
                const code = getStoryCode(story);
                const isSelected = story.id === selectedStoryId;
                return (
                  <button
                    className={cn(
                      "hover:bg-state-hover focus-visible:bg-state-active flex w-full min-w-0 items-center gap-3 rounded-lg px-2 py-2 text-left text-base focus-visible:outline-0",
                      { "bg-state-selected": isSelected },
                    )}
                    key={story.id}
                    onClick={() => {
                      onSelect(story.id);
                      changeOpen(false);
                    }}
                    type="button"
                  >
                    <StoryIcon className="h-[1.1rem] shrink-0" />
                    <Text
                      className="min-w-0 flex-1 truncate"
                      fontSize="md"
                      title={story.title}
                    >
                      <span className="text-text-muted mr-2">{code}</span>
                      {story.title}
                    </Text>
                    {isSelected ? (
                      <CheckIcon className="h-5 shrink-0" strokeWidth={2.1} />
                    ) : null}
                  </button>
                );
              })
            )}
          </Box>
        </Box>
      </Popover.Content>
    </Popover>
  );
};

export const CalendarScheduleDialog = ({
  candidateStories,
  editingBlock,
  isOpen,
  mode,
  onOpenChange,
  storyTerm,
  storyTermPlural,
}: {
  candidateStories: CalendarStoryOption[];
  editingBlock: CalendarScheduleBlock | null;
  isOpen: boolean;
  mode: "work" | "focus";
  onOpenChange: (value: boolean) => void;
  storyTerm: string;
  storyTermPlural: string;
}) => {
  const createBlock = useCreateCalendarScheduleBlock();
  const updateBlock = useUpdateCalendarScheduleBlock();
  const deleteBlock = useDeleteCalendarScheduleBlock();
  const defaultStart = roundToNextHalfHour(addHours(new Date(), 1));
  const defaultStoryId = candidateStories.at(0)?.id ?? "";
  const [selectedStoryId, setSelectedStoryId] = useState(
    editingBlock?.storyId ?? defaultStoryId,
  );
  const [title, setTitle] = useState(editingBlock?.title ?? "Focus time");
  const [startAt, setStartAt] = useState(() =>
    toDateTimeInputValue(editingBlock?.startAt ?? defaultStart),
  );
  const [endAt, setEndAt] = useState(() =>
    toDateTimeInputValue(editingBlock?.endAt ?? addHours(defaultStart, 1)),
  );
  const [isRangeInputValid, setIsRangeInputValid] = useState(true);
  const selectedStory = candidateStories.find(
    (story) => story.id === selectedStoryId,
  );
  const isWork = mode === "work";
  const parsedStartAt = new Date(startAt);
  const parsedEndAt = new Date(endAt);
  const earliestScheduleAt = addDays(new Date(), -CALENDAR_HISTORY_DAYS);
  const latestScheduleAt = addDays(new Date(), CALENDAR_LOOKAHEAD_DAYS);
  const hasChronologicalRange =
    Number.isFinite(parsedStartAt.getTime()) &&
    Number.isFinite(parsedEndAt.getTime()) &&
    parsedEndAt.getTime() > parsedStartAt.getTime();
  const isWithinScheduleHorizon =
    hasChronologicalRange &&
    parsedStartAt >= earliestScheduleAt &&
    parsedEndAt <= latestScheduleAt;
  const hasRequiredContent = isWork
    ? Boolean(selectedStoryId)
    : title.trim().length > 0;
  const canSubmit =
    hasRequiredContent &&
    hasChronologicalRange &&
    isWithinScheduleHorizon &&
    isRangeInputValid;
  const isSaving = createBlock.isPending || updateBlock.isPending;
  let dialogTitle = "Add focus time";
  if (editingBlock) {
    dialogTitle = "Edit calendar block";
  } else if (isWork) {
    dialogTitle = `Schedule ${storyTerm}`;
  }
  let submitLabel = isWork ? `Schedule ${storyTerm}` : "Add focus time";
  if (editingBlock) {
    submitLabel = "Save";
  }

  const close = () => {
    onOpenChange(false);
  };

  const submit = () => {
    if (!canSubmit) {
      return;
    }
    const input: CalendarScheduleBlockInput = {
      blockType: mode,
      title: isWork
        ? selectedStory?.title ?? editingBlock?.title ?? storyTerm
        : title,
      storyId: isWork ? selectedStoryId : null,
      startAt: new Date(startAt).toISOString(),
      endAt: new Date(endAt).toISOString(),
      isLocked: true,
    };

    if (editingBlock) {
      updateBlock.mutate(
        { blockId: editingBlock.id, input },
        { onSuccess: close },
      );
      return;
    }
    createBlock.mutate(input, { onSuccess: close });
  };

  const handleDelete = () => {
    if (!editingBlock) {
      return;
    }
    deleteBlock.mutate(editingBlock.id, { onSuccess: close });
  };

  return (
    <Dialog onOpenChange={onOpenChange} open={isOpen}>
      <Dialog.Content>
        <Dialog.Header>
          <Dialog.Title className="px-6 pt-0.5 text-lg">
            {dialogTitle}
          </Dialog.Title>
        </Dialog.Header>
        <Dialog.Body className="space-y-4">
          {isWork ? (
            <Box>
              <Text className="mb-2" fontSize="md" fontWeight="medium">
                {storyTerm}
              </Text>
              <CalendarStoryPicker
                onSelect={setSelectedStoryId}
                selectedStoryId={selectedStoryId}
                stories={candidateStories}
                storyTerm={storyTerm}
                storyTermPlural={storyTermPlural}
              />
              {candidateStories.length === 0 ? (
                <Text className="mt-2" color="muted" fontSize="md">
                  No assigned {storyTermPlural} found.
                </Text>
              ) : null}
            </Box>
          ) : (
            <Input
              className="text-base"
              label="Title"
              labelClassName="text-base"
              onChange={(event) => {
                setTitle(event.target.value);
              }}
              value={title}
            />
          )}
          <SmartDateTimeRangeInput
            endValue={endAt}
            label="Date and time"
            leftIcon={
              isWork ? (
                <ClockIcon className="h-5" />
              ) : (
                <TimeScheduleIcon className="h-5" />
              )
            }
            onChange={({ end, start }) => {
              setStartAt(start);
              setEndAt(end);
            }}
            onValidityChange={setIsRangeInputValid}
            startValue={startAt}
          />
          {!hasChronologicalRange ? (
            <Text color="danger" fontSize="md">
              End time must be after start time.
            </Text>
          ) : null}
          {hasChronologicalRange && !isWithinScheduleHorizon ? (
            <Text color="danger" fontSize="md">
              Choose a time from the last {CALENDAR_HISTORY_DAYS} days through
              the next {CALENDAR_LOOKAHEAD_DAYS} days.
            </Text>
          ) : null}
        </Dialog.Body>
        <Dialog.Footer className="justify-between gap-3 border-0 pt-2">
          {editingBlock ? (
            <Button
              className="text-base"
              color="danger"
              leftIcon={<DeleteIcon className="h-4" />}
              loading={deleteBlock.isPending}
              onClick={handleDelete}
              variant="naked"
            >
              Delete
            </Button>
          ) : (
            <span />
          )}
          <Flex align="center" gap={2}>
            <Button
              className="text-base"
              color="tertiary"
              onClick={close}
              variant="outline"
            >
              Cancel
            </Button>
            <Button
              className="text-base"
              color="primary"
              disabled={!canSubmit}
              loading={isSaving}
              onClick={submit}
            >
              {submitLabel}
            </Button>
          </Flex>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
