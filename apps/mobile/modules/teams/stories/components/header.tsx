import type { StoriesViewOptions } from "@/types/stories-view-options";
import { useState } from "react";
import { Pressable, View } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useLocalSearchParams } from "expo-router";
import { Back, Text } from "@/components/ui";
import { TeamSwitcher } from "@/components/ui/team-switcher";
import { StoryOptionsButton } from "@/modules/stories/components";
import { StoryListActions } from "@/modules/stories/components/story-list-actions";
import { useTeams } from "@/modules/teams/hooks/use-teams";
import { useTheme } from "@/hooks/theme";
import { themeColors } from "@/constants/colors";

type HeaderProps = {
  viewOptions: StoriesViewOptions;
  setViewOptions: (options: Partial<StoriesViewOptions>) => void;
  resetViewOptions: () => void;
  onFilters: () => void;
  filterCount: number;
};

export const Header = ({
  onFilters,
  filterCount,
  ...displayProps
}: HeaderProps) => {
  const { teamId } = useLocalSearchParams<{ teamId: string }>();
  const { data: teams = [] } = useTeams();
  const [switcherOpen, setSwitcherOpen] = useState(false);
  const team = teams.find((item) => item.id === teamId);
  const { resolvedTheme } = useTheme();
  const palette = themeColors[resolvedTheme];
  return (
    <>
      <View
        style={{
          paddingHorizontal: 20,
          paddingBottom: 12,
          flexDirection: "row",
          alignItems: "center",
          gap: 8,
          minHeight: 56,
        }}
      >
        <Back />
        <Pressable
          accessibilityRole={teams.length > 1 ? "button" : "header"}
          accessibilityLabel={
            teams.length > 1
              ? `${team?.name ?? "Team"}, switch team`
              : team?.name ?? "Team"
          }
          disabled={teams.length < 2}
          onPress={() => setSwitcherOpen(true)}
          style={({ pressed }) => ({
            flex: 1,
            minWidth: 0,
            minHeight: 44,
            justifyContent: "center",
            opacity: pressed ? 0.6 : 1,
          })}
        >
          <View style={{ flexDirection: "row", alignItems: "center", gap: 6 }}>
            <Text
              fontSize="xl"
              fontWeight="bold"
              numberOfLines={1}
              style={{ flexShrink: 1 }}
            >
              {team?.name ?? "Team"}
            </Text>
            {teams.length > 1 ? (
              <Ionicons name="chevron-down" size={18} color={palette.icon} />
            ) : null}
          </View>
        </Pressable>
        <StoryOptionsButton
          {...displayProps}
          renderTrigger={(onDisplay) => (
            <StoryListActions
              onFilters={onFilters}
              onDisplay={onDisplay}
              filterCount={filterCount}
            />
          )}
        />
      </View>
      <TeamSwitcher
        isOpened={switcherOpen}
        setIsOpened={setSwitcherOpen}
        teams={teams}
        currentTeamId={teamId}
      />
    </>
  );
};
