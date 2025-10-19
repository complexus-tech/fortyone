import React from "react";
import { BottomSheetModal } from "./bottom-sheet-modal";
import { ContextMenuButton } from "./context-menu-button";
import { View, Text as RNText, Pressable } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { colors } from "@/constants";
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
      isOpen={isOpened}
      onClose={() => setIsOpened(false)}
      spacing={40}
    >
      <View style={{ gap: 20 }}>
        <View
          style={{
            flexDirection: "row",
            alignItems: "center",
            justifyContent: "space-between",
          }}
        >
          <RNText style={{ fontSize: 16, fontWeight: "500" }}>Grouping</RNText>
          <ContextMenuButton actions={groupByOptions} withNoHost>
            <View
              style={{ flexDirection: "row", alignItems: "center", gap: 3 }}
            >
              <RNText
                style={{
                  color:
                    resolvedTheme === "light"
                      ? colors.dark.DEFAULT
                      : colors.gray[200],
                  fontSize: 16,
                }}
              >
                {viewOptions.groupBy === "status"
                  ? "Status"
                  : viewOptions.groupBy === "priority"
                    ? "Priority"
                    : "Assignee"}
              </RNText>
              <Ionicons
                name="chevron-up-down"
                size={11}
                color={
                  resolvedTheme === "light"
                    ? colors.dark.DEFAULT
                    : colors.gray[200]
                }
                style={{ opacity: 0.6 }}
              />
            </View>
          </ContextMenuButton>
        </View>
        <View
          style={{
            flexDirection: "row",
            alignItems: "center",
            justifyContent: "space-between",
          }}
        >
          <RNText style={{ fontSize: 16, fontWeight: "500" }}>Ordering</RNText>
          <ContextMenuButton actions={orderByOptions} withNoHost>
            <View
              style={{ flexDirection: "row", alignItems: "center", gap: 3 }}
            >
              <RNText
                style={{
                  color:
                    resolvedTheme === "light"
                      ? colors.dark.DEFAULT
                      : colors.gray[200],
                  fontSize: 16,
                }}
              >
                {viewOptions.orderBy === "created"
                  ? "Created"
                  : viewOptions.orderBy === "updated"
                    ? "Updated"
                    : viewOptions.orderBy === "deadline"
                      ? "Deadline"
                      : "Priority"}
              </RNText>
              <Ionicons
                name="chevron-up-down"
                size={11}
                color={
                  resolvedTheme === "light"
                    ? colors.dark.DEFAULT
                    : colors.gray[200]
                }
                style={{ opacity: 0.6 }}
              />
            </View>
          </ContextMenuButton>
        </View>
        <View
          style={{
            flexDirection: "row",
            alignItems: "center",
            justifyContent: "space-between",
          }}
        >
          <RNText style={{ fontSize: 16, fontWeight: "500" }}>
            Order direction
          </RNText>
          <ContextMenuButton actions={orderDirectionOptions} withNoHost>
            <View
              style={{ flexDirection: "row", alignItems: "center", gap: 3 }}
            >
              <RNText
                style={{
                  color:
                    resolvedTheme === "light"
                      ? colors.dark.DEFAULT
                      : colors.gray[200],
                  fontSize: 16,
                }}
              >
                {viewOptions.orderDirection === "desc"
                  ? "Descending"
                  : "Ascending"}
              </RNText>
              <Ionicons
                name="chevron-up-down"
                size={11}
                color={
                  resolvedTheme === "light"
                    ? colors.dark.DEFAULT
                    : colors.gray[200]
                }
                style={{ opacity: 0.6 }}
              />
            </View>
          </ContextMenuButton>
        </View>
      </View>
      <View style={{ gap: 16, alignItems: "flex-start" }}>
        <RNText style={{ opacity: 0.65, fontSize: 15, fontWeight: "500" }}>
          Display columns
        </RNText>
        <View style={{ flexDirection: "row", gap: 12, flexWrap: "wrap" }}>
          <Pressable
            onPress={() => toggleDisplayColumn("ID")}
            style={{
              paddingHorizontal: 12,
              paddingVertical: 6,
              borderRadius: 8,
              borderWidth: displayColumns.includes("ID") ? 1 : 0,
              borderColor:
                resolvedTheme === "light"
                  ? colors.dark.DEFAULT
                  : colors.gray[200],
              backgroundColor: displayColumns.includes("ID")
                ? "transparent"
                : "transparent",
            }}
          >
            <RNText
              style={{
                color:
                  resolvedTheme === "light"
                    ? colors.dark.DEFAULT
                    : colors.gray[200],
                fontSize: 14,
              }}
            >
              ID
            </RNText>
          </Pressable>
          <Pressable
            onPress={() => toggleDisplayColumn("Status")}
            style={{
              paddingHorizontal: 12,
              paddingVertical: 6,
              borderRadius: 8,
              borderWidth: displayColumns.includes("Status") ? 1 : 0,
              borderColor:
                resolvedTheme === "light"
                  ? colors.dark.DEFAULT
                  : colors.gray[200],
              backgroundColor: displayColumns.includes("Status")
                ? "transparent"
                : "transparent",
            }}
          >
            <RNText
              style={{
                color:
                  resolvedTheme === "light"
                    ? colors.dark.DEFAULT
                    : colors.gray[200],
                fontSize: 14,
              }}
            >
              Status
            </RNText>
          </Pressable>
          <Pressable
            onPress={() => toggleDisplayColumn("Assignee")}
            style={{
              paddingHorizontal: 12,
              paddingVertical: 6,
              borderRadius: 8,
              borderWidth: displayColumns.includes("Assignee") ? 1 : 0,
              borderColor:
                resolvedTheme === "light"
                  ? colors.dark.DEFAULT
                  : colors.gray[200],
              backgroundColor: displayColumns.includes("Assignee")
                ? "transparent"
                : "transparent",
            }}
          >
            <RNText
              style={{
                color:
                  resolvedTheme === "light"
                    ? colors.dark.DEFAULT
                    : colors.gray[200],
                fontSize: 14,
              }}
            >
              Assignee
            </RNText>
          </Pressable>
          <Pressable
            onPress={() => toggleDisplayColumn("Priority")}
            style={{
              paddingHorizontal: 12,
              paddingVertical: 6,
              borderRadius: 8,
              borderWidth: displayColumns.includes("Priority") ? 1 : 0,
              borderColor:
                resolvedTheme === "light"
                  ? colors.dark.DEFAULT
                  : colors.gray[200],
              backgroundColor: displayColumns.includes("Priority")
                ? "transparent"
                : "transparent",
            }}
          >
            <RNText
              style={{
                color:
                  resolvedTheme === "light"
                    ? colors.dark.DEFAULT
                    : colors.gray[200],
                fontSize: 14,
              }}
            >
              Priority
            </RNText>
          </Pressable>
        </View>
      </View>
      <View style={{ flexDirection: "row", justifyContent: "flex-end" }}>
        <Pressable
          onPress={resetViewOptions}
          style={{
            paddingHorizontal: 16,
            paddingVertical: 8,
            borderRadius: 8,
            backgroundColor: "rgba(0,0,0,0.1)",
          }}
        >
          <RNText style={{ color: colors.primary, fontSize: 14 }}>
            Reset default
          </RNText>
        </Pressable>
      </View>
    </BottomSheetModal>
  );
};
