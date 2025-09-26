import React, { useState } from "react";
import { View, StyleSheet, Text, Pressable, ScrollView } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { Ionicons } from "@expo/vector-icons";

const SettingsHeader = () => {
  const insets = useSafeAreaInsets();

  return (
    <View style={[styles.headerContainer, { paddingTop: insets.top }]}>
      <View style={styles.headerContent}>
        <Text style={styles.title}>Settings</Text>
      </View>
    </View>
  );
};

const SettingsItem = ({
  icon,
  title,
  value,
  onPress,
  showChevron = true,
  destructive = false,
}: {
  icon: string;
  title: string;
  value?: string;
  onPress?: () => void;
  showChevron?: boolean;
  destructive?: boolean;
}) => {
  return (
    <Pressable
      style={({ pressed }) => [
        styles.settingsItem,
        pressed && styles.pressedItem,
      ]}
      onPress={onPress}
    >
      <View style={styles.itemContent}>
        <View style={styles.iconContainer}>
          <Ionicons name={icon as any} size={20} color="#E5E5E7" />
        </View>
        <Text style={[styles.itemTitle, destructive && styles.destructiveText]}>
          {title}
        </Text>
        <View style={styles.rightContent}>
          {value && <Text style={styles.itemValue}>{value}</Text>}
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
    <View style={styles.section}>
      {title && <Text style={styles.sectionTitle}>{title}</Text>}
      <View style={styles.sectionContent}>{children}</View>
    </View>
  );
};

export default function Settings() {
  const [appearanceMode, setAppearanceMode] = useState(0);

  return (
    <View style={styles.container}>
      <SettingsHeader />
      <ScrollView
        style={styles.scrollView}
        showsVerticalScrollIndicator={false}
      >
        <SettingsSection title="Account & Settings">
          <SettingsItem
            icon="person-circle-outline"
            title="Account Details"
            onPress={() => console.log("Account Details")}
          />
          <SettingsItem
            icon="swap-horizontal-outline"
            title="Switch Workspace"
            onPress={() => console.log("Switch Workspace")}
          />
          <SettingsItem
            icon="color-palette-outline"
            title="Appearance"
            value="System"
            onPress={() => console.log("Appearance")}
          />
          <SettingsItem
            icon="log-out-outline"
            title="Log Out"
            destructive={true}
            showChevron={false}
            onPress={() => console.log("Log Out")}
          />
          <SettingsItem
            icon="settings-outline"
            title="Manage Account"
            onPress={() => console.log("Manage Account")}
          />
        </SettingsSection>

        <SettingsSection title="Support & Info">
          <SettingsItem
            icon="help-circle-outline"
            title="Support"
            onPress={() => console.log("Support")}
          />
          <SettingsItem
            icon="mail-outline"
            title="Send Feedback"
            showChevron={false}
            onPress={() => console.log("Send Feedback")}
          />
          <SettingsItem
            icon="book-outline"
            title="Help Center"
            onPress={() => console.log("Help Center")}
          />
          <SettingsItem
            icon="shield-checkmark-outline"
            title="Privacy Policy"
            onPress={() => console.log("Privacy Policy")}
          />
          <SettingsItem
            icon="logo-twitter"
            title="Follow on Twitter"
            onPress={() => console.log("Follow on Twitter")}
          />
          <SettingsItem
            icon="star-outline"
            title="Rate the App"
            onPress={() => console.log("Rate the App")}
          />
        </SettingsSection>
      </ScrollView>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#FFFFFF",
  },
  headerContainer: {
    backgroundColor: "#FFFFFF",
    borderBottomWidth: 1,
    borderBottomColor: "#E5E5EA",
  },
  headerContent: {
    paddingHorizontal: 16,
    paddingTop: 8,
    paddingBottom: 16,
  },
  title: {
    fontSize: 28,
    fontWeight: "600",
    color: "#000000",
  },
  scrollView: {
    flex: 1,
  },
  section: {
    marginTop: 32,
  },
  sectionTitle: {
    fontSize: 13,
    fontWeight: "400",
    color: "#8E8E93",
    textTransform: "uppercase",
    letterSpacing: 0.5,
    marginBottom: 8,
    marginHorizontal: 16,
  },
  sectionContent: {
    backgroundColor: "#FFFFFF",
  },
  settingsItem: {
    backgroundColor: "#FFFFFF",
    borderBottomWidth: 1,
    borderBottomColor: "#E5E5EA",
  },
  pressedItem: {
    backgroundColor: "#F2F2F7",
  },
  itemContent: {
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 16,
    paddingVertical: 12,
    minHeight: 44,
  },
  iconContainer: {
    width: 32,
    height: 32,
    backgroundColor: "#333333",
    borderRadius: 6,
    justifyContent: "center",
    alignItems: "center",
    marginRight: 12,
  },
  itemTitle: {
    fontSize: 16,
    fontWeight: "400",
    color: "#000000",
    flex: 1,
  },
  destructiveText: {
    color: "#FF3B30",
  },
  rightContent: {
    flexDirection: "row",
    alignItems: "center",
  },
  itemValue: {
    fontSize: 16,
    fontWeight: "400",
    color: "#8E8E93",
    marginRight: 8,
  },
});
