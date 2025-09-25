import React, { useState } from "react";
import { View, StyleSheet } from "react-native";
import {
  NotificationHeader,
  NotificationList,
  EmptyState,
  NotificationSkeleton,
} from "./components";

// Static data for now
const mockNotifications = [
  {
    id: "1",
    title: "Story Updated",
    message: "John Doe updated the status of 'Implement user authentication' to In Progress",
    type: "story_update" as const,
    actor: {
      name: "John Doe",
      avatar: "https://example.com/avatar1.jpg",
    },
    createdAt: "2024-01-15T10:30:00Z",
    readAt: null,
    entityId: "story-123",
  },
  {
    id: "2",
    title: "New Comment",
    message: "Sarah Wilson commented on 'Design mobile UI components'",
    type: "story_comment" as const,
    actor: {
      name: "Sarah Wilson",
      avatar: "https://example.com/avatar2.jpg",
    },
    createdAt: "2024-01-15T09:15:00Z",
    readAt: "2024-01-15T09:20:00Z",
    entityId: "story-456",
  },
  {
    id: "3",
    title: "You were mentioned",
    message: "Mike Johnson mentioned you in 'Review API documentation'",
    type: "mention" as const,
    actor: {
      name: "Mike Johnson",
      avatar: "https://example.com/avatar3.jpg",
    },
    createdAt: "2024-01-15T08:45:00Z",
    readAt: null,
    entityId: "story-789",
  },
  {
    id: "4",
    title: "Story Updated",
    message: "Emma Davis updated the priority of 'Fix login bug' to High",
    type: "story_update" as const,
    actor: {
      name: "Emma Davis",
      avatar: "https://example.com/avatar4.jpg",
    },
    createdAt: "2024-01-15T07:30:00Z",
    readAt: "2024-01-15T08:00:00Z",
    entityId: "story-101",
  },
];

export const NotificationsPage = () => {
  const [notifications] = useState(mockNotifications);
  const [isLoading, setIsLoading] = useState(false);

  const unreadCount = notifications.filter(n => !n.readAt).length;

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
        <NotificationSkeleton />
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
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#F2F2F7",
  },
});
