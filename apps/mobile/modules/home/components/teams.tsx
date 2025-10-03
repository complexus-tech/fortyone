import React from "react";

import { Row, Text } from "@/components/ui";
import { useTeams } from "@/modules/teams/hooks/use-teams";
import { Team } from "@/modules/home/components/team";
import { TeamsSkeleton } from "./teams-skeleton";
import { View } from "react-native";

export const Teams = () => {
  const { data: teams = [], isPending } = useTeams();

  if (isPending) {
    return <TeamsSkeleton />;
  }
  return (
    <View>
      <Row asContainer>
        <Text color="muted" className="mb-1">
          Your Teams
        </Text>
      </Row>
      <View className="px-3">
        {teams?.map((team) => <Team key={team.id} {...team} />)}
      </View>
    </View>
  );
};
