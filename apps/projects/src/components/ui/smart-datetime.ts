import { parse } from "chrono-node/en";

export type ParseSmartDateTimeResult =
  | { date: Date; error: null }
  | { date: null; error: "invalid" | "missing-time" };

export type ParseSmartDateTimeRangeResult =
  | { end: Date; error: null; start: Date }
  | {
      end: null;
      error: "invalid" | "invalid-range" | "missing-end" | "missing-time";
      start: null;
    };

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

export const parseSmartDateTimeRange = (
  input: string,
  referenceDate: Date,
): ParseSmartDateTimeRangeResult => {
  const value = input.trim();
  if (!value) {
    return { end: null, error: "invalid", start: null };
  }

  const result = parse(value, referenceDate, { forwardDate: true }).at(0);
  if (!result) {
    return { end: null, error: "invalid", start: null };
  }
  if (!result.end) {
    return { end: null, error: "missing-end", start: null };
  }
  if (!result.start.isCertain("hour") || !result.end.isCertain("hour")) {
    return { end: null, error: "missing-time", start: null };
  }

  const start = result.start.date();
  const end = result.end.date();
  if (!Number.isFinite(start.getTime()) || !Number.isFinite(end.getTime())) {
    return { end: null, error: "invalid", start: null };
  }
  if (end <= start) {
    return { end: null, error: "invalid-range", start: null };
  }

  return { end, error: null, start };
};
