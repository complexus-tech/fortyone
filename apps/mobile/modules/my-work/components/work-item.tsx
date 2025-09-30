import React from "react";
import { Pressable } from "react-native";
import { Row, Text, Avatar } from "@/components/ui";
import { StatusDot, PriorityIcon } from "@/components/story";
import { colors } from "@/constants";

type Priority = "Urgent" | "High" | "Medium" | "Low" | "No Priority";

interface WorkItemProps {
  story: {
    id: string;
    title: string;
    status: {
      id: string;
      name: string;
      color: string;
    };
    priority: Priority;
    assignee?: {
      id: string;
      name: string;
      avatarUrl?: string;
    };
  };
  onPress?: (storyId: string) => void;
}

export const WorkItem = ({ story, onPress }: WorkItemProps) => {
  return (
    <Pressable
      style={({ pressed }) => [
        {
          backgroundColor: pressed ? colors.gray[50] : "transparent",
          paddingVertical: 10,
          paddingHorizontal: 16,
          borderBottomWidth: 0.5,
          borderBottomColor: colors.gray[100],
        },
      ]}
      onPress={() => onPress?.(story.id)}
    >
      <Row justify="between" align="center" gap={3}>
        <Row align="center" gap={2} className="flex-1">
          <Text className="flex-1" numberOfLines={1}>
            {story.title}
          </Text>
        </Row>
        <Row align="center" gap={3}>
          <StatusDot color={story.status.color} size={12} />
          <PriorityIcon priority={story.priority} size={16} />
          <Avatar
            size="sm"
            name={story.assignee?.name}
            src={story.assignee?.avatarUrl}
          />
        </Row>
      </Row>
    </Pressable>
  );
};
