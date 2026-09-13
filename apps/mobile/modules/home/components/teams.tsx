import { QueryState } from "@/components/ui/query-state";
import React from "react";

import { Row, Text } from "@/components/ui";
import { useTeams } from "@/modules/teams/hooks/use-teams";
import { Team } from "@/modules/home/components/team";
import { TeamsSkeleton } from "./teams-skeleton";
import { View } from "react-native";

export const Teams = () => {
  const { data: teams = [], isPending, error, refetch } = useTeams();

  if (isPending) {
    return <TeamsSkeleton />;
  }
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
    <View>
      <Row asContainer>
        <Text color="muted" fontSize="sm" className="mb-1">
          Your Teams
        </Text>
      </Row>
      <View className="px-3">
        {teams.map((team) => (
          <Team key={team.id} {...team} />
        ))}
      </View>
    </View>
  );
};
