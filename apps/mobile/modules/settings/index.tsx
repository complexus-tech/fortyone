import React, { useState } from "react";
import { ScrollView, Linking, Alert } from "react-native";
import {
  SafeContainer,
  WorkspaceSwitcher,
  ThemeSwitcher,
} from "@/components/ui";
import { SettingsSection } from "./components/settings-section";
import { SettingsItem } from "./components/settings-item";
import { externalLinks } from "./external-links";
import { colors } from "@/constants";
import { useCurrentWorkspace } from "@/lib/hooks";
import { useRouter } from "expo-router";
import { useAuthStore } from "@/store";
import { useTheme } from "@/hooks";
import { toTitleCase } from "@/lib/utils";

const handleExternalLink = async (url: string) => {
  const canOpen = await Linking.canOpenURL(url);
  if (canOpen) {
    await Linking.openURL(url);
  }
};

export const Settings = () => {
  const router = useRouter();
  const { theme, resolvedTheme } = useTheme();
  const { workspace } = useCurrentWorkspace();
  const clearAuth = useAuthStore((state) => state.clearAuth);
  const [isOpened, setIsOpened] = useState(false);
  const [isAppearanceOpened, setIsAppearanceOpened] = useState(false);

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
    <SafeContainer isFull>
      <ScrollView
        style={{
          paddingTop: 48,
          flex: 1,
          backgroundColor:
            resolvedTheme === "light" ? colors.white : colors.dark[200],
        }}
      >
        <SettingsSection title="Workspace">
          <SettingsItem
            title="Switch Workspace"
            asOptions
            value={workspace?.name}
            onPress={() => {
              setIsOpened(true);
            }}
          />
          <SettingsItem
            title="Appearance"
            asOptions
            value={toTitleCase(theme)}
            onPress={() => setIsAppearanceOpened(true)}
          />
        </SettingsSection>
        <SettingsSection title="Support & Info">
          <SettingsItem
            title="Appearance"
            asOptions
            value={toTitleCase(theme)}
            onPress={() => setIsAppearanceOpened(true)}
          />
          {externalLinks.map((link) => (
            <SettingsItem
              title={link.title}
              key={link.title}
              asLink
              onPress={() => handleExternalLink(link.url)}
            />
          ))}
        </SettingsSection>
        <SettingsSection title="Account">
          <SettingsItem
            title="Manage Account"
            onPress={() => {
              router.push("/account");
            }}
          />
          <SettingsItem
            title="Log Out"
            destructive
            showChevron={false}
            onPress={handleLogout}
          />
        </SettingsSection>
      </ScrollView>
      <WorkspaceSwitcher isOpened={isOpened} setIsOpened={setIsOpened} />
      <ThemeSwitcher
        isOpened={isAppearanceOpened}
        setIsOpened={setIsAppearanceOpened}
      />
    </SafeContainer>
  );
};
