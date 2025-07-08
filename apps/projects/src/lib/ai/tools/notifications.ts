import { z } from "zod";
import { tool } from "ai";
import { headers } from "next/headers";
import { auth } from "@/auth";
import { getNotifications } from "@/modules/notifications/queries/get-notifications";
import { readNotification } from "@/modules/notifications/actions/read";
import { readAllNotifications } from "@/modules/notifications/actions/read-all";
import { deleteNotification } from "@/modules/notifications/actions/delete";
import { deleteAllNotifications } from "@/modules/notifications/actions/delete-all";
import { deleteReadNotifications } from "@/modules/notifications/actions/delete-read";
import { markUnread } from "@/modules/notifications/actions/mark-unread";
import { updateNotificationPreferences } from "@/modules/notifications/actions/update-preferences";
import type { AppNotification } from "@/modules/notifications/types";

export const notificationsTool = tool({
  description:
    "Manage user notifications including viewing, marking as read/unread, deleting, and managing notification preferences. Helps users stay on top of important updates.",
  parameters: z.object({
    action: z
      .enum([
        "list-notifications",
        "get-unread-count",
        "mark-as-read",
        "mark-all-as-read",
        "mark-as-unread",
        "delete-notification",
        "delete-all-notifications",
        "delete-read-notifications",
        "filter-notifications",
        "update-notification-preferences",
      ])
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

  execute: async ({
    action,
    notificationId,
    filterType,
    unreadOnly = false,
    limit = 20,
    preferenceType,
    emailEnabled,
    inAppEnabled,
    includeDetails = false,
  }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access notifications",
        };
      }

      // Get user's workspace and role for permissions
      const headersList = await headers();
      const subdomain = headersList.get("host")?.split(".")[0] || "";
      const workspace = session.workspaces.find(
        (w) => w.slug.toLowerCase() === subdomain.toLowerCase(),
      );

      if (!workspace) {
        return {
          success: false,
          error: "Unable to determine workspace",
        };
      }

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
          let notifications = await getNotifications(session);

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

          return {
            success: true,
            notifications: formattedNotifications,
            count: formattedNotifications.length,
            unreadCount: notifications.filter((n) => !n.readAt).length,
            message: `Found ${formattedNotifications.length} notifications.`,
          };
        }

        case "get-unread-count": {
          const notifications = await getNotifications(session);
          const unreadCount = notifications.filter((n) => !n.readAt).length;

          return {
            success: true,
            unreadCount,
            totalCount: notifications.length,
            message: `You have ${unreadCount} unread notifications out of ${notifications.length} total.`,
          };
        }

        case "mark-as-read": {
          if (!notificationId) {
            return {
              success: false,
              error: "Notification ID is required for mark-as-read action",
            };
          }

          const result = await readNotification(notificationId);

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
          const result = await readAllNotifications();

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

          const result = await markUnread(notificationId);

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

          const result = await deleteNotification(notificationId);

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
          const result = await deleteAllNotifications();

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
          const result = await deleteReadNotifications();

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
          let notifications = await getNotifications(session);

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
