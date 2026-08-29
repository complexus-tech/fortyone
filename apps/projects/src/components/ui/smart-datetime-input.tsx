"use client";

import type { ReactNode } from "react";
import { useState } from "react";
import { format } from "date-fns";
import { CalendarIcon } from "icons";
import { Input } from "ui";
import type {
  ParseSmartDateTimeRangeResult,
  ParseSmartDateTimeResult,
} from "./smart-datetime";
import { parseSmartDateTime, parseSmartDateTimeRange } from "./smart-datetime";

const formatDateTime = (value: string) => {
  const date = new Date(value);
  if (!Number.isFinite(date.getTime())) return value;
  return format(date, "MMM d, yyyy 'at' h:mm a");
};

const formatDateTimeRange = (startValue: string, endValue: string) => {
  const start = new Date(startValue);
  const end = new Date(endValue);
  if (!Number.isFinite(start.getTime()) || !Number.isFinite(end.getTime())) {
    return `${startValue} – ${endValue}`;
  }

  if (start.toDateString() === end.toDateString()) {
    return `${format(start, "MMM d, yyyy 'from' h:mm a")} to ${format(end, "h:mm a")}`;
  }

  return `${format(start, "MMM d, yyyy 'at' h:mm a")} – ${format(end, "MMM d, yyyy 'at' h:mm a")}`;
};

const toDateTimeInputValue = (value: Date) =>
  format(value, "yyyy-MM-dd'T'HH:mm");

const getErrorMessage = (error: ParseSmartDateTimeResult["error"]) => {
  if (error === "missing-time") {
    return "Include a time, for example “tomorrow at 3pm”.";
  }
  if (error === "invalid") {
    return "Enter a date and time like “tomorrow at 3pm”.";
  }
  return undefined;
};

const getRangeErrorMessage = (
  error: ParseSmartDateTimeRangeResult["error"],
) => {
  if (error === "missing-end") {
    return "Include an end time, for example “tomorrow 9am to 11am”.";
  }
  if (error === "missing-time") {
    return "Include both times, for example “tomorrow 9am to 11am”.";
  }
  if (error === "invalid-range") {
    return "End time must be after start time.";
  }
  if (error === "invalid") {
    return "Enter a range like “tomorrow 9am to 11am”.";
  }
  return undefined;
};

export const SmartDateTimeRangeInput = ({
  endValue,
  label,
  leftIcon,
  onChange,
  onValidityChange,
  referenceDate,
  startValue,
}: {
  endValue: string;
  label: string;
  leftIcon?: ReactNode;
  onChange: (range: { end: string; start: string }) => void;
  onValidityChange?: (isValid: boolean) => void;
  referenceDate?: Date;
  startValue: string;
}) => {
  const [draft, setDraft] = useState(() =>
    formatDateTimeRange(startValue, endValue),
  );
  const [error, setError] =
    useState<ParseSmartDateTimeRangeResult["error"]>(null);

  const parseDraft = (nextDraft: string) =>
    parseSmartDateTimeRange(nextDraft, referenceDate ?? new Date());

  const updateRange = (result: ParseSmartDateTimeRangeResult) => {
    onValidityChange?.(result.error === null);
    if (result.error) return;

    onChange({
      end: toDateTimeInputValue(result.end),
      start: toDateTimeInputValue(result.start),
    });
  };

  const commit = (nextDraft: string) => {
    const result = parseDraft(nextDraft);
    setError(result.error);
    updateRange(result);
    if (result.error) return;

    setDraft(
      formatDateTimeRange(
        toDateTimeInputValue(result.start),
        toDateTimeInputValue(result.end),
      ),
    );
  };

  return (
    <Input
      autoComplete="off"
      className="text-base"
      hasError={Boolean(error)}
      helpText={getRangeErrorMessage(error)}
      label={label}
      labelClassName="text-base"
      leftIcon={leftIcon}
      onBlur={(event) => {
        commit(event.target.value);
      }}
      onChange={(event) => {
        const nextDraft = event.target.value;
        setDraft(nextDraft);
        setError(null);
        updateRange(parseDraft(nextDraft));
      }}
      onKeyDown={(event) => {
        if (event.key !== "Enter") return;
        event.preventDefault();
        commit(event.currentTarget.value);
      }}
      placeholder="E.g. tomorrow 9am to 11am"
      value={draft}
    />
  );
};

export const SmartDateTimeInput = ({
  label,
  max,
  min,
  onChange,
  onValidityChange,
  referenceDate,
  value,
}: {
  label: string;
  max?: string;
  min?: string;
  onChange: (value: string) => void;
  onValidityChange?: (isValid: boolean) => void;
  referenceDate?: Date;
  value: string;
}) => {
  const [draft, setDraft] = useState(() => formatDateTime(value));
  const [error, setError] = useState<ParseSmartDateTimeResult["error"]>(null);

  const commit = (nextDraft: string) => {
    const result = parseSmartDateTime(nextDraft, referenceDate ?? new Date());
    setError(result.error);
    onValidityChange?.(Boolean(result.date));
    if (!result.date) return;

    const nextValue = toDateTimeInputValue(result.date);
    setDraft(formatDateTime(nextValue));
    onChange(nextValue);
  };

  const calendarControl = (
    <span className="text-icon focus-within:ring-ring hover:bg-state-hover absolute top-[2.4rem] right-3 z-10 grid size-7 place-items-center rounded-md transition-colors focus-within:ring-2">
      <CalendarIcon aria-hidden="true" className="h-4 w-auto" />
      <input
        aria-label={`Choose ${label.toLowerCase()}`}
        className="absolute inset-0 size-full cursor-pointer opacity-0"
        max={max}
        min={min}
        onChange={(event) => {
          const nextValue = event.target.value;
          setDraft(formatDateTime(nextValue));
          setError(null);
          onChange(nextValue);
          onValidityChange?.(true);
        }}
        type="datetime-local"
        value={value}
      />
    </span>
  );

  return (
    <div className="relative">
      <Input
        autoComplete="off"
        className="pr-11 text-base"
        hasError={Boolean(error)}
        helpText={getErrorMessage(error)}
        label={label}
        labelClassName="text-base"
        onBlur={(event) => {
          commit(event.target.value);
        }}
        onChange={(event) => {
          const nextDraft = event.target.value;
          setDraft(nextDraft);
          setError(null);
          const result = parseSmartDateTime(
            nextDraft,
            referenceDate ?? new Date(),
          );
          onValidityChange?.(Boolean(result.date));
          if (result.date) {
            onChange(toDateTimeInputValue(result.date));
          }
        }}
        onKeyDown={(event) => {
          if (event.key !== "Enter") return;
          event.preventDefault();
          commit(event.currentTarget.value);
        }}
        placeholder="E.g. tomorrow at 3pm"
        value={draft}
      />
      {calendarControl}
    </div>
  );
};
