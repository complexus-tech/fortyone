import type { StoriesViewOptions } from "@/types/stories-view-options";
import { Pressable, StyleSheet, View } from "react-native";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks/theme";
import { Text } from "./Text";

export type StoriesOptionsSheetProps = {
  isOpened: boolean;
  setIsOpened: (isOpened: boolean) => void;
  viewOptions: StoriesViewOptions;
  setViewOptions: (options: Partial<StoriesViewOptions>) => void;
  resetViewOptions: () => void;
};

const GROUPING_OPTIONS = [
  { value: "status", label: "Status" },
  { value: "priority", label: "Priority" },
  { value: "assignee", label: "Assignee" },
] as const;

const ORDERING_OPTIONS = [
  { value: "created", label: "Created" },
  { value: "updated", label: "Updated" },
  { value: "deadline", label: "Deadline" },
  { value: "priority", label: "Priority" },
] as const;

const DIRECTION_OPTIONS = [
  { value: "desc", label: "Descending" },
  { value: "asc", label: "Ascending" },
] as const;

export function getStoriesOptionRows({
  viewOptions,
  setViewOptions,
}: Pick<StoriesOptionsSheetProps, "viewOptions" | "setViewOptions">) {
  return [
    {
      id: "groupBy",
      label: "Grouping",
      value: GROUPING_OPTIONS.find(
        (option) => option.value === viewOptions.groupBy,
      )?.label,
      actions: GROUPING_OPTIONS.map((option) => ({
        label: option.label,
        selected: option.value === viewOptions.groupBy,
        onPress: () => setViewOptions({ groupBy: option.value }),
      })),
    },
    {
      id: "orderBy",
      label: "Ordering",
      value: ORDERING_OPTIONS.find(
        (option) => option.value === viewOptions.orderBy,
      )?.label,
      actions: ORDERING_OPTIONS.map((option) => ({
        label: option.label,
        selected: option.value === viewOptions.orderBy,
        onPress: () => setViewOptions({ orderBy: option.value }),
      })),
    },
    {
      id: "orderDirection",
      label: "Order direction",
      value: DIRECTION_OPTIONS.find(
        (option) => option.value === viewOptions.orderDirection,
      )?.label,
      actions: DIRECTION_OPTIONS.map((option) => ({
        label: option.label,
        selected: option.value === viewOptions.orderDirection,
        onPress: () => setViewOptions({ orderDirection: option.value }),
      })),
    },
  ];
}

export function StoriesDisplayOptions({
  viewOptions,
  setViewOptions,
  resetViewOptions,
}: Pick<
  StoriesOptionsSheetProps,
  "viewOptions" | "setViewOptions" | "resetViewOptions"
>) {
  const { resolvedTheme } = useTheme();
  const theme = themeColors[resolvedTheme];
  const displayColumns = viewOptions.displayColumns || [];

  return (
    <View style={{ gap: 12 }}>
      <View style={styles.sectionHeader}>
        <Text accessibilityRole="header" style={{ flex: 1 }}>
          Row properties
        </Text>
        <Pressable
          accessibilityRole="button"
          accessibilityLabel="Reset view options"
          onPress={resetViewOptions}
          style={({ pressed }) => [
            styles.reset,
            { opacity: pressed ? 0.6 : 1 },
          ]}
        >
          <Text fontSize="sm" color="muted">
            Reset
          </Text>
        </Pressable>
      </View>
      <View style={styles.chips}>
        {(["Status", "Priority", "Assignee", "ID"] as const).map((column) => {
          const selected = displayColumns.includes(column);
          return (
            <Pressable
              key={column}
              accessibilityRole="checkbox"
              accessibilityLabel={`Show ${column === "ID" ? "identifier" : column.toLowerCase()}`}
              accessibilityState={{ checked: selected }}
              onPress={() =>
                setViewOptions({
                  displayColumns: selected
                    ? displayColumns.filter((value) => value !== column)
                    : [...displayColumns, column],
                })
              }
              style={({ pressed }) => [
                styles.chip,
                {
                  opacity: pressed ? 0.6 : 1,
                  borderColor: selected ? theme.borderStrong : theme.border,
                  backgroundColor: selected ? theme.stateHover : "transparent",
                },
              ]}
            >
              <Text fontSize="sm" fontWeight="normal">
                {column}
              </Text>
            </Pressable>
          );
        })}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  sectionHeader: {
    flexDirection: "row",
    alignItems: "center",
    gap: 12,
    minHeight: 44,
  },
  reset: {
    minWidth: 44,
    minHeight: 44,
    justifyContent: "center",
    alignItems: "flex-end",
  },
  chips: { flexDirection: "row", flexWrap: "wrap", gap: 8 },
  chip: {
    minHeight: 44,
    maxWidth: "100%",
    paddingHorizontal: 16,
    paddingVertical: 10,
    borderRadius: 999,
    borderWidth: StyleSheet.hairlineWidth,
    alignItems: "center",
    justifyContent: "center",
  },
});
