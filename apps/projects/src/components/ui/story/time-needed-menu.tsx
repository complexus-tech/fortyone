"use client";

import {
  createContext,
  useContext,
  useId,
  useState,
  type FormEvent,
  type ReactNode,
} from "react";
import { Box, Button, Divider, Flex, Input, Menu, Popover, Text } from "ui";
import { cn } from "lib";
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

const FOCUS_BLOCK_PRESETS = [15, 30, 60, 120] as const;

const getPresetButtonClassName = (active: boolean) =>
  active ? "ring-primary ring-1" : "bg-state-hover! dark:bg-state-hover!";

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

type TimeNeededContentProps = {
  estimatedDurationMinutes?: number | null;
  minimumFocusBlockMinutes?: number | null;
  onDurationSelected?: () => void;
  setTimeNeeded: (value: TimeNeededValue) => void;
  showMinimumFocusBlock?: boolean;
};

const TimeNeededContent = ({
  estimatedDurationMinutes,
  minimumFocusBlockMinutes,
  onDurationSelected,
  setTimeNeeded,
  showMinimumFocusBlock = true,
}: TimeNeededContentProps) => {
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
    onDurationSelected?.();
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

  const presetFocusBlockOptions: number[] = [...FOCUS_BLOCK_PRESETS].filter(
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
    <>
      <Box>
        <Text fontWeight="medium">Time needed</Text>
        <Text className="mt-0.5 text-sm leading-5" color="muted">
          Estimate how long this work will take.
        </Text>
      </Box>

      <Box className="mt-3 grid grid-cols-3 gap-1.5">
        {TIME_NEEDED_PRESETS.map((minutes) => (
          <Button
            active={estimatedDurationMinutes === minutes}
            align="center"
            aria-pressed={estimatedDurationMinutes === minutes}
            className={cn(
              "w-full justify-center",
              getPresetButtonClassName(estimatedDurationMinutes === minutes),
            )}
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
        <Text className="mb-1.5 text-sm" color="muted">
          Custom time
        </Text>
        <Box className="grid grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-1.5">
          <Input
            aria-describedby={customError ? customErrorId : undefined}
            aria-label="Custom time needed"
            className="bg-surface-muted dark:bg-surface-muted h-[2.25rem] px-3"
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
            className="border-border h-[2.25rem] rounded-lg border p-0.5"
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
          <Button
            align="center"
            className="h-[2.25rem] px-3"
            color="primary"
            size="sm"
            type="submit"
          >
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
        <button
          className="text-foreground focus-visible:ring-primary mt-2 border-0 bg-transparent px-0 py-1 text-[0.95rem] focus-visible:ring-1 focus-visible:outline-none"
          onClick={() => {
            selectDuration(null);
          }}
          type="button"
        >
          Clear time
        </button>
      ) : null}

      {canSetFocusBlock ? (
        <>
          <Divider className="my-3" />
          <Box>
            <Text fontWeight="medium">Focus blocks</Text>
            <Text className="mt-0.5 text-[0.95rem] leading-5" color="muted">
              Automatic fills open time. Or choose a minimum block size.
            </Text>
          </Box>
          <Flex className="mt-2 gap-1.5" wrap>
            <Button
              active={!minimumFocusBlockMinutes}
              aria-pressed={!minimumFocusBlockMinutes}
              className={getPresetButtonClassName(!minimumFocusBlockMinutes)}
              color="tertiary"
              onClick={() => {
                setTimeNeeded({
                  estimatedDurationMinutes: estimatedDurationMinutes ?? null,
                  minimumFocusBlockMinutes: null,
                });
              }}
              size="sm"
              type="button"
              variant="outline"
            >
              Automatic
            </Button>
            {focusBlockOptions.map((minutes) => (
              <Button
                active={minimumFocusBlockMinutes === minutes}
                aria-pressed={minimumFocusBlockMinutes === minutes}
                className={getPresetButtonClassName(
                  minimumFocusBlockMinutes === minutes,
                )}
                color="tertiary"
                key={minutes}
                onClick={() => {
                  setTimeNeeded({
                    estimatedDurationMinutes: estimatedDurationMinutes ?? null,
                    minimumFocusBlockMinutes: minutes,
                  });
                }}
                size="sm"
                type="button"
                variant="outline"
              >
                {formatTimeNeeded(minutes)}
              </Button>
            ))}
          </Flex>
        </>
      ) : null}
    </>
  );
};

const Items = ({
  align = "center",
  ...contentProps
}: TimeNeededContentProps & {
  align?: "start" | "end" | "center";
}) => {
  const { setOpen } = useTimeNeededMenu();

  return (
    <Popover.Content align={align} className="w-80 p-3">
      <TimeNeededContent
        {...contentProps}
        onDurationSelected={() => {
          contentProps.onDurationSelected?.();
          setOpen(false);
        }}
      />
    </Popover.Content>
  );
};

const SubItems = (props: TimeNeededContentProps) => (
  <Menu.SubItems className="w-80 p-3">
    <TimeNeededContent {...props} />
  </Menu.SubItems>
);

TimeNeededMenu.Trigger = Trigger;
TimeNeededMenu.Items = Items;
TimeNeededMenu.SubItems = SubItems;
