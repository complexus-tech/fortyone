import React, { useState } from "react";
import { ScrollView, Linking } from "react-native";
import {
  SafeContainer,
  BottomSheetModal,
  WorkspaceSwitcher,
} from "@/components/ui";
import { SettingsSection } from "./components/settings-section";
import { SettingsItem } from "./components/settings-item";
import { externalLinks } from "./external-links";
import { useColorScheme } from "nativewind";
import { colors } from "@/constants";
import { HStack, Image, Spacer, Text, VStack } from "@expo/ui/swift-ui";
import { frame } from "@expo/ui/swift-ui/modifiers";
import { useCurrentWorkspace } from "@/lib/hooks";
import { SFSymbol } from "expo-symbols";

const handleExternalLink = async (url: string) => {
  const canOpen = await Linking.canOpenURL(url);
  if (canOpen) {
    await Linking.openURL(url);
  }
};

export const Settings = () => {
  const { colorScheme, setColorScheme } = useColorScheme();
  const { workspace } = useCurrentWorkspace();
  const [isOpened, setIsOpened] = useState(false);
  const [isAppearanceOpened, setIsAppearanceOpened] = useState(false);

  const themes = [
    {
      label: "Light",
      value: "light",
      icon: "sun.max.fill",
    },
    {
      label: "Dark",
      value: "dark",
      icon: "moon.fill",
    },
    {
      label: "Automatic",
      value: "system",
      icon: "platter.filled.top.iphone",
    },
  ];

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
      <WorkspaceSwitcher isOpened={isOpened} setIsOpened={setIsOpened} />

      <BottomSheetModal
        isOpen={isAppearanceOpened}
        onClose={() => setIsAppearanceOpened(false)}
      >
        <HStack>
          <Text weight="semibold" color={colors.gray.DEFAULT} size={15}>
            Appearance
          </Text>
        </HStack>
        {themes.map((theme) => (
          <HStack
            spacing={8}
            key={theme.value}
            onPress={() => setColorScheme(theme.value as any)}
          >
            <Image
              systemName={theme.icon as SFSymbol}
              color={colorScheme === "light" ? "black" : "white"}
              size={20}
              modifiers={[frame({ width: 30, height: 30 })]}
            />
            <VStack alignment="leading">
              <Text lineLimit={1} size={16}>
                {theme.label}
              </Text>
            </VStack>
            <Spacer />
            {theme.value === colorScheme && (
              <Image
                systemName="checkmark.circle.fill"
                color={colorScheme === "light" ? "black" : "white"}
                size={20}
              />
            )}
          </HStack>
        ))}
      </BottomSheetModal>
    </SafeContainer>
  );
};
