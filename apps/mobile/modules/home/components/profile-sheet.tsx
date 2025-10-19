import React, { useState } from "react";
import { Avatar, BottomSheetModal, WorkspaceSwitcher } from "@/components/ui";
import { colors } from "@/constants";
import { Pressable, View, Text as RNText, Alert } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useProfile } from "@/modules/users/hooks/use-profile";
import { useCurrentWorkspace } from "@/lib/hooks";
import { useTheme } from "@/hooks";
import { useSubscription } from "@/hooks/use-subscription";
import { toTitleCase } from "@/lib/utils";
import { useRouter } from "expo-router";
import { hexToRgba } from "@/lib/utils/colors";
import { useAuthStore } from "@/store";

export const ProfileSheet = ({
  isOpened,
  setIsOpened,
}: {
  isOpened: boolean;
  setIsOpened: (isOpened: boolean) => void;
}) => {
  const [isWorkspaceSwitcherOpened, setIsWorkspaceSwitcherOpened] =
    useState(false);
  const { resolvedTheme } = useTheme();
  const clearAuth = useAuthStore((state) => state.clearAuth);
  const router = useRouter();
  const { data: user } = useProfile();
  const { data: subscription } = useSubscription();
  const { workspace } = useCurrentWorkspace();
  const mutedTextColor =
    resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[300];
  const backgroundColor =
    resolvedTheme === "light"
      ? hexToRgba(colors.gray[200], 0.6)
      : hexToRgba(colors.dark[50], 0.6);

  const handleLogout = () => {
    Alert.alert("Logout", "Are you sure you want to logout?", [
      { text: "Cancel", style: "cancel" },
      {
        text: "Logout",
        style: "destructive",
        onPress: () => {
          clearAuth();
        },
      },
    ]);
  };

  return (
    <>
      <BottomSheetModal isOpen={isOpened} onClose={() => setIsOpened(false)}>
        <RNText
          style={{
            fontSize: 14,
            fontWeight: "500",
            color: mutedTextColor,
          }}
        >
          {`${user?.email}`}
        </RNText>
        <Pressable
          onPress={() => router.push("/settings")}
          style={{
            flexDirection: "row",
            alignItems: "center",
            paddingVertical: 12,
            paddingHorizontal: 16,
            gap: 8,
          }}
        >
          <View
            style={{
              width: 44,
              height: 44,
              backgroundColor: backgroundColor,
              borderRadius: 12,
              justifyContent: "center",
              alignItems: "center",
            }}
          >
            <Ionicons
              name="settings"
              size={24}
              color={
                resolvedTheme === "light"
                  ? colors.gray.DEFAULT
                  : colors.gray[200]
              }
            />
          </View>
          <View style={{ flex: 1 }}>
            <RNText
              style={{
                fontSize: 14,
                fontWeight: "600",
                color: resolvedTheme === "light" ? "black" : "white",
              }}
            >
              Settings
            </RNText>
            <RNText
              style={{
                fontSize: 13.5,
                fontWeight: "500",
                color: mutedTextColor,
              }}
            >
              Manage your account settings
            </RNText>
          </View>
        </Pressable>
        <Pressable
          onPress={() => setIsWorkspaceSwitcherOpened(true)}
          style={{
            flexDirection: "row",
            alignItems: "center",
            paddingVertical: 12,
            paddingHorizontal: 16,
            gap: 8,
          }}
        >
          <View style={{ width: 44, height: 44 }}>
            <Avatar
              name={workspace?.name}
              src={workspace?.avatarUrl}
              className="size-[44px]"
              rounded="2xl"
              style={{
                backgroundColor: workspace?.avatarUrl
                  ? undefined
                  : workspace?.color,
              }}
            />
          </View>
          <View style={{ flex: 1 }}>
            <RNText
              style={{
                fontSize: 14,
                fontWeight: "600",
                color: resolvedTheme === "light" ? "black" : "white",
              }}
            >
              Switch workspace
            </RNText>
            <RNText
              style={{
                fontSize: 13.5,
                fontWeight: "500",
                color: mutedTextColor,
              }}
            >
              {`${workspace?.name} • ${toTitleCase(subscription?.tier || "free")} Plan`}
            </RNText>
          </View>
          <Ionicons
            name="chevron-collapse"
            size={14}
            color={
              resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[200]
            }
          />
        </Pressable>
        <Pressable
          onPress={handleLogout}
          style={{
            flexDirection: "row",
            alignItems: "center",
            paddingVertical: 12,
            paddingHorizontal: 16,
            gap: 8,
          }}
        >
          <View style={{ width: 44, height: 44 }}>
            <Avatar
              name={user?.fullName || user?.username}
              src={user?.avatarUrl}
              color="primary"
              className="size-[44px]"
              rounded="2xl"
            />
          </View>
          <View style={{ flex: 1 }}>
            <RNText
              style={{
                fontSize: 14,
                fontWeight: "600",
                color: colors.danger,
              }}
            >
              Logout
            </RNText>
            <RNText
              style={{
                fontSize: 13.5,
                fontWeight: "500",
                color: mutedTextColor,
              }}
            >
              Log out of your account
            </RNText>
          </View>
        </Pressable>
      </BottomSheetModal>
      <WorkspaceSwitcher
        isOpened={isWorkspaceSwitcherOpened}
        setIsOpened={setIsWorkspaceSwitcherOpened}
      />
    </>
  );
};
