import React from "react";
import { Pressable } from "react-native";
import { useColorScheme } from "nativewind";
import { SymbolView } from "expo-symbols";
import { Avatar, Row, Text } from "@/components/ui";
import { useRouter } from "expo-router";
import { colors } from "@/constants";
import { useProfile } from "@/modules/users/hooks/use-profile";

const getTimeOfDay = () => {
  const hour = new Date().getHours();
  if (hour < 12) return "morning";
  if (hour < 18) return "afternoon";
  return "evening";
};

export const Header = () => {
  const { colorScheme } = useColorScheme();
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
      <Pressable
        onPress={() => {
          router.push("/settings");
        }}
      >
        <SymbolView
          name="gear"
          size={28}
          tintColor={
            colorScheme === "light" ? colors.dark[50] : colors.gray[300]
          }
        />
      </Pressable>
    </Row>
  );
};
