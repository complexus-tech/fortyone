import React from "react";
import { View, Pressable, ScrollView } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { Ionicons } from "@expo/vector-icons";
import { Text } from "@/components/ui";

const SettingsHeader = () => {
  const insets = useSafeAreaInsets();

  return (
    <View className="bg-white" style={{ paddingTop: insets.top }}>
      <View className="px-4 pt-2 pb-4">
        <Text fontSize="3xl" fontWeight="semibold" color="black">
          Settings
        </Text>
      </View>
    </View>
  );
};

const SettingsItem = ({
  title,
  value,
  onPress,
  showChevron = true,
  destructive = false,
}: {
  title: string;
  value?: string;
  onPress?: () => void;
  showChevron?: boolean;
  destructive?: boolean;
}) => {
  return (
    <Pressable
      className="bg-white"
      style={({ pressed }) => [pressed && { backgroundColor: "#F2F2F7" }]}
      onPress={onPress}
    >
      <View className="flex-row items-center px-4 py-3.5 min-h-[44px]">
        <Text color={destructive ? "primary" : "black"} className="flex-1">
          {title}
        </Text>
        <View className="flex-row items-center">
          {value && (
            <Text color="muted" className="mr-2">
              {value}
            </Text>
          )}
          {showChevron && (
            <Ionicons name="chevron-forward" size={16} color="#C7C7CC" />
          )}
        </View>
      </View>
    </Pressable>
  );
};

const SettingsSection = ({
  title,
  children,
}: {
  title?: string;
  children: React.ReactNode;
}) => {
  return (
    <View className="mb-7">
      {title && (
        <Text
          fontSize="sm"
          fontWeight="medium"
          color="muted"
          className="tracking-wide mb-2 mx-4"
        >
          {title}
        </Text>
      )}
      <View className="bg-white">{children}</View>
    </View>
  );
};

export default function Settings() {
  return (
    <View className="flex-1 bg-white">
      <SettingsHeader />
      <ScrollView className="flex-1" showsVerticalScrollIndicator={false}>
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
          <SettingsItem
            title="Support"
            onPress={() => console.log("Support")}
          />
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
    </View>
  );
}
