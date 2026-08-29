"use server";

import { createApiClient, get } from "api-client";
import { auth } from "@/auth";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { PublicPortalNotificationsPage } from "./types";

type NotificationPageInput = {
  page: number;
  pageSize?: number;
  portalSlug: string;
  unreadOnly?: boolean;
};

type NotificationInput = {
  notificationId: string;
  portalSlug: string;
};

export const getPublicPortalNotificationsAction = async ({
  page,
  pageSize = 10,
  portalSlug,
  unreadOnly = false,
}: NotificationPageInput): Promise<
  ApiResponse<PublicPortalNotificationsPage>
> => {
  try {
    const session = await auth();
    if (!session) {
      return {
        data: null,
        error: { message: "Please log in to view notifications" },
      };
    }

    const params = new URLSearchParams({
      page: String(page),
      pageSize: String(pageSize),
    });
    if (unreadOnly) {
      params.set("unreadOnly", "true");
    }
    return await get<ApiResponse<PublicPortalNotificationsPage>>(
      `portals/${portalSlug}/notifications?${params.toString()}`,
    );
  } catch (error) {
    const apiError = getApiError(error);
    return { data: null, error: apiError.error };
  }
};

export const getPublicPortalUnreadCountAction = async (
  portalSlug: string,
): Promise<ApiResponse<{ count: number }>> => {
  try {
    const session = await auth();
    if (!session) {
      return {
        data: null,
        error: { message: "Please log in to view notifications" },
      };
    }

    return await get<ApiResponse<{ count: number }>>(
      `portals/${portalSlug}/notifications/unread-count`,
    );
  } catch (error) {
    const apiError = getApiError(error);
    return { data: null, error: apiError.error };
  }
};

export const markPublicPortalNotificationReadAction = async ({
  notificationId,
  portalSlug,
}: NotificationInput): Promise<ApiResponse<null>> => {
  try {
    const session = await auth();
    if (!session) {
      return {
        data: null,
        error: { message: "Please log in to update notifications" },
      };
    }

    const client = createApiClient();
    await client.put(
      `portals/${portalSlug}/notifications/${notificationId}/read`,
    );
    return { data: null };
  } catch (error) {
    return getApiError(error);
  }
};

export const markAllPublicPortalNotificationsReadAction = async (
  portalSlug: string,
): Promise<ApiResponse<null>> => {
  try {
    const session = await auth();
    if (!session) {
      return {
        data: null,
        error: { message: "Please log in to update notifications" },
      };
    }

    const client = createApiClient();
    await client.put(`portals/${portalSlug}/notifications/read-all`);
    return { data: null };
  } catch (error) {
    return getApiError(error);
  }
};
