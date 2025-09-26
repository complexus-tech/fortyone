import React from "react";
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
      style={({ pressed }) => [
        styles.settingsItem,
        pressed && styles.pressedItem,
      ]}
      onPress={onPress}
    >
      <View style={styles.itemContent}>
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
  return (
    <View style={styles.container}>
      <SettingsHeader />
      <ScrollView
        style={styles.scrollView}
        showsVerticalScrollIndicator={false}
      >
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
            destructive={true}
            showChevron={false}
            onPress={() => console.log("Log Out")}
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
    marginBottom: 28,
  },
  sectionTitle: {
    fontSize: 14,
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
    // borderBottomWidth: 1,
    // borderBottomColor: "#E5E5EA",
  },
  pressedItem: {
    backgroundColor: "#F2F2F7",
  },
  itemContent: {
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 16,
    paddingVertical: 14,
    minHeight: 44,
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
