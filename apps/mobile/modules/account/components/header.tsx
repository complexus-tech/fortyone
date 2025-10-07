import React from "react";
import { Avatar, Col, Text } from "@/components/ui";
import { useProfile } from "@/modules/users/hooks/use-profile";

export const Header = () => {
  const { data: user } = useProfile();
  return (
    <Col asContainer justify="center" align="center" className="mb-4">
      <Avatar
        name={user?.fullName || user?.username}
        className="size-18"
        textClassName="text-xl"
        src={user?.avatarUrl}
      />
      <Text fontSize="xl" fontWeight="semibold" className="mb-6 mt-2">
        {user?.fullName || user?.username}
      </Text>
    </Col>
  );
};
