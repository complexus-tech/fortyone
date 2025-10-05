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
import { useWorkspaces, useCurrentWorkspace } from "@/lib/hooks";
import type { Workspace as WorkspaceType } from "@/types/workspace";

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
        <Text lineLimit={1}>{workspace.name}</Text>
        <Text size={16} color={colors.gray.DEFAULT}>
          {workspace.userRole}
        </Text>
      </VStack>

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
  const { data: workspaces = [] } = useWorkspaces();
  const { workspace } = useCurrentWorkspace();
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
            onPress={() => console.log(workspace)}
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
              <Text weight="semibold" color={colors.gray.DEFAULT} size={16}>
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
    </SafeContainer>
  );
};
