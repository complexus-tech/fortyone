import React, { useState } from "react";
import { Avatar, Row, Text, ContextMenuButton } from "@/components/ui";
import { useRouter } from "expo-router";
import { useProfile } from "@/modules/users/hooks/use-profile";
import { ProfileSheet } from "./profile-sheet";
import { Pressable, Alert, Linking } from "react-native";

const getTimeOfDay = () => {
  const hour = new Date().getHours();
  if (hour < 12) return "morning";
  if (hour < 18) return "afternoon";
  return "evening";
};

export const Header = () => {
  const router = useRouter();
  const { data: user } = useProfile();
  const [isOpened, setIsOpened] = useState(false);

  const handleLogout = () => {
    Alert.alert("Logout", "Are you sure you want to logout?", [
      { text: "Cancel", style: "cancel" },
      {
        text: "Logout",
        style: "destructive",
        onPress: () => {
          console.log("Logout");
        },
      },
    ]);
  };

  const handleHelpCenter = async () => {
    const canOpen = await Linking.canOpenURL("https://docs.fortyone.app");
    if (canOpen) {
      await Linking.openURL("https://docs.fortyone.app");
    }
  };

  return (
    <>
      <Row align="center" justify="between" className="mb-4" asContainer>
        <Row align="center" gap={2}>
          <Pressable
            onPress={() => {
              setIsOpened(true);
            }}
            style={{
              zIndex: 1,
            }}
          >
            <Avatar
              name={user?.fullName || user?.username}
              size="lg"
              src={user?.avatarUrl}
            />
          </Pressable>
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
              onPress: handleHelpCenter,
            },
            {
              systemImage: "xmark.circle",
              label: "Logout",
              onPress: handleLogout,
            },
          ]}
        />
      </Row>
      <ProfileSheet isOpened={isOpened} setIsOpened={setIsOpened} />
    </>
  );
};
