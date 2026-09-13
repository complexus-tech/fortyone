import type { NotificationsPage } from "../types";

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === "object" && value !== null && !Array.isArray(value);

export const parseNotificationsPage = (data: unknown): NotificationsPage => {
  if (
    !isRecord(data) ||
    !Array.isArray(data.notifications) ||
    !data.notifications.every(
      (item) => isRecord(item) && typeof item.id === "string",
    ) ||
    !isRecord(data.pagination) ||
    !Number.isInteger(data.pagination.page) ||
    Number(data.pagination.page) < 1 ||
    !Number.isInteger(data.pagination.pageSize) ||
    Number(data.pagination.pageSize) < 1 ||
    typeof data.pagination.hasMore !== "boolean" ||
    !Number.isInteger(data.pagination.nextPage) ||
    (data.pagination.hasMore &&
      Number(data.pagination.nextPage) <= Number(data.pagination.page))
  ) {
    throw new Error("The inbox response could not be read. Please try again.");
  }
  return data as NotificationsPage;
};
