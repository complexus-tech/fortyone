import React from "react";
import { Pressable } from "react-native";
import { Row, Text, Avatar } from "@/components/ui";
import { Dot, PriorityIcon } from "@/components/icons";
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
          paddingVertical: 12,
          paddingHorizontal: 16,
          borderBottomWidth: 1,
          borderBottomColor: colors.gray[50],
        },
      ]}
      onPress={() => onPress?.(story.id)}
    >
      <Row justify="between" align="center" gap={3}>
        <Row align="center" gap={2} className="flex-1">
          <PriorityIcon priority={story.priority} size={16} />
          <Text className="flex-1" numberOfLines={1}>
            {story.title}
          </Text>
        </Row>
        <Row align="center" gap={3}>
          <Dot color={story.status.color} size={12} />
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
