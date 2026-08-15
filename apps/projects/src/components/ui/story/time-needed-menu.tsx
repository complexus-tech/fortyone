"use client";

import {
  createContext,
  useContext,
  useId,
  useState,
  type FormEvent,
  type ReactNode,
} from "react";
import { Box, Button, Divider, Flex, Input, Popover, Text } from "ui";
import { CheckIcon, Time02Icon } from "icons";
import {
  formatTimeNeeded,
  MAX_TIME_NEEDED_MINUTES,
  normalizeTimeNeeded,
  parseTimeNeededInput,
  TIME_NEEDED_PRESETS,
  type TimeNeededUnit,
} from "@/lib/time-needed";

type TimeNeededValue = {
  estimatedDurationMinutes: number | null;
  minimumFocusBlockMinutes: number | null;
};

const TimeNeededContext = createContext<{
  open: boolean;
  setOpen: (open: boolean) => void;
}>({
  open: false,
  setOpen: () => {},
});

const useTimeNeededMenu = () => useContext(TimeNeededContext);

export const TimeNeededMenu = ({ children }: { children: ReactNode }) => {
  const [open, setOpen] = useState(false);

  return (
    <TimeNeededContext.Provider value={{ open, setOpen }}>
      <Popover onOpenChange={setOpen} open={open}>
        {children}
      </Popover>
    </TimeNeededContext.Provider>
  );
};

const Trigger = ({ children }: { children: ReactNode }) => (
  <Popover.Trigger asChild>{children}</Popover.Trigger>
);

const Items = ({
  align = "center",
  estimatedDurationMinutes,
  minimumFocusBlockMinutes,
  setTimeNeeded,
  showMinimumFocusBlock = true,
}: {
  align?: "start" | "end" | "center";
  estimatedDurationMinutes?: number | null;
  minimumFocusBlockMinutes?: number | null;
  setTimeNeeded: (value: TimeNeededValue) => void;
  showMinimumFocusBlock?: boolean;
}) => {
  const { setOpen } = useTimeNeededMenu();
  const [customValue, setCustomValue] = useState("");
  const [customUnit, setCustomUnit] = useState<TimeNeededUnit>("hours");
  const [customError, setCustomError] = useState<string | null>(null);
  const customErrorId = useId();

  const selectDuration = (duration: number | null) => {
    setTimeNeeded(
      normalizeTimeNeeded({
        estimatedDurationMinutes: duration,
        minimumFocusBlockMinutes,
      }),
    );
    setOpen(false);
  };

  const submitCustomDuration = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const duration = parseTimeNeededInput(customValue, customUnit);
    if (!duration) {
      setCustomError("Enter a time between 1 minute and 40 hours.");
      return;
    }

    setCustomError(null);
    selectDuration(duration);
  };

  const presetFocusBlockOptions: number[] = [...TIME_NEEDED_PRESETS].filter(
    (minutes) =>
      estimatedDurationMinutes && minutes <= estimatedDurationMinutes,
  );
  const focusBlockOptions =
    minimumFocusBlockMinutes &&
    estimatedDurationMinutes &&
    minimumFocusBlockMinutes <= estimatedDurationMinutes
      ? [
          ...new Set([...presetFocusBlockOptions, minimumFocusBlockMinutes]),
        ].sort((left, right) => left - right)
      : presetFocusBlockOptions;
  const canSetFocusBlock =
    showMinimumFocusBlock &&
    Boolean(
      estimatedDurationMinutes &&
        (estimatedDurationMinutes > 30 || minimumFocusBlockMinutes),
    ) &&
    focusBlockOptions.length > 0;

  return (
    <Popover.Content align={align} className="w-72 p-3">
      <Flex align="center" gap={2}>
        <Time02Icon aria-hidden className="text-text-muted h-4.5" />
        <Box>
          <Text fontWeight="medium">Time needed</Text>
          <Text className="text-xs" color="muted">
            Used to reserve enough calendar time.
          </Text>
        </Box>
      </Flex>

      <Box className="mt-3 grid grid-cols-3 gap-1.5">
        {TIME_NEEDED_PRESETS.map((minutes) => (
          <Button
            active={estimatedDurationMinutes === minutes}
            align="center"
            aria-pressed={estimatedDurationMinutes === minutes}
            className="w-full justify-center"
            color="tertiary"
            key={minutes}
            onClick={() => {
              selectDuration(minutes);
            }}
            size="sm"
            type="button"
            variant="outline"
          >
            {formatTimeNeeded(minutes)}
          </Button>
        ))}
      </Box>

      <form className="mt-3" noValidate onSubmit={submitCustomDuration}>
        <Text className="mb-1.5 text-xs" color="muted">
          Custom time
        </Text>
        <Box className="grid grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-1.5">
          <Input
            aria-describedby={customError ? customErrorId : undefined}
            aria-label="Custom time needed"
            className="h-[2.1rem] px-2"
            max={customUnit === "minutes" ? MAX_TIME_NEEDED_MINUTES : 40}
            min="0"
            onChange={(event) => {
              setCustomValue(event.target.value);
              setCustomError(null);
            }}
            placeholder={customUnit === "minutes" ? "45" : "1.5"}
            step={customUnit === "minutes" ? "1" : "0.25"}
            type="number"
            value={customValue}
          />
          <Flex
            aria-label="Custom time unit"
            className="border-border rounded-lg border p-0.5"
            role="group"
          >
            {(["minutes", "hours"] as const).map((unit) => (
              <Button
                active={customUnit === unit}
                aria-pressed={customUnit === unit}
                className="border-0"
                color="tertiary"
                key={unit}
                onClick={() => {
                  setCustomUnit(unit);
                  setCustomError(null);
                }}
                size="xs"
                type="button"
                variant="naked"
              >
                {unit === "minutes" ? "min" : "hrs"}
              </Button>
            ))}
          </Flex>
          <Button align="center" color="primary" size="sm" type="submit">
            Set
          </Button>
        </Box>
        {customError ? (
          <Text
            aria-live="polite"
            className="mt-1 text-xs"
            color="danger"
            id={customErrorId}
          >
            {customError}
          </Text>
        ) : null}
      </form>

      {estimatedDurationMinutes ? (
        <Button
          className="mt-2 px-0"
          color="tertiary"
          onClick={() => {
            selectDuration(null);
          }}
          size="xs"
          type="button"
          variant="naked"
        >
          Clear time needed
        </Button>
      ) : null}

      {canSetFocusBlock ? (
        <>
          <Divider className="my-3" />
          <Box>
            <Text fontWeight="medium">Minimum focus block</Text>
            <Text className="mt-0.5 text-xs" color="muted">
              Automatic uses 30m, or the full duration when shorter.
            </Text>
          </Box>
          <Flex className="mt-2 gap-1.5" wrap>
            <Button
              active={!minimumFocusBlockMinutes}
              aria-pressed={!minimumFocusBlockMinutes}
              color="tertiary"
              onClick={() => {
                setTimeNeeded({
                  estimatedDurationMinutes: estimatedDurationMinutes ?? null,
                  minimumFocusBlockMinutes: null,
                });
              }}
              size="xs"
              type="button"
              variant="outline"
            >
              Automatic
            </Button>
            {focusBlockOptions.map((minutes) => (
              <Button
                active={minimumFocusBlockMinutes === minutes}
                aria-pressed={minimumFocusBlockMinutes === minutes}
                color="tertiary"
                key={minutes}
                onClick={() => {
                  setTimeNeeded({
                    estimatedDurationMinutes: estimatedDurationMinutes ?? null,
                    minimumFocusBlockMinutes: minutes,
                  });
                }}
                rightIcon={
                  minimumFocusBlockMinutes === minutes ? (
                    <CheckIcon className="h-3.5" strokeWidth={2.1} />
                  ) : undefined
                }
                size="xs"
                type="button"
                variant="outline"
              >
                {formatTimeNeeded(minutes)}
              </Button>
            ))}
          </Flex>
        </>
      ) : null}
    </Popover.Content>
  );
};

TimeNeededMenu.Trigger = Trigger;
TimeNeededMenu.Items = Items;
