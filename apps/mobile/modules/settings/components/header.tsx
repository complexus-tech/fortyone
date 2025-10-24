import React from "react";
import { Avatar, Col, Text } from "@/components/ui";
import { useProfile } from "@/modules/users/hooks/use-profile";

export const Header = () => {
  const { data: user } = useProfile();
  return (
    <Col asContainer justify="center" align="center" className="mb-6">
      <Avatar
        name={user?.fullName || user?.username}
        className="size-24"
        textClassName="text-xl"
        src={user?.avatarUrl}
      />
      <Text
        fontSize="2xl"
        fontWeight="semibold"
        className="mt-4 mb-1"
        numberOfLines={1}
      >
        {user?.fullName || user?.username}
      </Text>
      <Text color="muted" numberOfLines={1}>{`@${user?.username}`}</Text>
    </Col>
  );
};
