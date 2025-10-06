import React from "react";
import { Avatar, Row, Text, ContextMenuButton } from "@/components/ui";
import { useRouter } from "expo-router";
import { useProfile } from "@/modules/users/hooks/use-profile";

const getTimeOfDay = () => {
  const hour = new Date().getHours();
  if (hour < 12) return "morning";
  if (hour < 18) return "afternoon";
  return "evening";
};

export const Header = () => {
  const router = useRouter();
  const { data: user } = useProfile();

  return (
    <Row align="center" justify="between" className="mb-4" asContainer>
      <Row align="center" gap={2}>
        <Avatar
          name={user?.fullName || user?.username}
          size="md"
          src={user?.avatarUrl}
        />
        <Text fontSize="2xl" fontWeight="semibold" numberOfLines={1}>
          Good {getTimeOfDay()}!
        </Text>
      </Row>
      <ContextMenuButton
        actions={[
          {
            systemImage: "gear",
            label: "Settings",
            onPress: () => router.push("/settings"),
          },
          {
            systemImage: "questionmark.circle",
            label: "Help Center",
            onPress: () => {},
          },
        ]}
      />
    </Row>
  );
};
