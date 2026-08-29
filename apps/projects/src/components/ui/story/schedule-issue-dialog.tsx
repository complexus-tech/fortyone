"use client";

import { useState } from "react";
import { addMinutes, format } from "date-fns";
import { InfoIcon } from "icons";
import { Box, Button, Dialog, Flex, Text, Wrapper } from "ui";
import { SmartDateTimeInput } from "@/components/ui/smart-datetime-input";
import { formatTimeNeeded } from "@/lib/time-needed";

export type ScheduleIssueDialogDetails = {
  estimatedDurationMinutes: number | null;
  remainingDurationMinutes?: number;
  storyCode: string;
  storyTitle: string;
};

const roundToNextHalfHour = (date: Date) => {
  const next = new Date(date);
  next.setMinutes(next.getMinutes() < 30 ? 30 : 60, 0, 0);
  return next;
};

const toDateTimeInputValue = (value: Date) =>
  format(value, "yyyy-MM-dd'T'HH:mm");

export const ScheduleIssueDialog = ({
  issue,
  isSaving,
  now,
  onClose,
  onSubmit,
}: {
  issue: ScheduleIssueDialogDetails;
  isSaving: boolean;
  now: number;
  onClose: () => void;
  onSubmit: (startAt: string) => void;
}) => {
  const defaultStart = roundToNextHalfHour(addMinutes(new Date(now), 30));
  const [startAt, setStartAt] = useState(() =>
    toDateTimeInputValue(defaultStart),
  );
  const [isStartInputValid, setIsStartInputValid] = useState(true);
  const parsedStartAt = new Date(startAt);
  const durationMinutes =
    issue.remainingDurationMinutes ?? issue.estimatedDurationMinutes ?? 0;
  const canSubmit =
    durationMinutes > 0 &&
    isStartInputValid &&
    Number.isFinite(parsedStartAt.getTime()) &&
    parsedStartAt.getTime() >= now - 5 * 60 * 1000;
  let scheduleFeedback = null;
  if (canSubmit) {
    scheduleFeedback = (
      <Wrapper className="border-warning bg-warning/10 dark:border-warning/20 dark:bg-warning/10 flex items-start gap-2 rounded-xl border px-4 py-3 shadow-none">
        <InfoIcon className="text-warning dark:text-warning mt-1 h-4.5 shrink-0" />
        <Text fontSize="md">
          Starts {format(parsedStartAt, "EEEE, MMM d 'at' h:mm a")}. Conflicts
          are shown, but your time stays locked.
        </Text>
      </Wrapper>
    );
  } else if (isStartInputValid) {
    scheduleFeedback = (
      <Text color="danger" fontSize="md">
        Choose a future start time.
      </Text>
    );
  }

  return (
    <Dialog
      onOpenChange={(isOpen) => {
        if (!isOpen) onClose();
      }}
      open
    >
      <Dialog.Content>
        <Dialog.Header>
          <Dialog.Title className="px-6 pt-0.5 text-lg">
            Choose a time for {issue.storyCode}
          </Dialog.Title>
        </Dialog.Header>
        <Dialog.Body className="space-y-4">
          <Box>
            <Text fontWeight="medium">{issue.storyTitle}</Text>
            <Text className="mt-1" color="muted" fontSize="md">
              Maya will reserve {formatTimeNeeded(durationMinutes, "full")} and
              keep this exact time locked.
            </Text>
          </Box>
          <SmartDateTimeInput
            label="Start"
            min={toDateTimeInputValue(new Date(now))}
            onChange={setStartAt}
            onValidityChange={setIsStartInputValid}
            referenceDate={new Date(now)}
            value={startAt}
          />
          {scheduleFeedback}
        </Dialog.Body>
        <Dialog.Footer className="justify-end border-0 pt-2">
          <Flex align="center" gap={2}>
            <Button color="tertiary" onClick={onClose} variant="outline">
              Cancel
            </Button>
            <Button
              color="invert"
              disabled={!canSubmit}
              loading={isSaving}
              onClick={() => {
                onSubmit(parsedStartAt.toISOString());
              }}
            >
              Lock this time
            </Button>
          </Flex>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
