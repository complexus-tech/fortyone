import { useState } from "react";
import { Pressable, View } from "react-native";
import { Text } from "@/components/ui";
import { QueryState } from "@/components/ui/query-state";
import { useTeams } from "@/modules/teams/hooks/use-teams";
import { Team } from "./team";
import { TeamsSkeleton } from "./teams-skeleton";
import { TeamSwitcher } from "@/components/ui/team-switcher";

const HOME_TEAM_LIMIT = 5;

export const Teams = () => {
  const { data: teams = [], isPending, error, refetch } = useTeams();
  const [switcherOpen, setSwitcherOpen] = useState(false);
  if (isPending) return <TeamsSkeleton />;
  if (error)
    return (
      <QueryState
        title="Could not load your teams"
        message={error.message}
        onRetry={() => {
          void refetch();
        }}
      />
    );

  return (
    <View style={{ paddingTop: 24 }}>
      <View
        style={{
          flexDirection: "row",
          alignItems: "center",
          gap: 8,
          paddingHorizontal: 20,
          minHeight: 44,
        }}
      >
        <Text
          color="muted"
          style={{ fontSize: 15, lineHeight: 20, fontWeight: "500" }}
        >
          Teams
        </Text>
        <Text color="muted" style={{ flex: 1, fontSize: 15, lineHeight: 20 }}>
          {teams.length}
        </Text>
        {teams.length > HOME_TEAM_LIMIT ? (
          <Pressable
            accessibilityRole="button"
            accessibilityLabel={`View all ${teams.length} teams`}
            onPress={() => setSwitcherOpen(true)}
            style={({ pressed }) => ({
              minHeight: 44,
              justifyContent: "center",
              paddingLeft: 12,
              opacity: pressed ? 0.6 : 1,
            })}
          >
            <Text color="muted" fontSize="sm">
              View all
            </Text>
          </Pressable>
        ) : null}
      </View>
      {teams.length ? (
        // The API returns the user's saved team order. Keep it intact here and
        // pass the full collection to the switcher, including overflow teams.
        teams
          .slice(0, HOME_TEAM_LIMIT)
          .map((team) => <Team key={team.id} {...team} />)
      ) : (
        <Text
          color="muted"
          style={{ paddingHorizontal: 20, paddingVertical: 14 }}
        >
          You haven&apos;t joined a team yet.
        </Text>
      )}
      <TeamSwitcher
        isOpened={switcherOpen}
        setIsOpened={setSwitcherOpen}
        teams={teams}
      />
    </View>
  );
};
