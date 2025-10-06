import React from "react";
import { BottomSheetModal, ContextMenuButton } from "@/components/ui";
import { Text, HStack, Spacer, Button, Image, VStack } from "@expo/ui/swift-ui";
import { colors } from "@/constants";
import { opacity } from "@expo/ui/swift-ui/modifiers";
import type {
  DisplayColumn,
  StoriesViewOptions,
} from "@/types/stories-view-options";
import { useColorScheme } from "nativewind";

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
  setViewOptions: (options: Partial<StoriesViewOptions>) => void;
  resetViewOptions: () => void;
}) => {
  const { colorScheme } = useColorScheme();
  const displayColumns = viewOptions.displayColumns || [];
  const groupByOptions = [
    {
      label: "Status",
      onPress: () => setViewOptions({ groupBy: "status" }),
    },
    {
      label: "Priority",
      onPress: () => setViewOptions({ groupBy: "priority" }),
    },
    {
      label: "Assignee",
      onPress: () => setViewOptions({ groupBy: "assignee" }),
    },
  ];

  const orderByOptions = [
    {
      label: "Created",
      onPress: () => setViewOptions({ orderBy: "created" }),
    },
    {
      label: "Updated",
      onPress: () => setViewOptions({ orderBy: "updated" }),
    },
    {
      label: "Deadline",
      onPress: () => setViewOptions({ orderBy: "deadline" }),
    },
    {
      label: "Priority",
      onPress: () => setViewOptions({ orderBy: "priority" }),
    },
  ];

  const orderDirectionOptions = [
    {
      label: "Descending",
      onPress: () => setViewOptions({ orderDirection: "desc" }),
    },
    {
      label: "Ascending",
      onPress: () => setViewOptions({ orderDirection: "asc" }),
    },
  ];

  const toggleDisplayColumn = (column: DisplayColumn) => {
    setViewOptions({
      displayColumns: displayColumns.includes(column)
        ? displayColumns.filter((c) => c !== column)
        : [...displayColumns, column],
    });
  };

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
              <Text
                color={
                  colorScheme === "light"
                    ? colors.dark.DEFAULT
                    : colors.gray[200]
                }
              >
                {viewOptions.groupBy === "status"
                  ? "Status"
                  : viewOptions.groupBy === "priority"
                    ? "Priority"
                    : "Assignee"}
              </Text>
              <Image
                systemName="chevron.up.chevron.down"
                modifiers={[opacity(0.6)]}
                color={
                  colorScheme === "light"
                    ? colors.dark.DEFAULT
                    : colors.gray[200]
                }
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
              <Text
                color={
                  colorScheme === "light"
                    ? colors.dark.DEFAULT
                    : colors.gray[200]
                }
              >
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
                color={
                  colorScheme === "light"
                    ? colors.dark.DEFAULT
                    : colors.gray[200]
                }
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
              <Text
                color={
                  colorScheme === "light"
                    ? colors.dark.DEFAULT
                    : colors.gray[200]
                }
              >
                {viewOptions.orderDirection === "desc"
                  ? "Descending"
                  : "Ascending"}
              </Text>
              <Image
                systemName="chevron.up.chevron.down"
                modifiers={[opacity(0.6)]}
                color={
                  colorScheme === "light"
                    ? colors.dark.DEFAULT
                    : colors.gray[200]
                }
                size={14}
              />
            </HStack>
          </ContextMenuButton>
        </HStack>
      </VStack>
      <VStack spacing={16} alignment="leading">
        <Text modifiers={[opacity(0.75)]}>Display columns</Text>
        <HStack spacing={12}>
          <Button
            variant={displayColumns.includes("ID") ? "bordered" : "plain"}
            color={
              colorScheme === "light" ? colors.dark.DEFAULT : colors.gray[200]
            }
            onPress={() => toggleDisplayColumn("ID")}
          >
            ID
          </Button>
          <Button
            variant={displayColumns.includes("Status") ? "bordered" : "plain"}
            color={
              colorScheme === "light" ? colors.dark.DEFAULT : colors.gray[200]
            }
            onPress={() => toggleDisplayColumn("Status")}
          >
            Status
          </Button>
          <Button
            color={
              colorScheme === "light" ? colors.dark.DEFAULT : colors.gray[200]
            }
            variant={displayColumns.includes("Assignee") ? "bordered" : "plain"}
            onPress={() => toggleDisplayColumn("Assignee")}
          >
            Assignee
          </Button>
          <Button
            color={
              colorScheme === "light" ? colors.dark.DEFAULT : colors.gray[200]
            }
            variant={displayColumns.includes("Priority") ? "bordered" : "plain"}
            onPress={() => toggleDisplayColumn("Priority")}
          >
            Priority
          </Button>
        </HStack>
      </VStack>
      <HStack>
        <Spacer />
        <Button variant="glass" color="primary" onPress={resetViewOptions}>
          Reset default
        </Button>
      </HStack>
    </BottomSheetModal>
  );
};
