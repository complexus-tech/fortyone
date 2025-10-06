import React from "react";
import { BottomSheetModal, ContextMenuButton } from "@/components/ui";
import { Text, HStack, Spacer, Button, Image, VStack } from "@expo/ui/swift-ui";
import { colors } from "@/constants";
import { opacity } from "@expo/ui/swift-ui/modifiers";
import type { StoriesViewOptions } from "@/types/stories-view-options";

export const StoriesOptionsSheet = ({
  isOpened,
  setIsOpened,
  viewOptions,
  setViewOptions,
  resetViewOptions,
}: {
  isOpened: boolean;
  setIsOpened: (isOpened: boolean) => void;
  viewOptions: StoriesViewOptions;
  setViewOptions: (options: StoriesViewOptions) => void;
  resetViewOptions: () => void;
}) => {
  const groupByOptions = [
    {
      label: "Status",
      onPress: () => setViewOptions({ ...viewOptions, groupBy: "status" }),
    },
    {
      label: "Priority",
      onPress: () => setViewOptions({ ...viewOptions, groupBy: "priority" }),
    },
    {
      label: "Assignee",
      onPress: () => setViewOptions({ ...viewOptions, groupBy: "assignee" }),
    },
  ];

  const orderByOptions = [
    {
      label: "Created",
      onPress: () => setViewOptions({ ...viewOptions, orderBy: "created" }),
    },
    {
      label: "Updated",
      onPress: () => setViewOptions({ ...viewOptions, orderBy: "updated" }),
    },
    {
      label: "Deadline",
      onPress: () => setViewOptions({ ...viewOptions, orderBy: "deadline" }),
    },
    {
      label: "Priority",
      onPress: () => setViewOptions({ ...viewOptions, orderBy: "priority" }),
    },
  ];

  const orderDirectionOptions = [
    {
      label: "Descending",
      onPress: () => setViewOptions({ ...viewOptions, orderDirection: "desc" }),
    },
    {
      label: "Ascending",
      onPress: () => setViewOptions({ ...viewOptions, orderDirection: "asc" }),
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
          <ContextMenuButton actions={groupByOptions} withNoHost>
            <HStack spacing={3}>
              <Text color={colors.dark.DEFAULT}>
                {viewOptions.groupBy === "status"
                  ? "Status"
                  : viewOptions.groupBy === "priority"
                    ? "Priority"
                    : "Assignee"}
              </Text>
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
          <ContextMenuButton actions={orderByOptions} withNoHost>
            <HStack spacing={3}>
              <Text color={colors.dark.DEFAULT}>
                {viewOptions.orderBy === "created"
                  ? "Created"
                  : viewOptions.orderBy === "updated"
                    ? "Updated"
                    : viewOptions.orderBy === "deadline"
                      ? "Deadline"
                      : "Priority"}
              </Text>
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
          <ContextMenuButton actions={orderDirectionOptions} withNoHost>
            <HStack spacing={3}>
              <Text color={colors.dark.DEFAULT}>
                {viewOptions.orderDirection === "desc"
                  ? "Descending"
                  : "Ascending"}
              </Text>
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
