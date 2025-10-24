import React, { useState } from "react";
import {
  Button,
  Col,
  Row,
  Text,
  ThemeSwitcher,
  WorkspaceSwitcher,
  Wrapper,
} from "@/components/ui";
import { useProfile } from "@/modules/users/hooks/use-profile";

import { SymbolView } from "expo-symbols";
import { useTheme } from "@/hooks";
import { colors } from "@/constants";
import { Alert, Pressable, Linking } from "react-native";
import { useAuthStore } from "@/store";
import { toTitleCase, truncateText } from "@/lib/utils";
import { Ionicons } from "@expo/vector-icons";
import { useCurrentWorkspace } from "@/lib/hooks/use-workspaces";

export const ProfileForm = () => {
  const [isWorkspaceOpen, setIsWorkspaceOpen] = useState(false);
  const [isAppearanceOpen, setIsAppearanceOpen] = useState(false);
  const { resolvedTheme, theme } = useTheme();
  const { workspace } = useCurrentWorkspace();
  const { data: profile } = useProfile();
  const clearAuth = useAuthStore((state) => state.clearAuth);
  const iconColor =
    resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[300];

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

  const handleExternalLink = async (url: string) => {
    const canOpen = await Linking.canOpenURL(url);
    if (canOpen) {
      await Linking.openURL(url);
    }
  };

  return (
    <>
      <Wrapper className="border-0 dark:bg-dark-100/50 py-4 rounded-3xl bg-gray-100/70 mb-4">
        <Row
          justify="between"
          align="center"
          className="pb-4 border-b border-gray-200/80 dark:border-dark-100"
        >
          <Row align="center" gap={2}>
            <SymbolView
              name="building.2.fill"
              size={20}
              tintColor={iconColor}
            />
            <Text>Workspace</Text>
          </Row>
          <Pressable
            onPress={() => setIsWorkspaceOpen(true)}
            className="flex-row items-center gap-1"
          >
            <Text color="muted">{truncateText(workspace?.name, 20)}</Text>
            <Ionicons name="chevron-expand" size={15} color={iconColor} />
          </Pressable>
        </Row>
        <Row justify="between" align="center" className="pt-4">
          <Row align="center" gap={2}>
            <SymbolView
              name="paintpalette.fill"
              size={20}
              tintColor={iconColor}
            />
            <Text>Appearance</Text>
          </Row>
          <Pressable
            onPress={() => setIsAppearanceOpen(true)}
            className="flex-row items-center gap-1"
          >
            <Text color="muted">{toTitleCase(theme)}</Text>
            <Ionicons name="chevron-expand" size={15} color={iconColor} />
          </Pressable>
        </Row>
      </Wrapper>
      <Wrapper className="border-0 dark:bg-dark-100/50 py-4 rounded-3xl bg-gray-100/70">
        <Row
          justify="between"
          align="center"
          className="pb-4 border-b border-gray-200/80 dark:border-dark-100"
        >
          <Row align="center" gap={2}>
            <SymbolView name="person.fill" size={20} tintColor={iconColor} />
            <Text>Name</Text>
          </Row>
          <Text color="muted">{truncateText(profile?.fullName, 28)}</Text>
        </Row>
        <Row
          justify="between"
          align="center"
          className="py-4 border-b border-gray-200/80 dark:border-dark-100"
        >
          <Row align="center" gap={2}>
            <SymbolView name="envelope.fill" size={20} tintColor={iconColor} />
            <Text>Email</Text>
          </Row>
          <Text color="muted">{truncateText(profile?.email, 32)}</Text>
        </Row>
        <Row justify="between" align="center" className="pt-4">
          <Row align="center" gap={2}>
            <SymbolView name="at" size={20} tintColor={iconColor} />
            <Text>Username</Text>
          </Row>
          <Text color="muted">{`@${truncateText(profile?.username, 24)}`}</Text>
        </Row>
      </Wrapper>

      <Col className="mt-4" gap={4}>
        <Button
          rounded="full"
          color="tertiary"
          onPress={() => handleExternalLink("https://fortyone.app/contact")}
        >
          Send Feedback
        </Button>
        <Button rounded="full" color="tertiary">
          Rate App
        </Button>
        <Button
          rounded="full"
          color="tertiary"
          isDestructive
          onPress={handleSignOut}
        >
          Sign Out
        </Button>
      </Col>
      <ThemeSwitcher
        isOpened={isAppearanceOpen}
        setIsOpened={setIsAppearanceOpen}
      />
      <WorkspaceSwitcher
        isOpened={isWorkspaceOpen}
        setIsOpened={setIsWorkspaceOpen}
      />
    </>
  );
};
