import React from "react";
import { BottomSheetModal } from "./bottom-sheet-modal";
import { ContextMenuButton } from "./context-menu-button";
import { Text, HStack, Spacer, Button, Image, VStack } from "@expo/ui/swift-ui";
import { colors } from "@/constants";
import {
  opacity,
  font,
  foregroundStyle,
  buttonStyle,
  tint,
} from "@expo/ui/swift-ui/modifiers";
import type {
  DisplayColumn,
  StoriesViewOptions,
} from "@/types/stories-view-options";
import { useTheme } from "@/hooks";

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
  const { resolvedTheme } = useTheme();
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
      nativeContent
      isOpen={isOpened}
      onClose={() => setIsOpened(false)}
      spacing={40}
    >
      <VStack spacing={20}>
        <HStack>
          <Text modifiers={[font({ size: 15, weight: "medium" })]}>
            Grouping
          </Text>
          <Spacer />
          <ContextMenuButton actions={groupByOptions} withNoHost>
            <HStack spacing={3}>
              <Text
                modifiers={[
                  font({ size: 15 }),
                  foregroundStyle(
                    resolvedTheme === "light"
                      ? colors.dark.DEFAULT
                      : colors.gray[200],
                  ),
                ]}
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
                  resolvedTheme === "light"
                    ? colors.dark.DEFAULT
                    : colors.gray[200]
                }
                size={11}
              />
            </HStack>
          </ContextMenuButton>
        </HStack>
        <HStack>
          <Text modifiers={[font({ size: 15, weight: "medium" })]}>
            Ordering
          </Text>
          <Spacer />
          <ContextMenuButton actions={orderByOptions} withNoHost>
            <HStack spacing={3}>
              <Text
                modifiers={[
                  font({ size: 15 }),
                  foregroundStyle(
                    resolvedTheme === "light"
                      ? colors.dark.DEFAULT
                      : colors.gray[200],
                  ),
                ]}
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
                  resolvedTheme === "light"
                    ? colors.dark.DEFAULT
                    : colors.gray[200]
                }
                size={11}
              />
            </HStack>
          </ContextMenuButton>
        </HStack>
        <HStack>
          <Text modifiers={[font({ size: 15, weight: "medium" })]}>
            Order direction
          </Text>
          <Spacer />
          <ContextMenuButton actions={orderDirectionOptions} withNoHost>
            <HStack spacing={3}>
              <Text
                modifiers={[
                  font({ size: 15 }),
                  foregroundStyle(
                    resolvedTheme === "light"
                      ? colors.dark.DEFAULT
                      : colors.gray[200],
                  ),
                ]}
              >
                {viewOptions.orderDirection === "desc"
                  ? "Descending"
                  : "Ascending"}
              </Text>
              <Image
                systemName="chevron.up.chevron.down"
                modifiers={[opacity(0.6)]}
                color={
                  resolvedTheme === "light"
                    ? colors.dark.DEFAULT
                    : colors.gray[200]
                }
                size={11}
              />
            </HStack>
          </ContextMenuButton>
        </HStack>
      </VStack>
      <VStack spacing={16} alignment="leading">
        <Text modifiers={[opacity(0.65), font({ size: 15, weight: "medium" })]}>
          Display columns
        </Text>
        <HStack spacing={12}>
          {(["ID", "Status", "Assignee", "Priority"] as const).map((column) => (
            <Button
              key={column}
              label={column}
              modifiers={[
                buttonStyle(
                  displayColumns.includes(column) ? "bordered" : "plain",
                ),
                tint(
                  resolvedTheme === "light"
                    ? colors.dark.DEFAULT
                    : colors.gray[200],
                ),
              ]}
              onPress={() => toggleDisplayColumn(column)}
            />
          ))}
        </HStack>
      </VStack>
      <HStack>
        <Spacer />
        <Button
          label="Reset defaults"
          modifiers={[buttonStyle("bordered"), tint(colors.primary)]}
          onPress={resetViewOptions}
        />
      </HStack>
    </BottomSheetModal>
  );
};
