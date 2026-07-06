import { format } from "date-fns";

const isoTimestampPattern =
  /\b\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z\b/g;

export const formatActivityReasonDates = (reason: string) =>
  reason.replaceAll(isoTimestampPattern, (value) =>
    format(new Date(value), "PP 'at' p"),
  );
