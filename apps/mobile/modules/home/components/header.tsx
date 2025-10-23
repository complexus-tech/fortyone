import React, { useState } from "react";
import { Avatar, Row, Text, WorkspaceSwitcher } from "@/components/ui";
import { useProfile } from "@/modules/users/hooks/use-profile";
import { ProfileSheet } from "./profile-sheet";
import { Pressable } from "react-native";
import { SymbolView } from "expo-symbols";
import { useTheme } from "@/hooks";
import { colors } from "@/constants/colors";
import { useCurrentWorkspace } from "@/lib/hooks/use-workspaces";
import { truncateText } from "@/lib/utils";

export const Header = () => {
  const { data: user } = useProfile();
  const { workspace } = useCurrentWorkspace();
  const { resolvedTheme } = useTheme();
  const [isOpened, setIsOpened] = useState(false);
  const [isWorkspaceSwitcherOpened, setIsWorkspaceSwitcherOpened] =
    useState(false);
  const iconColor =
    resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[300];

  return (
    <>
      <Row align="end" justify="between" className="mb-5" asContainer>
        <Pressable onPress={() => setIsWorkspaceSwitcherOpened(true)}>
          <Row align="center" gap={1}>
            <Avatar
              name={workspace?.name}
              className="size-[34px] mr-0.5"
              rounded="xl"
              style={{
                backgroundColor: workspace?.color,
              }}
              src={workspace?.avatarUrl}
            />
            <Text fontSize="xl" fontWeight="semibold">
              {truncateText(workspace?.name, 10)}
            </Text>
            <SymbolView
              name="chevron.down"
              weight="semibold"
              size={14}
              tintColor={iconColor}
            />
          </Row>
        </Pressable>
        <Pressable
          onPress={() => {
            setIsOpened(true);
          }}
          style={{
            zIndex: 1,
          }}
        >
          <Avatar
            name={user?.fullName || user?.username}
            className="size-[34px]"
            color={user?.avatarUrl ? "tertiary" : "primary"}
            src={user?.avatarUrl}
          />
        </Pressable>
      </Row>
      <ProfileSheet isOpened={isOpened} setIsOpened={setIsOpened} />
      <WorkspaceSwitcher
        isOpened={isWorkspaceSwitcherOpened}
        setIsOpened={setIsWorkspaceSwitcherOpened}
      />
    </>
  );
};
