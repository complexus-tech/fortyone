import React from "react";
import { BottomSheetModal, ContextMenuButton } from "@/components/ui";
import { Text, HStack, Spacer, Button, Image, VStack } from "@expo/ui/swift-ui";
import { colors } from "@/constants";
import { opacity } from "@expo/ui/swift-ui/modifiers";

export const OptionsSheet = ({
  isOpened,
  setIsOpened,
}: {
  isOpened: boolean;
  setIsOpened: (isOpened: boolean) => void;
}) => {
  const statuses = [
    {
      label: "Status",
      onPress: () => {},
    },
    {
      label: "Priority",
      onPress: () => {},
    },
    {
      label: "Assignee",
      onPress: () => {},
    },
  ];
  const ordering = [
    {
      label: "Created",
      onPress: () => {},
    },

    {
      label: "Updated",
      onPress: () => {},
    },
    {
      label: "Deadline",
      onPress: () => {},
    },
    {
      label: "Priority",
      onPress: () => {},
    },
  ];
  const orderingDirection = [
    {
      label: "Descending",
      onPress: () => {},
    },
    {
      label: "Ascending",
      onPress: () => {},
    },
  ];
  const displayColumns = [
    {
      label: "ID",
      onPress: () => {},
    },
    {
      label: "Status",
      onPress: () => {},
    },
    {
      label: "Assignee",
      onPress: () => {},
    },
    {
      label: "Priority",
      onPress: () => {},
    },
  ];

  return (
    <BottomSheetModal
      isOpen={isOpened}
      onClose={() => setIsOpened(false)}
      spacing={28}
    >
      <VStack spacing={20}>
        <HStack>
          <Text>Grouping</Text>
          <Spacer />
          <ContextMenuButton actions={statuses} withNoHost>
            <HStack spacing={3}>
              <Text color={colors.dark.DEFAULT}>Status</Text>
              <Image
                systemName="chevron.up.chevron.down"
                modifiers={[opacity(0.6)]}
                color={colors.dark.DEFAULT}
                size={14}
              />
            </HStack>
          </ContextMenuButton>
        </HStack>
        <HStack>
          <Text>Ordering</Text>
          <Spacer />
          <ContextMenuButton actions={ordering} withNoHost>
            <HStack spacing={3}>
              <Text color={colors.dark.DEFAULT}>Created</Text>
              <Image
                systemName="chevron.up.chevron.down"
                modifiers={[opacity(0.6)]}
                color={colors.dark.DEFAULT}
                size={14}
              />
            </HStack>
          </ContextMenuButton>
        </HStack>
        <HStack>
          <Text>Order direction</Text>
          <Spacer />
          <ContextMenuButton actions={orderingDirection} withNoHost>
            <HStack spacing={3}>
              <Text color={colors.dark.DEFAULT}>Descending</Text>
              <Image
                systemName="chevron.up.chevron.down"
                modifiers={[opacity(0.6)]}
                color={colors.dark.DEFAULT}
                size={14}
              />
            </HStack>
          </ContextMenuButton>
        </HStack>
      </VStack>
      <VStack spacing={16} alignment="leading">
        <Text modifiers={[opacity(0.75)]}>Display columns</Text>
        <HStack spacing={12}>
          <Button variant="bordered" color={colors.black}>
            ID
          </Button>
          <Button variant="bordered" color={colors.black}>
            Status
          </Button>
          <Button variant="plain">Assignee</Button>
          <Button variant="plain">Priority</Button>
        </HStack>
      </VStack>
    </BottomSheetModal>
  );
};
