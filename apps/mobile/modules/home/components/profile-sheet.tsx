import React, { useState } from "react";
import { Avatar, BottomSheetModal, WorkspaceSwitcher } from "@/components/ui";
import { colors } from "@/constants";
import { HStack, Image, Spacer, Text, VStack } from "@expo/ui/swift-ui";
import { useProfile } from "@/modules/users/hooks/use-profile";
import { useCurrentWorkspace } from "@/lib/hooks";
import { frame } from "@expo/ui/swift-ui/modifiers";
import { useColorScheme } from "nativewind";

export const ProfileSheet = ({
  isOpened,
  setIsOpened,
}: {
  isOpened: boolean;
  setIsOpened: (isOpened: boolean) => void;
}) => {
  const [isWorkspaceSwitcherOpened, setIsWorkspaceSwitcherOpened] =
    useState(false);
  const { colorScheme } = useColorScheme();
  const { data: user } = useProfile();
  const { workspace } = useCurrentWorkspace();
  const mutedTextColor =
    colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300];

  return (
    <>
      <BottomSheetModal
        isOpen={isOpened}
        onClose={() => setIsOpened(false)}
        spacing={16}
      >
        <HStack>
          <Text size={16} color={mutedTextColor}>
            {`${user?.email}`}
          </Text>
        </HStack>
        <HStack spacing={8}>
          <HStack modifiers={[frame({ width: 48, height: 48 })]}>
            <Avatar
              name={user?.fullName || user?.username}
              src={user?.avatarUrl}
              className="size-[48px]"
              rounded="2xl"
            />
          </HStack>
          <VStack alignment="leading" spacing={2}>
            <Text
              size={16}
              weight="semibold"
              lineLimit={1}
              color={mutedTextColor}
            >
              {`${user?.fullName || user?.username}`}
            </Text>
            <Text size={16} color={mutedTextColor}>
              {`Free Plan`}
            </Text>
          </VStack>
          <Spacer />
          <Image
            systemName="checkmark"
            color={
              colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[200]
            }
            size={20}
          />
        </HStack>
        <HStack spacing={8} onPress={() => setIsWorkspaceSwitcherOpened(true)}>
          <HStack modifiers={[frame({ width: 48, height: 48 })]}>
            <Avatar
              name={workspace?.name}
              src={workspace?.avatarUrl}
              className="size-[48px]"
              rounded="2xl"
              style={{
                backgroundColor: workspace?.avatarUrl
                  ? undefined
                  : workspace?.color,
              }}
            />
          </HStack>
          <VStack alignment="leading" spacing={2}>
            <Text size={16} weight="semibold" lineLimit={1}>
              Switch workspace
            </Text>
            <Text size={16} color={mutedTextColor}>
              {`${workspace?.name} • ${workspace?.userRole}`}
            </Text>
          </VStack>
          <Spacer />
          <Image
            systemName="chevron.up.chevron.down"
            color={
              colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[200]
            }
            size={20}
          />
        </HStack>
      </BottomSheetModal>
      <WorkspaceSwitcher
        isOpened={isWorkspaceSwitcherOpened}
        setIsOpened={setIsWorkspaceSwitcherOpened}
      />
    </>
  );
};
