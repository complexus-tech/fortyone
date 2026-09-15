import type { StoryListActionsProps } from "./story-list-actions.types";
import { Button, Host, HStack, Image } from "@expo/ui/swift-ui";
import {
  accessibilityLabel,
  buttonStyle,
  contentShape,
  frame,
  glassEffect,
  shapes,
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
          <Image
            systemName="line.3.horizontal.decrease"
            size={22}
            color={filterCount ? colors.primary : palette.foreground}
            modifiers={[
              frame({ width: 44, height: 44 }),
              contentShape(shapes.rectangle()),
            ]}
          />
        </Button>
        <Button
          onPress={onDisplay}
          modifiers={[
            buttonStyle("plain"),
            accessibilityLabel("Display options"),
          ]}
        >
          <Image
            systemName="slider.horizontal.3"
            size={22}
            color={palette.foreground}
            modifiers={[
              frame({ width: 44, height: 44 }),
              contentShape(shapes.rectangle()),
            ]}
          />
        </Button>
      </HStack>
    </Host>
  );
}
