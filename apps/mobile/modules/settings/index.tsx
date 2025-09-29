import React from "react";
import { ScrollView } from "react-native";
import { Header } from "./components/header";
import { SettingsSection } from "./components/settings-section";
import { SettingsItem } from "./components/settings-item";

export const Settings = () => {
  return (
    <ScrollView className="flex-1" showsVerticalScrollIndicator={false}>
      <Header />
      <SettingsSection>
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
        <SettingsItem title="Support" onPress={() => console.log("Support")} />
        <SettingsItem
          title="Send Feedback"
          showChevron={false}
          onPress={() => console.log("Send Feedback")}
        />
        <SettingsItem
          title="Help Center"
          onPress={() => console.log("Help Center")}
        />
        <SettingsItem
          title="Privacy Policy"
          onPress={() => console.log("Privacy Policy")}
        />
        <SettingsItem
          title="Follow on Twitter"
          onPress={() => console.log("Follow on Twitter")}
        />
        <SettingsItem
          title="Rate the App"
          onPress={() => console.log("Rate the App")}
        />
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
  );
};
