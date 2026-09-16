import type { StoryListActionsProps } from "./story-list-actions.types";
import { FilterIcon } from "@/components/icons/filter";
import { PreferencesIcon } from "@/components/icons/preferences";
import { NativeToolbarIcon } from "@/components/ui/native-toolbar-icon";
import { Button, Host, HStack } from "@expo/ui/swift-ui";
import {
  accessibilityLabel,
  buttonStyle,
  glassEffect,
} from "@expo/ui/swift-ui/modifiers";
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
    <Host
      matchContents
      colorScheme={resolvedTheme}
      style={{ width: 88, height: 44 }}
    >
      <HStack
        spacing={0}
        modifiers={[
          glassEffect({
            glass: { variant: "regular", interactive: true },
            shape: "capsule",
          }),
        ]}
      >
        <Button
          onPress={onFilters}
          modifiers={[
            buttonStyle("plain"),
            accessibilityLabel(
              filterCount ? `Filters, ${filterCount} active` : "Filters",
            ),
          ]}
        >
          <NativeToolbarIcon>
            <FilterIcon
              size={22}
              color={filterCount ? colors.primary : palette.foreground}
            />
          </NativeToolbarIcon>
        </Button>
        <Button
          onPress={onDisplay}
          modifiers={[
            buttonStyle("plain"),
            accessibilityLabel("Display options"),
          ]}
        >
          <NativeToolbarIcon>
            <PreferencesIcon size={22} color={palette.foreground} />
          </NativeToolbarIcon>
        </Button>
      </HStack>
    </Host>
  );
}
