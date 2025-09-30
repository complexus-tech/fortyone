import React from "react";
import { Pressable } from "react-native";
import { Dot, PriorityIcon } from "@/components/icons";
import { colors } from "@/constants";
import { Row } from "./row";
import { Text } from "./text";
import { Avatar } from "./avatar";

type Priority = "Urgent" | "High" | "Medium" | "Low" | "No Priority";

interface StoryProps {
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

export const Story = ({ story, onPress }: StoryProps) => {
  return (
    <Pressable
      style={({ pressed }) => [
        {
          backgroundColor: pressed ? colors.gray[50] : "transparent",
          paddingVertical: 11,
          paddingHorizontal: 16,
        },
      ]}
      onPress={() => onPress?.(story.id)}
    >
      <Row justify="between" align="center" gap={3}>
        <Row align="center" gap={2} className="flex-1">
          <PriorityIcon priority={story.priority} size={18} />
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
