"use client";

import { useState } from "react";
import { parse } from "chrono-node/en";
import { format } from "date-fns";
import { CalendarIcon } from "icons";
import { Input } from "ui";

type ParseSmartDateTimeResult =
  | { date: Date; error: null }
  | { date: null; error: "invalid" | "missing-time" };

const formatDateTime = (value: string) => {
  const date = new Date(value);
  if (!Number.isFinite(date.getTime())) return value;
  return format(date, "MMM d, yyyy 'at' h:mm a");
};

const toDateTimeInputValue = (value: Date) =>
  format(value, "yyyy-MM-dd'T'HH:mm");

export const parseSmartDateTime = (
  input: string,
  referenceDate: Date,
): ParseSmartDateTimeResult => {
  const value = input.trim();
  if (!value) return { date: null, error: "invalid" };

  const result = parse(value, referenceDate, { forwardDate: true }).at(0);
  if (!result) return { date: null, error: "invalid" };
  if (!result.start.isCertain("hour")) {
    return { date: null, error: "missing-time" };
  }

  const date = result.start.date();
  if (!Number.isFinite(date.getTime())) {
    return { date: null, error: "invalid" };
  }
  return { date, error: null };
};

const getErrorMessage = (error: ParseSmartDateTimeResult["error"]) => {
  if (error === "missing-time") {
    return "Include a time, for example “tomorrow at 3pm”.";
  }
  if (error === "invalid") {
    return "Enter a date and time like “tomorrow at 3pm”.";
  }
  return undefined;
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
