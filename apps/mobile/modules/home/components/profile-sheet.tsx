import React, { useState } from "react";
import { Avatar, BottomSheetModal, WorkspaceSwitcher } from "@/components/ui";
import { colors } from "@/constants";
import { HStack, Image, Spacer, Text, VStack } from "@expo/ui/swift-ui";
import { useProfile } from "@/modules/users/hooks/use-profile";
import { useCurrentWorkspace } from "@/lib/hooks";
import { frame } from "@expo/ui/swift-ui/modifiers";
import { useTheme } from "@/hooks";
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
  const { resolvedTheme } = useTheme();
  const router = useRouter();
  const { data: user } = useProfile();
  const { data: subscription } = useSubscription();
  const { workspace } = useCurrentWorkspace();
  const mutedTextColor =
    resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[300];

  return (
    <>
      <BottomSheetModal isOpen={isOpened} onClose={() => setIsOpened(false)}>
        <Text size={14} weight="medium" color={mutedTextColor}>
          {`${user?.email}`}
        </Text>
        <HStack spacing={8} onPress={() => router.push("/account")}>
          <HStack modifiers={[frame({ width: 44, height: 44 })]}>
            <Avatar
              name={user?.fullName || user?.username}
              src={user?.avatarUrl}
              className="size-[44px]"
              rounded="2xl"
            />
          </HStack>
          <VStack alignment="leading" spacing={2}>
            <Text
              size={14}
              weight="semibold"
              lineLimit={1}
              color={mutedTextColor}
            >
              {`${user?.fullName || user?.username}`}
            </Text>
            <Text size={13.5} weight="medium" color={mutedTextColor}>
              {`${workspace?.userRole}`}
            </Text>
          </VStack>
          <Spacer />
          <Image
            systemName="checkmark"
            color={
              resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[200]
            }
            size={16}
          />
        </HStack>
        <HStack spacing={8} onPress={() => setIsWorkspaceSwitcherOpened(true)}>
          <HStack modifiers={[frame({ width: 44, height: 44 })]}>
            <Avatar
              name={workspace?.name}
              src={workspace?.avatarUrl}
              className="size-[44px]"
              rounded="2xl"
              style={{
                backgroundColor: workspace?.avatarUrl
                  ? undefined
                  : workspace?.color,
              }}
            />
          </HStack>
          <VStack alignment="leading" spacing={2}>
            <Text size={14} weight="semibold" lineLimit={1}>
              Switch workspace
            </Text>
            <Text
              size={13.5}
              weight="medium"
              color={mutedTextColor}
              lineLimit={1}
            >
              {`${workspace?.name} • ${toTitleCase(subscription?.tier || "free")} Plan`}
            </Text>
          </VStack>
          <Spacer />
          <Image
            systemName="chevron.up.chevron.down"
            color={
              resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[200]
            }
            size={14}
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
