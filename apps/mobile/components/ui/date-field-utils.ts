import { isValid, parseISO } from "date-fns";

export function dateValue(value: string | null | undefined) {
  const parsed = value ? parseISO(value.slice(0, 10)) : new Date();
  return isValid(parsed) ? parsed : new Date();
}

/** Preserve the local calendar day used by formatISO's date-only serialization. */
export function calendarDate(date: Date) {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate());
}
