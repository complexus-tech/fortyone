"use client";

import { useState } from "react";
import { addMinutes, format } from "date-fns";
import { Box, Button, Dialog, Input, Text } from "ui";
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
  const parsedStartAt = new Date(startAt);
  const durationMinutes =
    issue.remainingDurationMinutes ?? issue.estimatedDurationMinutes ?? 0;
  const endAt = addMinutes(parsedStartAt, durationMinutes);
  const canSubmit =
    durationMinutes > 0 &&
    Number.isFinite(parsedStartAt.getTime()) &&
    parsedStartAt.getTime() >= now - 5 * 60 * 1000;

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
          <Input
            className="text-base"
            label="Start"
            labelClassName="text-base"
            min={toDateTimeInputValue(new Date(now))}
            onChange={(event) => {
              setStartAt(event.target.value);
            }}
            type="datetime-local"
            value={startAt}
          />
          {canSubmit ? (
            <Text color="muted" fontSize="md">
              Ends {format(endAt, "EEEE, MMM d 'at' h:mm a")}. If this overlaps
              a meeting or another task, FortyOne will show the conflict but
              keep your choice.
            </Text>
          ) : (
            <Text color="danger" fontSize="md">
              Choose a future start time.
            </Text>
          )}
        </Dialog.Body>
        <Dialog.Footer className="gap-3 border-0 pt-2">
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
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
