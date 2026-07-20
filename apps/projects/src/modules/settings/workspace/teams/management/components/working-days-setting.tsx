"use client";

import { useState } from "react";
import { Box, Button, Dialog, Flex, Text } from "ui";

const DEFAULT_WORKING_DAYS = [1, 2, 3, 4, 5];

const WEEKDAYS = [
  { value: 1, shortLabel: "Mon", label: "Monday" },
  { value: 2, shortLabel: "Tue", label: "Tuesday" },
  { value: 3, shortLabel: "Wed", label: "Wednesday" },
  { value: 4, shortLabel: "Thu", label: "Thursday" },
  { value: 5, shortLabel: "Fri", label: "Friday" },
  { value: 6, shortLabel: "Sat", label: "Saturday" },
  { value: 7, shortLabel: "Sun", label: "Sunday" },
] as const;

type WorkingDaysSettingProps = {
  isPending: boolean;
  sprintTerm: string;
  value?: number[];
  onSave: (workingDays: number[], onSuccess: () => void) => void;
};

const workingDaysSummary = (workingDays: number[]) => {
  if (
    workingDays.length === DEFAULT_WORKING_DAYS.length &&
    DEFAULT_WORKING_DAYS.every((day) => workingDays.includes(day))
  ) {
    return "Monday–Friday";
  }

  return WEEKDAYS.filter((day) => workingDays.includes(day.value))
    .map((day) => day.shortLabel)
    .join(", ");
};

export const WorkingDaysSetting = ({
  isPending,
  sprintTerm,
  value = DEFAULT_WORKING_DAYS,
  onSave,
}: WorkingDaysSettingProps) => {
  const [isOpen, setIsOpen] = useState(false);
  const [draft, setDraft] = useState(value);
  const hasChanges =
    draft.length !== value.length || draft.some((day) => !value.includes(day));

  const handleOpenChange = (open: boolean) => {
    if (open) {
      setDraft(value);
    }
    if (!isPending) {
      setIsOpen(open);
    }
  };

  const toggleDay = (day: number) => {
    setDraft((current) => {
      if (current.includes(day)) {
        return current.length === 1
          ? current
          : current.filter((value) => value !== day);
      }
      return [...current, day].sort((left, right) => left - right);
    });
  };

  const handleSave = () => {
    onSave(draft, () => {
      setIsOpen(false);
    });
  };

  return (
    <>
      <Flex align="center" className="gap-4 px-6 py-4" justify="between">
        <Box>
          <Text className="font-medium">Working days</Text>
          <Text className="line-clamp-2" color="muted">
            Used for {sprintTerm} progress and Maya planning. Calendar dates
            remain continuous.
          </Text>
        </Box>
        <Button
          className="dark:bg-surface-elevated shrink-0"
          color="tertiary"
          onClick={() => {
            handleOpenChange(true);
          }}
          size="sm"
          variant="outline"
        >
          {workingDaysSummary(value)}
        </Button>
      </Flex>

      <Dialog onOpenChange={handleOpenChange} open={isOpen}>
        <Dialog.Content size="sm">
          <Dialog.Header className="px-6 pt-6 pb-2">
            <Dialog.Title className="text-lg">Working days</Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="space-y-4 px-6 pt-2 pb-4">
            <Text color="muted">
              Choose the days that count toward progress and when Maya can
              schedule work. This does not change {sprintTerm} start or end
              dates.
            </Text>
            <Flex className="flex-wrap gap-2">
              {WEEKDAYS.map((day) => {
                const isSelected = draft.includes(day.value);
                return (
                  <Button
                    active={isSelected}
                    aria-pressed={isSelected}
                    className="min-w-12 justify-center"
                    color="tertiary"
                    key={day.value}
                    onClick={() => {
                      toggleDay(day.value);
                    }}
                    size="sm"
                    title={day.label}
                    variant="outline"
                  >
                    {day.shortLabel}
                  </Button>
                );
              })}
            </Flex>
          </Dialog.Body>
          <Dialog.Footer className="justify-end gap-2">
            <Button
              color="tertiary"
              disabled={isPending}
              onClick={() => {
                setIsOpen(false);
              }}
              size="sm"
              variant="outline"
            >
              Cancel
            </Button>
            <Button
              disabled={!hasChanges}
              loading={isPending}
              onClick={handleSave}
              size="sm"
            >
              Save
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
