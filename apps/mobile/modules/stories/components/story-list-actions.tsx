import type { StoryListActionsProps } from "./story-list-actions.types";
import { StyleSheet, View } from "react-native";
import { IconButton } from "@/components/ui/icon-button";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks/theme";

export function StoryListActions({
  onFilters,
  onDisplay,
  filterCount,
}: StoryListActionsProps) {
  const { resolvedTheme } = useTheme();
  const palette = themeColors[resolvedTheme];
  return (
    <View
      style={[
        styles.capsule,
        {
          backgroundColor: palette.surfaceElevated,
          borderColor: palette.border,
        },
      ]}
    >
      <IconButton
        icon="filter-outline"
        label={filterCount ? `Filters, ${filterCount} active` : "Filters"}
        onPress={onFilters}
        selected={filterCount > 0}
      />
      <IconButton
        icon="options-outline"
        label="Display options"
        onPress={onDisplay}
      />
    </View>
  );
}
const styles = StyleSheet.create({
  capsule: {
    flexDirection: "row",
    borderRadius: 999,
    borderWidth: StyleSheet.hairlineWidth,
    overflow: "hidden",
  },
});
