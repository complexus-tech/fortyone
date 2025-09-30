import React, { useState } from "react";
import { View } from "react-native";
import {
  NotificationList,
  EmptyState,
  NotificationSkeleton,
} from "./components";
import { SafeContainer } from "@/components/ui";
import { Header } from "./components/header";

// Static data for now - FortyOne style
const mockNotifications = [
  {
    id: "1",
    title: "draft 2",
    message: "Closed by FortyOne",
    type: "story_update" as const,
    actor: {
      name: "FortyOne",
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
    message: "Closed by FortyOne",
    type: "story_update" as const,
    actor: {
      name: "FortyOne",
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
  {
    id: "8",
    title: "Fix authentication bug",
    message: "New comment from Sarah Chen",
    type: "story_comment" as const,
    actor: {
      name: "Sarah Chen",
      avatar: "https://example.com/avatar8.jpg",
    },
    createdAt: "2024-01-15T03:45:00Z",
    readAt: null,
    entityId: "story-105",
    status: "comment",
  },
  {
    id: "9",
    title: "Update dependencies",
    message: "Assigned by Alex Rodriguez",
    type: "story_update" as const,
    actor: {
      name: "Alex Rodriguez",
      avatar: "https://example.com/avatar9.jpg",
    },
    createdAt: "2024-01-15T02:20:00Z",
    readAt: "2024-01-15T03:00:00Z",
    entityId: "story-106",
    status: "assigned",
  },
  {
    id: "10",
    title: "Design system updates",
    message: "Priority set as urgent by Design Team",
    type: "story_update" as const,
    actor: {
      name: "Design Team",
      avatar: "https://example.com/avatar10.jpg",
    },
    createdAt: "2024-01-15T01:15:00Z",
    readAt: null,
    entityId: "story-107",
    status: "urgent",
  },
  {
    id: "11",
    title: "Database migration",
    message: "Completed by DevOps Team",
    type: "story_update" as const,
    actor: {
      name: "DevOps Team",
      avatar: "https://example.com/avatar11.jpg",
    },
    createdAt: "2024-01-15T00:30:00Z",
    readAt: "2024-01-15T01:00:00Z",
    entityId: "story-108",
    status: "completed",
  },
  {
    id: "12",
    title: "API documentation",
    message: "Closed by Technical Writer",
    type: "story_update" as const,
    actor: {
      name: "Technical Writer",
      avatar: "https://example.com/avatar12.jpg",
    },
    createdAt: "2024-01-14T23:45:00Z",
    readAt: "2024-01-15T00:00:00Z",
    entityId: "story-109",
    status: "closed",
  },
  {
    id: "13",
    title: "Performance optimization",
    message: "New comment from Performance Team",
    type: "story_comment" as const,
    actor: {
      name: "Performance Team",
      avatar: "https://example.com/avatar13.jpg",
    },
    createdAt: "2024-01-14T22:30:00Z",
    readAt: null,
    entityId: "story-110",
    status: "comment",
  },
  {
    id: "14",
    title: "Security audit",
    message: "Assigned by Security Team",
    type: "story_update" as const,
    actor: {
      name: "Security Team",
      avatar: "https://example.com/avatar14.jpg",
    },
    createdAt: "2024-01-14T21:15:00Z",
    readAt: "2024-01-14T22:00:00Z",
    entityId: "story-111",
    status: "assigned",
  },
  {
    id: "15",
    title: "Mobile app testing",
    message: "Priority set as urgent by QA Team",
    type: "story_update" as const,
    actor: {
      name: "QA Team",
      avatar: "https://example.com/avatar15.jpg",
    },
    createdAt: "2024-01-14T20:00:00Z",
    readAt: null,
    entityId: "story-112",
    status: "urgent",
  },
  {
    id: "16",
    title: "User feedback integration",
    message: "Completed by Product Team",
    type: "story_update" as const,
    actor: {
      name: "Product Team",
      avatar: "https://example.com/avatar16.jpg",
    },
    createdAt: "2024-01-14T19:30:00Z",
    readAt: "2024-01-14T20:00:00Z",
    entityId: "story-113",
    status: "completed",
  },
  {
    id: "17",
    title: "Feature flag implementation",
    message: "Closed by Engineering Team",
    type: "story_update" as const,
    actor: {
      name: "Engineering Team",
      avatar: "https://example.com/avatar17.jpg",
    },
    createdAt: "2024-01-14T18:45:00Z",
    readAt: "2024-01-14T19:00:00Z",
    entityId: "story-114",
    status: "closed",
  },
  {
    id: "18",
    title: "Code review process",
    message: "New comment from Senior Developer",
    type: "story_comment" as const,
    actor: {
      name: "Senior Developer",
      avatar: "https://example.com/avatar18.jpg",
    },
    createdAt: "2024-01-14T17:20:00Z",
    readAt: null,
    entityId: "story-115",
    status: "comment",
  },
  {
    id: "19",
    title: "Monitoring setup",
    message: "Assigned by Infrastructure Team",
    type: "story_update" as const,
    actor: {
      name: "Infrastructure Team",
      avatar: "https://example.com/avatar19.jpg",
    },
    createdAt: "2024-01-14T16:10:00Z",
    readAt: "2024-01-14T17:00:00Z",
    entityId: "story-116",
    status: "assigned",
  },
  {
    id: "20",
    title: "Bug triage session",
    message: "Priority set as urgent by Support Team",
    type: "story_update" as const,
    actor: {
      name: "Support Team",
      avatar: "https://example.com/avatar20.jpg",
    },
    createdAt: "2024-01-14T15:00:00Z",
    readAt: null,
    entityId: "story-117",
    status: "urgent",
  },
];

export const Notifications = () => {
  const [notifications] = useState(mockNotifications);
  const [isLoading, setIsLoading] = useState(false);

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

  if (isLoading && notifications.length === 0) {
    return (
      <View className="flex-1 ">
        <Header />
        <View className="flex-1">
          <NotificationSkeleton />
        </View>
      </View>
    );
  }

  if (notifications.length === 0) {
    return <EmptyState />;
  }

  return (
    <SafeContainer isFull>
      <Header />

      <View className="flex-1">
        <NotificationSkeleton />
      </View>
      {/* <NotificationList
        notifications={notifications}
        isLoading={isLoading}
        onRefresh={handleRefresh}
        onNotificationPress={handleNotificationPress}
        onNotificationLongPress={handleNotificationLongPress}
      /> */}
    </SafeContainer>
  );
};
