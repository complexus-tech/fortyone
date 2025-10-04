import React from "react";
import { Pressable } from "react-native";
import { useRouter } from "expo-router";
import { Dot, PriorityIcon } from "@/components/icons";
import { colors } from "@/constants";
import { Row } from "./row";
import { Text } from "./text";
import { Avatar } from "./avatar";
import { Story as StoryType } from "@/modules/stories/types";
import { useTeamStatuses } from "@/modules/statuses";
import { useMembers } from "@/modules/members";
import { useColorScheme } from "nativewind";

export const Story = ({
  id,
  title,
  statusId,
  priority,
  assigneeId,
  teamId,
}: StoryType) => {
  const { colorScheme } = useColorScheme();
  const router = useRouter();
  const { data: statuses = [] } = useTeamStatuses(teamId);
  const { data: members = [] } = useMembers();
  const status = statuses.find((status) => status.id === statusId);
  const assignee = members.find((member) => member.id === assigneeId);

  const handlePress = () => {
    router.push(`/story/${id}`);
  };

  return (
    <Pressable
      style={({ pressed }) => [
        {
          backgroundColor:
            pressed && colorScheme === "light"
              ? colors.gray[50]
              : pressed && colorScheme === "dark"
                ? colors.dark[200]
                : "transparent",
          paddingVertical: 11,
          paddingHorizontal: 16,
        },
      ]}
      onPress={handlePress}
    >
      <Row justify="between" align="center" gap={3}>
        <Row align="center" gap={2} className="flex-1">
          <PriorityIcon priority={priority} size={18} />
          <Text className="flex-1" numberOfLines={1}>
            {title}
          </Text>
        </Row>
        <Row align="center" gap={3}>
          <Dot color={status?.color || colors.gray.DEFAULT} size={12} />
          <Avatar
            size="sm"
            name={assignee?.fullName || assignee?.username}
            src={assignee?.avatarUrl}
          />
        </Row>
      </Row>
    </Pressable>
  );
};
