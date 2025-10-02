import React from "react";

import { Section } from "@/components/ui";
import { useTeams } from "@/modules/teams/hooks/use-teams";
import { Team } from "@/modules/home/components/team";
import { TeamsSkeleton } from "./teams-skeleton";

export const Teams = () => {
  const { data: teams = [], isPending } = useTeams();
  if (isPending) {
    return <TeamsSkeleton />;
  }
  return (
    <Section title="Your Teams">
      {teams?.map((team) => <Team key={team.id} {...team} />)}
    </Section>
  );
};
