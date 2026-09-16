import type { StoryFiltersSheetProps } from "./story-filters.types";
import { Pressable, ScrollView } from "react-native";
import { Text } from "@/components/ui/Text";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks/theme";
import { getTeamStoryFilterCount } from "@/modules/teams/stories/team-story-filters";
import { useStoryFilterSections } from "../hooks/use-story-filter-sections";
import { changeStoryFilter } from "./story-filter-selection";
import { WebIcon } from "@/components/icons/web-icon";

type Props = Pick<
  StoryFiltersSheetProps,
  "filters" | "onChange" | "teamId" | "allowAssignee"
>;
export function ActiveStoryFilters(props: Props) {
  const count = getTeamStoryFilterCount(props.filters);
  return count ? <ActiveFiltersContent {...props} /> : null;
}
function ActiveFiltersContent(props: Props) {
  const { resolvedTheme } = useTheme();
  const palette = themeColors[resolvedTheme];
  const sections = useStoryFilterSections({
    ...props,
    isOpen: true,
    onClose: () => {},
  });
  return (
    <ScrollView
      horizontal
      showsHorizontalScrollIndicator={false}
      style={{ flexGrow: 0 }}
      contentContainerStyle={{
        paddingHorizontal: 20,
        paddingBottom: 4,
        gap: 8,
      }}
    >
      {sections
        .filter((section) => section.selectedIds.length)
        .map((section) => (
          <Pressable
            key={section.id}
            accessibilityRole="button"
            accessibilityLabel={`Remove ${section.label.toLowerCase()} filter, ${section.value}`}
            onPress={() =>
              props.onChange((current) =>
                current.teamId === props.filters.teamId
                  ? changeStoryFilter(current, section.id, null)
                  : current,
              )
            }
            style={({ pressed }) => ({
              flexDirection: "row",
              alignItems: "center",
              gap: 6,
              minHeight: 44,
              maxWidth: 240,
              paddingHorizontal: 12,
              borderRadius: 12,
              backgroundColor: palette.stateHover,
              opacity: pressed ? 0.6 : 1,
            })}
          >
            <Text
              fontSize="sm"
              numberOfLines={1}
              style={{ flexShrink: 1 }}
            >{`${section.label}: ${section.value}`}</Text>
            <WebIcon name="close" size={14} color={palette.icon} />
          </Pressable>
        ))}
    </ScrollView>
  );
}
