import React from "react";
import { Avatar, Row, Text } from "@/components/ui";
import { View } from "react-native";
import { formatDistanceToNow } from "date-fns";
import { useMembers } from "@/modules/members/hooks/use-members";
import type { Comment } from "@/types";

export const CommentItem = ({ userId, comment, createdAt }: Comment) => {
  const { data: members = [] } = useMembers();
  const member = members.find((m) => m.id === userId);

  return (
    <View className="my-2.5 pl-0.5">
      <Row align="center">
        <Avatar
          name={member?.fullName || member?.username}
          src={member?.avatarUrl}
          size="xs"
          className="mr-2"
        />
        <Text fontSize="sm" className="opacity-90" fontWeight="medium">
          {member?.username || "Unknown"}
        </Text>
        <Text fontSize="sm" color="muted" className="ml-2">
          {formatDistanceToNow(new Date(createdAt), { addSuffix: true })}
        </Text>
      </Row>
      <Text fontSize="sm" className="mt-1 pl-7">
        {comment}
      </Text>
    </View>
  );
};
