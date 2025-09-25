import React, { useState } from "react";
import { View, StyleSheet } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import {
  NotificationHeader,
  NotificationList,
  EmptyState,
  NotificationSkeleton,
} from "./components";

// Static data for now - Linear style
const mockNotifications = [
  {
    id: "1",
    title: "draft 2",
    message: "Closed by Linear",
    type: "story_update" as const,
    actor: {
      name: "Linear",
      avatar: "https://example.com/avatar1.jpg",
    },
    createdAt: "2024-01-15T10:30:00Z",
    readAt: null,
    entityId: "story-123",
    status: "closed",
  },
  {
    id: "2",
    title: "Main issue",
    message: "Closed by Linear",
    type: "story_update" as const,
    actor: {
      name: "Linear",
      avatar: "https://example.com/avatar2.jpg",
    },
    createdAt: "2024-01-15T09:15:00Z",
    readAt: "2024-01-15T09:20:00Z",
    entityId: "story-456",
    status: "closed",
  },
  {
    id: "3",
    title: "upgrade nextjs",
    message:
      "Seba-hs replied: To upgrade Next.js, run: npm install next@latest",
    type: "story_comment" as const,
    actor: {
      name: "Seba-hs",
      avatar: "https://example.com/avatar3.jpg",
    },
    createdAt: "2024-01-15T08:45:00Z",
    readAt: null,
    entityId: "story-789",
    status: "comment",
  },
  {
    id: "4",
    title: "First issue the name should be too long. let...",
    message: "Priority set as urgent by greatwingoho",
    type: "story_update" as const,
    actor: {
      name: "greatwingoho",
      avatar: "https://example.com/avatar4.jpg",
    },
    createdAt: "2024-01-15T07:30:00Z",
    readAt: "2024-01-15T08:00:00Z",
    entityId: "story-101",
    status: "urgent",
  },
  {
    id: "5",
    title: "third",
    message: "Assigned by greatwingoho",
    type: "story_update" as const,
    actor: {
      name: "greatwingoho",
      avatar: "https://example.com/avatar5.jpg",
    },
    createdAt: "2024-01-15T06:15:00Z",
    readAt: "2024-01-15T07:00:00Z",
    entityId: "story-102",
    status: "assigned",
  },
  {
    id: "6",
    title: "first",
    message: "Assigned by greatwingoho",
    type: "story_update" as const,
    actor: {
      name: "greatwingoho",
      avatar: "https://example.com/avatar6.jpg",
    },
    createdAt: "2024-01-15T05:30:00Z",
    readAt: "2024-01-15T06:00:00Z",
    entityId: "story-103",
    status: "assigned",
  },
  {
    id: "7",
    title: "ttest56",
    message: "Completed past due date",
    type: "story_update" as const,
    actor: {
      name: "System",
      avatar: "https://example.com/avatar7.jpg",
    },
    createdAt: "2024-01-15T04:30:00Z",
    readAt: "2024-01-15T05:00:00Z",
    entityId: "story-104",
    status: "completed",
  },
];

export const NotificationsPage = () => {
  const [notifications] = useState(mockNotifications);
  const [isLoading, setIsLoading] = useState(false);
  const insets = useSafeAreaInsets();

  const unreadCount = notifications.filter((n) => !n.readAt).length;

  const handleRefresh = () => {
    setIsLoading(true);
    // Simulate refresh
    setTimeout(() => {
      setIsLoading(false);
    }, 1000);
  };

  const handleNotificationPress = (notification: any) => {
    console.log("Notification pressed:", notification.id);
    // Navigate to notification details or related story
  };

  const handleNotificationLongPress = (notification: any) => {
    console.log("Notification long pressed:", notification.id);
    // Show context menu (mark as read/unread, delete)
  };

  const handleFilterPress = () => {
    console.log("Filter pressed");
    // Show filter options
  };

  const handleMarkAllRead = () => {
    console.log("Mark all read pressed");
    // Mark all notifications as read
  };

  const handleDeleteAll = () => {
    console.log("Delete all pressed");
    // Delete all notifications
  };

  if (isLoading && notifications.length === 0) {
    return (
      <View style={styles.container}>
        <NotificationHeader unreadCount={unreadCount} />
        <View style={styles.contentContainer}>
          <NotificationSkeleton />
        </View>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <NotificationHeader
        unreadCount={unreadCount}
        onFilterPress={handleFilterPress}
        onMarkAllRead={handleMarkAllRead}
        onDeleteAll={handleDeleteAll}
      />

      <View style={styles.contentContainer}>
        {notifications.length === 0 ? (
          <EmptyState />
        ) : (
          <NotificationList
            notifications={notifications}
            isLoading={isLoading}
            onRefresh={handleRefresh}
            onNotificationPress={handleNotificationPress}
            onNotificationLongPress={handleNotificationLongPress}
          />
        )}
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#F2F2F7",
  },
  contentContainer: {
    flex: 1,
  },
});
