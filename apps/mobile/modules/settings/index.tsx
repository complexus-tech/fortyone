import React, { useState } from "react";
import { ScrollView, Linking, useWindowDimensions } from "react-native";
import { SafeContainer } from "@/components/ui";
import { SettingsSection } from "./components/settings-section";
import { SettingsItem } from "./components/settings-item";
import { externalLinks } from "./external-links";
import { useColorScheme } from "nativewind";
import { colors } from "@/constants";
import {
  BottomSheet,
  Button,
  ContextMenu,
  Host,
  HStack,
  Image,
  Picker,
  Text,
  VStack,
} from "@expo/ui/swift-ui";
import {
  cornerRadius,
  frame,
  glassEffect,
  padding,
} from "@expo/ui/swift-ui/modifiers";
import { useRouter } from "expo-router";

const handleExternalLink = async (url: string) => {
  const canOpen = await Linking.canOpenURL(url);
  if (canOpen) {
    await Linking.openURL(url);
  }
};

export const Settings = () => {
  const { colorScheme } = useColorScheme();
  const { width } = useWindowDimensions();
  const [isOpened, setIsOpened] = useState(false);

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
            onPress={() => {
              setIsOpened(true);
            }}
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

      <Host>
        <BottomSheet
          isOpened={isOpened}
          onIsOpenedChange={(e) => setIsOpened(e)}
        >
          <VStack modifiers={[padding({ vertical: 20 })]} backgroundColor="red">
            <HStack>
              <Text>Switch Workspace</Text>
            </HStack>
          </VStack>
        </BottomSheet>
      </Host>
    </SafeContainer>
  );
};
