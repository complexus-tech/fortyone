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
import { useCurrentWorkspace } from "@/lib/hooks";
import { useRouter } from "expo-router";
import { useAuthStore } from "@/store";
import { useTheme } from "@/hooks";
import { toTitleCase } from "@/lib/utils";
import { Header } from "./components/header";

const handleExternalLink = async (url: string) => {
  const canOpen = await Linking.canOpenURL(url);
  if (canOpen) {
    await Linking.openURL(url);
  }
};

export const Settings = () => {
  const router = useRouter();
  const { theme } = useTheme();
  const { workspace } = useCurrentWorkspace();
  const clearAuth = useAuthStore((state) => state.clearAuth);
  const [isOpened, setIsOpened] = useState(false);
  const [isAppearanceOpened, setIsAppearanceOpened] = useState(false);

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
    <SafeContainer isFull>
      <Header />
      <ScrollView style={{ paddingTop: 16 }}>
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
          <SettingsItem
            title="Manage Account"
            onPress={() => {
              router.push("/account");
            }}
          />
        </SettingsSection>
        <SettingsSection title="Support & Info">
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
            title="Sign Out"
            destructive
            showChevron={false}
            onPress={handleSignOut}
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
