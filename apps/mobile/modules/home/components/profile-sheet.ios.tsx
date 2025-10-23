import React, { useState } from "react";
import { Avatar, BottomSheetModal, WorkspaceSwitcher } from "@/components/ui";
import { colors } from "@/constants";
import { HStack, Image, Spacer, Text, VStack } from "@expo/ui/swift-ui";
import { useProfile } from "@/modules/users/hooks/use-profile";
import { useCurrentWorkspace } from "@/lib/hooks";
import { background, clipShape, frame } from "@expo/ui/swift-ui/modifiers";
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

  const handleManageAccount = () => {
    setIsOpened(false);
    router.push("/account");
  };

  return (
    <>
      <BottomSheetModal isOpen={isOpened} onClose={() => setIsOpened(false)}>
        <Text size={14} weight="medium" color={mutedTextColor}>
          {`${user?.email}`}
        </Text>
        <HStack spacing={8} onPress={handleManageAccount}>
          <HStack modifiers={[frame({ width: 40, height: 40 })]}>
            <Avatar
              name={user?.fullName || user?.username}
              src={user?.avatarUrl}
              color="primary"
              className="size-[40px]"
              rounded="xl"
            />
          </HStack>
          <VStack alignment="leading" spacing={2}>
            <Text size={14} weight="semibold" lineLimit={1}>
              Manage Account
            </Text>
            <Text
              size={13.5}
              weight="medium"
              color={mutedTextColor}
              lineLimit={1}
            >
              Manage your account settings
            </Text>
          </VStack>
          <Spacer />
          <Image
            systemName="chevron.right"
            color={
              resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[200]
            }
            size={14}
          />
        </HStack>
        <HStack spacing={8} onPress={() => setIsWorkspaceSwitcherOpened(true)}>
          <Image
            systemName="xmark.circle.fill"
            size={18}
            color={colors.white}
            modifiers={[
              frame({ width: 40, height: 40 }),
              background(colors.danger),
              clipShape("roundedRectangle", 10),
            ]}
          />
          <VStack alignment="leading" spacing={2}>
            <Text size={14} weight="semibold" lineLimit={1}>
              Logout
            </Text>
            <Text
              size={13.5}
              weight="medium"
              color={mutedTextColor}
              lineLimit={1}
            >
              Log out of your account
            </Text>
          </VStack>
          <Spacer />
          <Image
            systemName="chevron.right"
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
