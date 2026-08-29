"use client";

import { useState } from "react";
import { Box, Button, Dialog, Flex, Input, Switch, Text } from "ui";
import { DEFAULT_WORKING_DAYS } from "@/modules/settings/lib/work-schedule";

const WEEKDAYS = [
  { value: 1, shortLabel: "Mon", label: "Monday" },
  { value: 2, shortLabel: "Tue", label: "Tuesday" },
  { value: 3, shortLabel: "Wed", label: "Wednesday" },
  { value: 4, shortLabel: "Thu", label: "Thursday" },
  { value: 5, shortLabel: "Fri", label: "Friday" },
  { value: 6, shortLabel: "Sat", label: "Saturday" },
  { value: 7, shortLabel: "Sun", label: "Sunday" },
] as const;

const workingDaysSummary = (workingDays: number[]) => {
  if (
    workingDays.length === DEFAULT_WORKING_DAYS.length &&
    DEFAULT_WORKING_DAYS.every((day) => workingDays.includes(day))
  ) {
    return "Mon–Fri";
  }

  return WEEKDAYS.filter((day) => workingDays.includes(day.value))
    .map((day) => day.shortLabel)
    .join(", ");
};

const formatTime = (minutes: number) => {
  const hour = Math.floor(minutes / 60);
  const minute = minutes % 60;
  const suffix = hour >= 12 ? "PM" : "AM";
  const displayHour = hour % 12 || 12;
  return `${displayHour}:${minute.toString().padStart(2, "0")} ${suffix}`;
};

const toTimeInput = (minutes: number) =>
  `${Math.floor(minutes / 60)
    .toString()
    .padStart(2, "0")}:${(minutes % 60).toString().padStart(2, "0")}`;

const fromTimeInput = (value: string) => {
  const [hours, minutes] = value.split(":").map(Number);
  if (
    !Number.isInteger(hours) ||
    !Number.isInteger(minutes) ||
    hours < 0 ||
    hours > 23 ||
    minutes < 0 ||
    minutes > 59
  ) {
    return null;
  }
  return hours * 60 + minutes;
};

export const WorkingDaysSetting = ({
  allowInheritance = false,
  effectiveValue,
  isInherited = false,
  isPending,
  onSave,
}: {
  allowInheritance?: boolean;
  effectiveValue: number[];
  isInherited?: boolean;
  isPending: boolean;
  onSave: (workingDays: number[] | null, onSuccess: () => void) => void;
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [inherit, setInherit] = useState(isInherited);
  const [draft, setDraft] = useState(effectiveValue);

  const handleOpenChange = (open: boolean) => {
    if (open) {
      setDraft(effectiveValue);
      setInherit(isInherited);
    }
    if (!isPending) setIsOpen(open);
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

  return (
    <>
      <Flex align="center" className="gap-4 px-6 py-4" justify="between">
        <Box>
          <Text className="font-medium">Working days</Text>
          <Text color="muted">Days Maya can schedule focused work.</Text>
        </Box>
        <Button
          className="dark:bg-surface-elevated shrink-0"
          color="tertiary"
          disabled={isPending}
          onClick={() => {
            handleOpenChange(true);
          }}
          size="sm"
          variant="outline"
        >
          {workingDaysSummary(effectiveValue)}
          {isInherited ? " · Default" : ""}
        </Button>
      </Flex>

      <Dialog onOpenChange={handleOpenChange} open={isOpen}>
        <Dialog.Content size="sm">
          <Dialog.Header className="px-6 pt-6 pb-2">
            <Dialog.Title className="text-lg">Working days</Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="space-y-4 px-6 pt-2 pb-4">
            {allowInheritance ? (
              <Flex align="center" justify="between">
                <Box>
                  <Text fontWeight="medium">Use workspace default</Text>
                  <Text color="muted" fontSize="sm">
                    Follow the organization schedule.
                  </Text>
                </Box>
                <Switch checked={inherit} onCheckedChange={setInherit} />
              </Flex>
            ) : null}
            <Flex className="flex-wrap gap-2">
              {WEEKDAYS.map((day) => {
                const isSelected = draft.includes(day.value);
                return (
                  <Button
                    aria-pressed={isSelected}
                    className="min-w-12 justify-center"
                    color={isSelected ? "primary" : "tertiary"}
                    disabled={inherit}
                    key={day.value}
                    onClick={() => {
                      toggleDay(day.value);
                    }}
                    size="sm"
                    title={day.label}
                    variant={isSelected ? "solid" : "outline"}
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
              color="primary"
              loading={isPending}
              onClick={() => {
                onSave(inherit ? null : draft, () => {
                  setIsOpen(false);
                });
              }}
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

export const WorkingHoursSetting = ({
  allowInheritance = false,
  effectiveStartMinute,
  effectiveEndMinute,
  isInherited = false,
  isPending,
  onSave,
}: {
  allowInheritance?: boolean;
  effectiveStartMinute: number;
  effectiveEndMinute: number;
  isInherited?: boolean;
  isPending: boolean;
  onSave: (
    value: { endMinute: number; startMinute: number } | null,
    onSuccess: () => void,
  ) => void;
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [inherit, setInherit] = useState(isInherited);
  const [startTime, setStartTime] = useState(() =>
    toTimeInput(effectiveStartMinute),
  );
  const [endTime, setEndTime] = useState(() => toTimeInput(effectiveEndMinute));
  const startMinute = fromTimeInput(startTime);
  const endMinute = fromTimeInput(endTime);
  const isValid =
    inherit ||
    (startMinute !== null && endMinute !== null && endMinute > startMinute);

  const handleOpenChange = (open: boolean) => {
    if (open) {
      setStartTime(toTimeInput(effectiveStartMinute));
      setEndTime(toTimeInput(effectiveEndMinute));
      setInherit(isInherited);
    }
    if (!isPending) setIsOpen(open);
  };

  return (
    <>
      <Flex
        align="center"
        className="border-border gap-4 border-t px-6 py-4"
        justify="between"
      >
        <Box>
          <Text className="font-medium">Working hours</Text>
          <Text color="muted">Hours Maya can reserve on working days.</Text>
        </Box>
        <Button
          className="dark:bg-surface-elevated shrink-0"
          color="tertiary"
          disabled={isPending}
          onClick={() => {
            handleOpenChange(true);
          }}
          size="sm"
          variant="outline"
        >
          {formatTime(effectiveStartMinute)}–{formatTime(effectiveEndMinute)}
          {isInherited ? " · Default" : ""}
        </Button>
      </Flex>

      <Dialog onOpenChange={handleOpenChange} open={isOpen}>
        <Dialog.Content size="sm">
          <Dialog.Header className="px-6 pt-6 pb-2">
            <Dialog.Title className="text-lg">Working hours</Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="space-y-4 px-6 pt-2 pb-4">
            {allowInheritance ? (
              <Flex align="center" justify="between">
                <Box>
                  <Text fontWeight="medium">Use workspace default</Text>
                  <Text color="muted" fontSize="sm">
                    Follow the organization schedule.
                  </Text>
                </Box>
                <Switch checked={inherit} onCheckedChange={setInherit} />
              </Flex>
            ) : null}
            <Box className="grid grid-cols-2 gap-3">
              <Input
                disabled={inherit}
                label="Starts"
                onChange={(event) => {
                  setStartTime(event.target.value);
                }}
                step={900}
                type="time"
                value={startTime}
              />
              <Input
                disabled={inherit}
                label="Ends"
                onChange={(event) => {
                  setEndTime(event.target.value);
                }}
                step={900}
                type="time"
                value={endTime}
              />
            </Box>
            {!isValid ? (
              <Text color="danger" fontSize="sm">
                End time must be after start time.
              </Text>
            ) : null}
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
              color="primary"
              disabled={!isValid}
              loading={isPending}
              onClick={() => {
                if (inherit) {
                  onSave(null, () => {
                    setIsOpen(false);
                  });
                  return;
                }
                if (startMinute === null || endMinute === null) {
                  return;
                }
                onSave({ startMinute, endMinute }, () => {
                  setIsOpen(false);
                });
              }}
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
