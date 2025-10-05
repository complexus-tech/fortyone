import React, { useState } from "react";
import { ScrollView, Linking } from "react-native";
import { Avatar, SafeContainer } from "@/components/ui";
import { SettingsSection } from "./components/settings-section";
import { SettingsItem } from "./components/settings-item";
import { externalLinks } from "./external-links";
import { useColorScheme } from "nativewind";
import { colors } from "@/constants";
import {
  BottomSheet,
  Host,
  HStack,
  Image,
  Spacer,
  Text,
  VStack,
} from "@expo/ui/swift-ui";
import { frame, padding } from "@expo/ui/swift-ui/modifiers";
import { useWorkspaces, useCurrentWorkspace } from "@/lib/hooks";
import type { Workspace as WorkspaceType } from "@/types/workspace";
import { SFSymbol } from "expo-symbols";

const handleExternalLink = async (url: string) => {
  const canOpen = await Linking.canOpenURL(url);
  if (canOpen) {
    await Linking.openURL(url);
  }
};

const Workspace = ({
  isActive,
  workspace,
}: {
  isActive?: boolean;
  workspace: WorkspaceType;
}) => {
  return (
    <HStack spacing={8}>
      <HStack modifiers={[frame({ width: 30, height: 30 })]}>
        <Avatar
          style={{
            backgroundColor: workspace.avatarUrl ? undefined : workspace.color,
          }}
          name={workspace.name}
          size="md"
          rounded="lg"
          src={workspace.avatarUrl}
        />
      </HStack>
      <VStack alignment="leading">
        <Text lineLimit={1} size={16}>
          {workspace.name}
        </Text>
        <Text size={16} color={colors.gray.DEFAULT}>
          {workspace.userRole}
        </Text>
      </VStack>
      <Spacer />
      {isActive && (
        <Image systemName="checkmark.circle.fill" color="white" size={20} />
      )}
    </HStack>
  );
};

export const Settings = () => {
  const { colorScheme, setColorScheme } = useColorScheme();
  const { data: workspaces = [] } = useWorkspaces();
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
              <Text weight="semibold" color={colors.gray.DEFAULT} size={15}>
                Switch Workspace
              </Text>
            </HStack>
            {workspaces.map((wk) => (
              <Workspace
                key={wk.id}
                isActive={wk.id === workspace?.id}
                workspace={wk}
              />
            ))}
          </VStack>
        </BottomSheet>
      </Host>

      <Host matchContents style={{ position: "absolute" }}>
        <BottomSheet
          isOpened={isAppearanceOpened}
          onIsOpenedChange={(e) => setIsAppearanceOpened(e)}
          presentationDragIndicator="visible"
        >
          <VStack
            spacing={20}
            modifiers={[padding({ all: 24 })]}
            alignment="leading"
          >
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
          </VStack>
        </BottomSheet>
      </Host>
    </SafeContainer>
  );
};
