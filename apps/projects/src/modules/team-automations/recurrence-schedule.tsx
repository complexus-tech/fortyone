import type { ComponentProps } from "react";
import { useId } from "react";
import { Input, Text } from "ui";
import { cn } from "lib";
import { AutomationSelect } from "./select-field";
import type { RecurrenceConfiguration } from "./types";

type Schedule = RecurrenceConfiguration["schedule"];

const WEEKDAYS = [
  "Sunday",
  "Monday",
  "Tuesday",
  "Wednesday",
  "Thursday",
  "Friday",
  "Saturday",
].map((label, value) => ({ value: String(value), label }));

type ScheduleValidation = {
  canonicalTimezone: string | null;
  timezoneError: string | null;
  error: string | null;
};

export const validateRecurrenceSchedule = (
  schedule: Schedule,
): ScheduleValidation => {
  const timezone = schedule.timezone.trim();
  let canonicalTimezone: string | null = null;
  if (timezone && !/^[+-]/.test(timezone)) {
    try {
      canonicalTimezone = Intl.DateTimeFormat("en", {
        timeZone: timezone,
      }).resolvedOptions().timeZone;
    } catch {
      canonicalTimezone = null;
    }
  }
  const timezoneError = canonicalTimezone
    ? null
    : "Enter a valid timezone, such as Africa/Harare or UTC.";
  const error =
    timezoneError ??
    (schedule.frequency === "monthly" &&
    (!Number.isInteger(schedule.monthDay) ||
      schedule.monthDay < 1 ||
      schedule.monthDay > 31)
      ? "Choose a day of the month from 1 to 31."
      : null);
  return { canonicalTimezone, timezoneError, error };
};

const ScheduleInput = ({
  className,
  label,
  ...props
}: Omit<ComponentProps<typeof Input>, "id" | "className" | "label"> & {
  label: string;
  className: string;
}) => {
  const id = useId();
  return (
    <div className={cn("max-w-full min-w-0 space-y-1.5", className)}>
      <label className="block" htmlFor={id}>
        {label}
      </label>
      <Input
        className="border-border h-[2.1rem] min-w-0 px-2 py-1 leading-normal"
        id={id}
        {...props}
      />
    </div>
  );
};

const scheduleSummary = (schedule: Schedule) => {
  let frequency = "Daily";
  if (schedule.frequency === "weekly") {
    frequency = `Every ${WEEKDAYS[schedule.weekday].label}`;
  } else if (schedule.frequency === "monthly") {
    frequency = `Monthly on day ${schedule.monthDay}`;
  }
  return `${frequency} at ${schedule.localTime} · ${schedule.timezone}`;
};

type RecurrenceScheduleProps = {
  schedule: Schedule;
  onChange: (partial: Partial<Schedule>) => void;
  name: string;
  onNameChange: (name: string) => void;
  disabled: boolean;
  validation: ScheduleValidation;
};

export const RecurrenceSchedule = ({
  schedule,
  onChange,
  name,
  onNameChange,
  disabled,
  validation,
}: RecurrenceScheduleProps) => (
  <section
    aria-labelledby="recurrence-schedule"
    className="border-border min-w-0 space-y-4 border-t-[0.5px] pt-4"
  >
    <Text as="h3" className="sr-only" id="recurrence-schedule">
      Schedule
    </Text>
    <div className="flex flex-wrap items-end gap-3 sm:gap-4">
      <AutomationSelect
        disabled={disabled}
        label="Repeat"
        layout="compact"
        onChange={(value) => {
          onChange({ frequency: value as Schedule["frequency"] });
        }}
        options={[
          { value: "daily", label: "Daily" },
          { value: "weekly", label: "Weekly" },
          { value: "monthly", label: "Monthly" },
        ]}
        value={schedule.frequency}
      />
      {schedule.frequency === "weekly" ? (
        <AutomationSelect
          disabled={disabled}
          label="Day of week"
          layout="compact"
          onChange={(value) => {
            onChange({ weekday: Number(value) });
          }}
          options={WEEKDAYS}
          value={String(schedule.weekday)}
        />
      ) : null}
      {schedule.frequency === "monthly" ? (
        <ScheduleInput
          className="w-28"
          disabled={disabled}
          label="Day of month"
          max={31}
          min={1}
          onChange={(event) => {
            onChange({ monthDay: Number(event.target.value) });
          }}
          type="number"
          value={schedule.monthDay}
        />
      ) : null}
    </div>
    <div className="flex flex-wrap items-end gap-3 sm:gap-4">
      <ScheduleInput
        className="w-48 sm:w-44"
        disabled={disabled}
        label="Start date"
        onChange={(event) => {
          onChange({ startsOn: event.target.value });
        }}
        type="date"
        value={schedule.startsOn}
      />
      <ScheduleInput
        className="w-40"
        disabled={disabled}
        label="Time"
        onChange={(event) => {
          onChange({ localTime: event.target.value });
        }}
        type="time"
        value={schedule.localTime}
      />
      <ScheduleInput
        aria-invalid={Boolean(validation.timezoneError)}
        className="w-44"
        disabled={disabled}
        label="Timezone"
        maxLength={100}
        onChange={(event) => {
          onChange({ timezone: event.target.value });
        }}
        placeholder="Africa/Harare"
        title={schedule.timezone}
        value={schedule.timezone}
      />
      <ScheduleInput
        className="w-64"
        disabled={disabled}
        label="Schedule name (optional)"
        maxLength={100}
        onChange={(event) => {
          onNameChange(event.target.value);
        }}
        placeholder="Use the task title"
        value={name}
      />
    </div>
    {validation.error ? null : (
      <Text aria-live="polite" color="muted">
        {scheduleSummary({
          ...schedule,
          timezone: validation.canonicalTimezone ?? schedule.timezone.trim(),
        })}
      </Text>
    )}
    {schedule.frequency === "monthly" && schedule.monthDay > 28 ? (
      <Text color="muted">Shorter months use their last day.</Text>
    ) : null}
  </section>
);
