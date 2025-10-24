import React, { useState } from "react";
import {
  Avatar,
  ContextMenuButton,
  Row,
  Text,
  WorkspaceSwitcher,
} from "@/components/ui";
import { Alert, Pressable } from "react-native";
import { SymbolView } from "expo-symbols";
import { useTheme } from "@/hooks";
import { colors } from "@/constants/colors";
import { useCurrentWorkspace } from "@/lib/hooks/use-workspaces";
import { truncateText } from "@/lib/utils";
import { HStack } from "@expo/ui/swift-ui";
import { frame } from "@expo/ui/swift-ui/modifiers";
import { useProfile } from "@/modules/users/hooks/use-profile";
import { useRouter } from "expo-router";
import { useAuthStore } from "@/store/auth";

export const Header = () => {
  const router = useRouter();
  const clearAuth = useAuthStore((state) => state.clearAuth);
  const { workspace } = useCurrentWorkspace();
  const { data: profile } = useProfile();
  const { resolvedTheme } = useTheme();
  const [isWorkspaceSwitcherOpened, setIsWorkspaceSwitcherOpened] =
    useState(false);
  const iconColor =
    resolvedTheme === "light" ? colors.dark.DEFAULT : colors.white;

  const handleSignOut = () => {
    Alert.alert("Sign Out", "Are you sure you want to sign out?", [
      { text: "Cancel", style: "cancel" },
      {
        text: "Sign Out",
        style: "destructive",
        onPress: () => {
          clearAuth();
        },
      },
    ]);
  };

  return (
    <>
      <Row align="end" justify="between" className="mb-5" asContainer>
        <Pressable onPress={() => setIsWorkspaceSwitcherOpened(true)}>
          <Row align="center" gap={1}>
            <Avatar
              name={workspace?.name}
              className="size-[34px] mr-0.5"
              rounded="xl"
              style={{
                backgroundColor: workspace?.color,
              }}
              src={workspace?.avatarUrl}
            />
            <Text fontSize="xl" fontWeight="semibold">
              {truncateText(workspace?.name, 10)}
            </Text>
            <SymbolView
              name="chevron.down"
              weight="semibold"
              size={13}
              tintColor={iconColor}
            />
          </Row>
        </Pressable>
        <ContextMenuButton
          actions={[
            {
              label: "Settings",
              systemImage: "gear",
              onPress: () => {
                router.push("/settings");
              },
            },
            {
              label: "Sign Out",
              systemImage: "rectangle.portrait.and.arrow.forward",
              onPress: handleSignOut,
            },
          ]}
          hostStyle={{ width: 36, height: 36 }}
        >
          <HStack modifiers={[frame({ width: 36, height: 36 })]}>
            <Avatar
              name={profile?.fullName || profile?.username}
              className="size-[36px]"
              color={profile?.avatarUrl ? "tertiary" : "primary"}
              src={profile?.avatarUrl}
            />
          </HStack>
        </ContextMenuButton>
      </Row>
      <WorkspaceSwitcher
        isOpened={isWorkspaceSwitcherOpened}
        setIsOpened={setIsWorkspaceSwitcherOpened}
      />
    </>
  );
};
