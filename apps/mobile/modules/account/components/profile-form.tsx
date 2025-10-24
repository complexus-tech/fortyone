import React from "react";
import { Button, Col, Row, Text, Wrapper } from "@/components/ui";
import { useProfile } from "@/modules/users/hooks/use-profile";

import { SymbolView } from "expo-symbols";
import { useTheme } from "@/hooks";
import { colors } from "@/constants";
import { Alert } from "react-native";
import { useAuthStore } from "@/store";
import { truncateText } from "@/lib/utils";

export const ProfileForm = () => {
  const { resolvedTheme } = useTheme();
  const { data: profile } = useProfile();
  const clearAuth = useAuthStore((state) => state.clearAuth);
  const iconColor =
    resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[200];

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
      <Wrapper className="border-0 dark:bg-dark-100/50 py-4 rounded-3xl bg-gray-100/80">
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

      <Wrapper className="border-0 dark:bg-dark-100/50 py-4 rounded-3xl bg-gray-100/80 mt-4">
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
          <Text color="muted">{truncateText(profile?.fullName, 28)}</Text>
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
          <Text color="muted">{truncateText(profile?.email, 32)}</Text>
        </Row>
      </Wrapper>
      <Col className="mt-4" gap={4}>
        <Button rounded="full" color="tertiary">
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
    </>
  );
};
