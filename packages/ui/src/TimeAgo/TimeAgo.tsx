"use client";

import {
  isThisWeek,
  isThisYear,
  isToday,
  isYesterday,
  format,
  parseISO,
  parse,
  isDate,
} from "date-fns";
import ReactTimeAgo from "react-time-ago";

const parseTimestamp = (timestamp: string): Date => {
  // Try parsing as ISO string first
  let date = parseISO(timestamp);

  // If invalid, try parsing as "dd/MM/yyyy" format
  if (isNaN(date.getTime())) {
    date = parse(timestamp, "dd/MM/yyyy", new Date());
  }

  // If still invalid, throw an error
  if (isNaN(date.getTime())) {
    throw new Error(
      "Invalid date format. Please use ISO string or dd/MM/yyyy format."
    );
  }

  return date;
};

export const formatContentTime = (timestamp: Date | string): string => {
  const date =
    typeof timestamp === "string" ? parseTimestamp(timestamp) : timestamp;

  if (!date) {
    return "Invalid date";
  }

  if (isToday(date)) {
    return format(date, "h:mm a");
  }

  if (isYesterday(date)) {
    return "Yesterday";
  }

  if (isThisWeek(date)) {
    return format(date, "EEEE"); // Full day name
  }

  if (isThisYear(date)) {
    return format(date, "d MMM");
  }

  // For messages older than a year
  return format(date, "dd/MM/yyyy");
};

export const TimeAgo = ({
  timestamp,
  className,
  forceOriginal,
}: {
  timestamp: string;
  className?: string;
  forceOriginal?: boolean;
}) => {
  if (!isDate(new Date(timestamp))) {
    return <span className={className}>{timestamp}</span>;
  }
  return (
    <span className={className}>
      {isToday(new Date(timestamp)) || forceOriginal ? (
        <time title={new Date(timestamp).toLocaleString()}>
          {formatContentTime(timestamp)}
        </time>
      ) : (
        <ReactTimeAgo date={new Date(timestamp)} />
      )}
    </span>
  );
};
