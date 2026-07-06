import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { AppNotification, NotificationsPage } from "../types";

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
) => {
  try {
    const res = await get<ApiResponse<NotificationsPage>>(
      `notifications?page=${page}&pageSize=${pageSize}`,
      ctx,
    );
    return res.data ?? emptyNotificationsPage(page, pageSize);
  } catch (error) {
    return emptyNotificationsPage(page, pageSize);
  }
};

export const getNotifications = async (ctx: WorkspaceCtx) => {
  const page = await getNotificationsPage(ctx);
  return page.notifications;
};
