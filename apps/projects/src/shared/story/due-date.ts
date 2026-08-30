import { addDays } from "date-fns";

export type DueDateTone = "neutral" | "overdue" | "due-soon";

const DATE_ONLY_PATTERN = /^\d{4}-\d{2}-\d{2}/;

/**
 * Story deadlines are calendar dates. Normalize legacy timestamp payloads to
 * their date portion so a viewer's time zone cannot shift the displayed day.
 */
export const parseDueDate = (value: string) => {
  const dateOnly = DATE_ONLY_PATTERN.exec(value)?.[0];
  if (!dateOnly) return new Date(value);

  const [yearNumber, monthNumber, dayNumber] = dateOnly.split("-").map(Number);
  const date = new Date(yearNumber, monthNumber - 1, dayNumber);

  return date.getFullYear() === yearNumber &&
    date.getMonth() === monthNumber - 1 &&
    date.getDate() === dayNumber
    ? date
    : new Date(value);
};

export const getDueDateTone = (
  dueDate: Date,
  now: Date | null,
): DueDateTone => {
  if (!now) return "neutral";
  if (dueDate < now) return "overdue";
  return dueDate <= addDays(now, 7) ? "due-soon" : "neutral";
};
