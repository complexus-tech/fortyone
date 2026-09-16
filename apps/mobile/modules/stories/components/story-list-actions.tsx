import type { StoryListActionsProps } from "./story-list-actions.types";
import { FilterIcon } from "@/components/icons/filter";
import { PreferencesIcon } from "@/components/icons/preferences";
import { StyleSheet, View } from "react-native";
import { IconButton } from "@/components/ui/icon-button";
import { colors, themeColors } from "@/constants/colors";
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
        label={filterCount ? `Filters, ${filterCount} active` : "Filters"}
        onPress={onFilters}
        selected={filterCount > 0}
      >
        <FilterIcon
          size={22}
          color={filterCount ? colors.primary : palette.foreground}
        />
      </IconButton>
      <IconButton label="Display options" onPress={onDisplay}>
        <PreferencesIcon size={22} color={palette.foreground} />
      </IconButton>
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
