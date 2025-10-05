import React, { useState } from "react";
import { ScrollView, Linking, useWindowDimensions, View } from "react-native";
import { Avatar, SafeContainer } from "@/components/ui";
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
  Spacer,
  Text,
  VStack,
  Switch,
} from "@expo/ui/swift-ui";
import {
  cornerRadius,
  frame,
  glassEffect,
  padding,
  background,
  clipShape,
} from "@expo/ui/swift-ui/modifiers";
import { useRouter } from "expo-router";
import { Image as ExpoUIImage } from "expo-image";

const handleExternalLink = async (url: string) => {
  const canOpen = await Linking.canOpenURL(url);
  if (canOpen) {
    await Linking.openURL(url);
  }
};

const Workspace = ({ isActive }: { isActive?: boolean }) => {
  return (
    <HStack spacing={8}>
      <HStack modifiers={[frame({ width: 32, height: 32 })]}>
        <Avatar color="primary" name="John Doe" size="md" rounded="lg" />
      </HStack>
      <Text lineLimit={1}>Airplane Mode</Text>
      {isActive && (
        <>
          <Spacer />
          <Image systemName="checkmark.circle.fill" color="white" size={20} />
        </>
      )}
    </HStack>
  );
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

      <Host matchContents style={{ position: "absolute" }}>
        <BottomSheet
          isOpened={isOpened}
          onIsOpenedChange={(e) => setIsOpened(e)}
          presentationDragIndicator="visible"
        >
          <VStack
            spacing={20}
            modifiers={[padding({ all: 24 })]}
            alignment="leading"
          >
            <HStack>
              <Text
                weight="semibold"
                color={
                  colorScheme === "light"
                    ? colors.gray.DEFAULT
                    : colors.gray[300]
                }
              >
                Switch Workspace
              </Text>
            </HStack>
            <Workspace />
            <Workspace isActive />
          </VStack>
        </BottomSheet>
      </Host>
    </SafeContainer>
  );
};
