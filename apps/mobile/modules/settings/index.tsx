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
import { useColorScheme } from "nativewind";
import { colors } from "@/constants";
import { useCurrentWorkspace } from "@/lib/hooks";
import { useRouter } from "expo-router";

const handleExternalLink = async (url: string) => {
  const canOpen = await Linking.canOpenURL(url);
  if (canOpen) {
    await Linking.openURL(url);
  }
};

export const Settings = () => {
  const router = useRouter();
  const { colorScheme } = useColorScheme();
  const { workspace } = useCurrentWorkspace();
  const [isOpened, setIsOpened] = useState(false);
  const [isAppearanceOpened, setIsAppearanceOpened] = useState(false);

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

  return (
    <SafeContainer isFull>
      <ScrollView
        style={{
          paddingTop: 36,
          flex: 1,
          backgroundColor:
            colorScheme === "light" ? colors.white : colors.dark[300],
        }}
      >
        <SettingsSection>
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
            value={colorScheme === "dark" ? "Dark" : "Light"}
            onPress={() => setIsAppearanceOpened(true)}
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
