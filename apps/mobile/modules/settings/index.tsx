import React from "react";
import { ScrollView, Linking } from "react-native";
import { SafeContainer } from "@/components/ui";
import { SettingsSection } from "./components/settings-section";
import { SettingsItem } from "./components/settings-item";
import { externalLinks } from "./external-links";
import { useColorScheme } from "nativewind";
import { colors } from "@/constants";

const handleExternalLink = async (url: string) => {
  const canOpen = await Linking.canOpenURL(url);
  if (canOpen) {
    await Linking.openURL(url);
  }
};

export const Settings = () => {
  const { colorScheme } = useColorScheme();
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
        <SettingsSection title="General">
          <SettingsItem
            title="Account Details"
            onPress={() => console.log("Account Details")}
          />
          <SettingsItem
            title="Switch Workspace"
            onPress={() => console.log("Switch Workspace")}
          />
          <SettingsItem
            title="Appearance"
            value="System"
            onPress={() => console.log("Appearance")}
          />
        </SettingsSection>
        <SettingsSection title="Support & Info">
          {externalLinks.map((link) => (
            <SettingsItem
              title={link.title}
              key={link.title}
              onPress={() => handleExternalLink(link.url)}
            />
          ))}
        </SettingsSection>
        <SettingsSection title="Account">
          <SettingsItem
            title="Manage Account"
            onPress={() => console.log("Manage Account")}
          />
          <SettingsItem
            title="Log Out"
            destructive
            showChevron={false}
            onPress={() => console.log("Log Out")}
          />
        </SettingsSection>
      </ScrollView>
    </SafeContainer>
  );
};
