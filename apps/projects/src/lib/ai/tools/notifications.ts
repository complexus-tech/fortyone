import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getUnreadNotificationsStrict } from "@/modules/notifications/queries/get-unread";
import { getNotificationsPageStrict } from "@/modules/notifications/queries/get-notifications";
import { readNotification } from "@/modules/notifications/actions/read";
import { readAllNotifications } from "@/modules/notifications/actions/read-all";
import { deleteNotification } from "@/modules/notifications/actions/delete";
import { deleteAllNotifications } from "@/modules/notifications/actions/delete-all";
import { deleteReadNotifications } from "@/modules/notifications/actions/delete-read";
import { markUnread } from "@/modules/notifications/actions/mark-unread";
import { updateNotificationPreferences } from "@/modules/notifications/actions/update-preferences";
import type { AppNotification } from "@/modules/notifications/types";
import { MAYA_TOOL_ACTIONS } from "@/lib/ai/tool-actions";
import { paginateRecords } from "./tool-helpers";

export const notificationsTool = tool({
  description:
    "Manage user notifications including viewing, marking as read/unread, deleting, and managing notification preferences. Helps users stay on top of important updates.",
  inputSchema: z.object({
    action: z
      .enum(MAYA_TOOL_ACTIONS.notifications)
      .describe("The notification action to perform"),

    notificationId: z
      .string()
      .optional()
      .describe("Notification ID for single notification operations"),

    filterType: z
      .enum([
        "story_update",
        "objective_update",
        "comment_reply",
        "mention",
        "key_result_update",
        "story_comment",
      ])
      .optional()
      .describe("Filter notifications by type"),

    unreadOnly: z
      .boolean()
      .optional()
      .describe("Show only unread notifications (default: false)"),

    limit: z
      .number()
      .min(1)
      .max(100)
      .optional()
      .describe(
        "Limit number of notifications returned (default: 20, max: 100)",
      ),

    page: z.number().min(1).optional().describe("Page number. Default 1."),

    pageSize: z
      .number()
      .min(1)
      .max(100)
      .optional()
      .describe("Notifications per page. Default 20, max 100."),

    preferenceType: z
      .enum([
        "story_update",
        "objective_update",
        "comment_reply",
        "mention",
        "key_result_update",
        "story_comment",
      ])
      .optional()
      .describe("Notification type for preference updates"),

    emailEnabled: z
      .boolean()
      .optional()
      .describe("Enable/disable email notifications for the specified type"),

    inAppEnabled: z
      .boolean()
      .optional()
      .describe("Enable/disable in-app notifications for the specified type"),

    includeDetails: z
      .boolean()
      .optional()
      .describe("Include detailed notification information (default: false)"),
  }),

  execute: async (
    {
      action,
      notificationId,
      filterType,
      unreadOnly = false,
      limit = 20,
      page,
      pageSize,
      preferenceType,
      emailEnabled,
      inAppEnabled,
      includeDetails = false,
    },
    { experimental_context: experimentalContext },
  ) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access notifications",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug: string })
        .workspaceSlug;

      const ctx = { session, workspaceSlug };

      // Helper function to format notification
      const formatNotification = (notification: AppNotification) => {
        const baseNotification = {
          id: notification.id,
          type: notification.type,
          entityType: notification.entityType,
          entityId: notification.entityId,
          title: notification.title,
          isRead: Boolean(notification.readAt),
          createdAt: notification.createdAt,
          readAt: notification.readAt,
        };

        if (includeDetails) {
          return {
            ...baseNotification,
            message: notification.message,
            actorId: notification.actorId,
          };
        }

        return baseNotification;
      };

      switch (action) {
        case "list-notifications": {
          const requestedPage = page ?? 1;
          const requestedPageSize = pageSize ?? limit;
          const requiresLocalFiltering = Boolean(filterType || unreadOnly);
          const notificationsPage = await getNotificationsPageStrict(
            ctx,
            requiresLocalFiltering ? 1 : requestedPage,
            requiresLocalFiltering ? 100 : requestedPageSize,
          );
          let notifications = notificationsPage.notifications;

          // Apply filters
          if (filterType) {
            notifications = notifications.filter((n) => n.type === filterType);
          }

          if (unreadOnly) {
            notifications = notifications.filter((n) => !n.readAt);
          }

          const pagedNotifications = requiresLocalFiltering
            ? paginateRecords(notifications, {
                page: requestedPage,
                pageSize: requestedPageSize,
              })
            : {
                records: notifications,
                pagination: notificationsPage.pagination,
              };

          const formattedNotifications =
            pagedNotifications.records.map(formatNotification);

          return {
            success: true,
            notifications: formattedNotifications,
            count: formattedNotifications.length,
            unreadCount: await getUnreadNotificationsStrict(ctx),
            pagination: pagedNotifications.pagination,
            resultsCapped:
              requiresLocalFiltering && notificationsPage.pagination.hasMore,
            message: `Found ${formattedNotifications.length} notifications.`,
          };
        }

        case "get-unread-count": {
          const unreadCount = await getUnreadNotificationsStrict(ctx);

          return {
            success: true,
            unreadCount,
            message: `You have ${unreadCount} unread notifications.`,
          };
        }

        case "mark-as-read": {
          if (!notificationId) {
            return {
              success: false,
              error: "Notification ID is required for mark-as-read action",
            };
          }

          const result = await readNotification(notificationId, workspaceSlug);

          if (result?.error) {
            return {
              success: false,
              error:
                result.error.message || "Failed to mark notification as read",
            };
          }

          return {
            success: true,
            message: "Successfully marked notification as read.",
          };
        }

        case "mark-all-as-read": {
          const result = await readAllNotifications(workspaceSlug);

          if (result?.error) {
            return {
              success: false,
              error:
                result.error.message ||
                "Failed to mark all notifications as read",
            };
          }

          return {
            success: true,
            message: "Successfully marked all notifications as read.",
          };
        }

        case "mark-as-unread": {
          if (!notificationId) {
            return {
              success: false,
              error: "Notification ID is required for mark-as-unread action",
            };
          }

          const result = await markUnread(notificationId, workspaceSlug);

          if (result?.error) {
            return {
              success: false,
              error:
                result.error.message || "Failed to mark notification as unread",
            };
          }

          return {
            success: true,
            message: "Successfully marked notification as unread.",
          };
        }

        case "delete-notification": {
          if (!notificationId) {
            return {
              success: false,
              error:
                "Notification ID is required for delete-notification action",
            };
          }

          const result = await deleteNotification(
            notificationId,
            workspaceSlug,
          );

          if (result?.error) {
            return {
              success: false,
              error: result.error.message || "Failed to delete notification",
            };
          }

          return {
            success: true,
            message: "Successfully deleted notification.",
          };
        }

        case "delete-all-notifications": {
          const result = await deleteAllNotifications(workspaceSlug);

          if (result?.error) {
            return {
              success: false,
              error:
                result.error.message || "Failed to delete all notifications",
            };
          }

          return {
            success: true,
            message: "Successfully deleted all notifications.",
          };
        }

        case "delete-read-notifications": {
          const result = await deleteReadNotifications(workspaceSlug);

          if (result?.error) {
            return {
              success: false,
              error:
                result.error.message || "Failed to delete read notifications",
            };
          }

          return {
            success: true,
            message: "Successfully deleted all read notifications.",
          };
        }

        case "filter-notifications": {
          const notificationsPage = await getNotificationsPageStrict(
            ctx,
            1,
            100,
          );
          let notifications = notificationsPage.notifications;

          // Apply filters
          if (filterType) {
            notifications = notifications.filter((n) => n.type === filterType);
          }

          if (unreadOnly) {
            notifications = notifications.filter((n) => !n.readAt);
          }

          // Apply limit
          if (limit) {
            notifications = notifications.slice(0, limit);
          }

          const formattedNotifications = notifications.map(formatNotification);

          const filterDescription = [
            filterType && `type: ${filterType}`,
            unreadOnly && "unread only",
          ]
            .filter(Boolean)
            .join(", ");

          return {
            success: true,
            notifications: formattedNotifications,
            count: formattedNotifications.length,
            resultsCapped: notificationsPage.pagination.hasMore,
            filters: { filterType, unreadOnly },
            message: `Found ${formattedNotifications.length} notifications${filterDescription ? ` with filters: ${filterDescription}` : ""}.`,
          };
        }

        case "update-notification-preferences": {
          if (!preferenceType) {
            return {
              success: false,
              error:
                "Preference type is required for update-notification-preferences action",
            };
          }

          if (emailEnabled === undefined && inAppEnabled === undefined) {
            return {
              success: false,
              error:
                "At least one preference (emailEnabled or inAppEnabled) must be specified",
            };
          }

          const preferences: {
            emailEnabled?: boolean;
            inAppEnabled?: boolean;
          } = {};
          if (emailEnabled !== undefined)
            preferences.emailEnabled = emailEnabled;
          if (inAppEnabled !== undefined)
            preferences.inAppEnabled = inAppEnabled;

          const result = await updateNotificationPreferences(
            preferences,
            preferenceType,
            workspaceSlug,
          );

          if (result?.error) {
            return {
              success: false,
              error:
                result.error.message ||
                "Failed to update notification preferences",
            };
          }

          const updatedSettings = [];
          if (preferences.emailEnabled !== undefined) {
            updatedSettings.push(
              `email ${preferences.emailEnabled ? "enabled" : "disabled"}`,
            );
          }
          if (preferences.inAppEnabled !== undefined) {
            updatedSettings.push(
              `in-app ${preferences.inAppEnabled ? "enabled" : "disabled"}`,
            );
          }

          return {
            success: true,
            message: `Successfully updated ${preferenceType} preferences: ${updatedSettings.join(", ")}.`,
          };
        }

        default:
          return {
            success: false,
            error: "Invalid notification action",
          };
      }
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred while managing notifications",
      };
    }
  },
});
