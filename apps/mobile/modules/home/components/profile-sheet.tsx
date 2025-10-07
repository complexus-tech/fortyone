import React, { useState } from "react";
import { Avatar, BottomSheetModal, WorkspaceSwitcher } from "@/components/ui";
import { colors } from "@/constants";
import { HStack, Image, Spacer, Text, VStack } from "@expo/ui/swift-ui";
import { useProfile } from "@/modules/users/hooks/use-profile";
import { useCurrentWorkspace } from "@/lib/hooks";
import { frame } from "@expo/ui/swift-ui/modifiers";
import { useColorScheme } from "nativewind";
import { useSubscription } from "@/hooks/use-subscription";
import { toTitleCase } from "@/lib/utils";
import { useRouter } from "expo-router";

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
  const router = useRouter();
  const { data: user } = useProfile();
  const { data: subscription } = useSubscription();
  const { workspace } = useCurrentWorkspace();
  const mutedTextColor =
    colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300];

  return (
    <>
      <BottomSheetModal isOpen={isOpened} onClose={() => setIsOpened(false)}>
        <Text size={16} color={mutedTextColor}>
          {`${user?.email}`}
        </Text>
        <HStack spacing={8} onPress={() => router.push("/account")}>
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
              {`${workspace?.userRole}`}
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
            <Text size={16} color={mutedTextColor} lineLimit={1}>
              {`${workspace?.name} • ${toTitleCase(subscription?.tier)} Plan`}
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
