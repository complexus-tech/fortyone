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

export const getNotificationsPage = async (
  ctx: WorkspaceCtx,
  page = 1,
  pageSize = 25,
  search = "",
) => {
  try {
    const params = new URLSearchParams({
      page: String(page),
      pageSize: String(pageSize),
    });
    if (search.trim()) params.set("search", search.trim());
    const res = await get<ApiResponse<NotificationsPage>>(
      `notifications?${params.toString()}`,
      ctx,
    );
    return res.data ?? emptyNotificationsPage(page, pageSize);
  } catch (error) {
    return emptyNotificationsPage(page, pageSize);
  }
};

export const getNotifications = async (ctx: WorkspaceCtx, search = "") => {
  const page = await getNotificationsPage(ctx, 1, 25, search);
  return page.notifications;
};
