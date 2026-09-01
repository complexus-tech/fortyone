import { formatISO } from "date-fns";

export const formatBulkStoryDeadline = (day: Date) =>
  formatISO(day, { representation: "date" });
