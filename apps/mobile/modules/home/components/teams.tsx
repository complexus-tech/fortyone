import React from "react";

import { Text, Section } from "@/components/ui";
import { useTeams } from "@/modules/teams/hooks/use-teams";
import { Team } from "@/modules/home/components/team";

export const Teams = () => {
  const { data: teams = [], isPending } = useTeams();
  if (isPending) {
    return <Text>Loading...</Text>;
  }
  return (
    <Section title="Your Teams">
      {teams?.map((team) => <Team key={team.id} {...team} />)}
    </Section>
  );
};
