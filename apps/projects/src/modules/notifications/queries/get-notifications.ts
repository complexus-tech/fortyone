import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { NotificationsPage } from "../types";

const emptyNotificationsPage = (
  page = 1,
  pageSize = 25,
): NotificationsPage => ({
  notifications: [],
  pagination: {
    page,
    pageSize,
    hasMore: false,
    nextPage: page + 1,
  },
});

export const getNotificationsPageStrict = async (
  ctx: WorkspaceCtx,
  page = 1,
  pageSize = 25,
  search = "",
) => {
  const params = new URLSearchParams({
    page: String(page),
    pageSize: String(pageSize),
  });
  if (search.trim()) params.set("search", search.trim());
  const res = await get<ApiResponse<NotificationsPage>>(
    `notifications?${params.toString()}`,
    ctx,
  );
  if (res.error?.message) throw new Error(res.error.message);
  return res.data ?? emptyNotificationsPage(page, pageSize);
};

export const getNotificationsPage = async (
  ctx: WorkspaceCtx,
  page = 1,
  pageSize = 25,
  search = "",
) => {
  try {
    return await getNotificationsPageStrict(ctx, page, pageSize, search);
  } catch {
    return emptyNotificationsPage(page, pageSize);
  }
};

export const getNotifications = async (
  ctx: WorkspaceCtx,
  search = "",
  pageSize = 25,
) => {
  const page = await getNotificationsPage(ctx, 1, pageSize, search);
  return page.notifications;
};
