"use client";

import { useParams } from "next/navigation";
import { ObjectivesHeader } from "./components/header";
import { ListObjectives } from "./components/list-objectives";
import { TeamObjectivesHeader } from "./components/team-header";
import { useObjectives, useTeamObjectives } from "./hooks/use-objectives";

export const ObjectivesList = () => {
  const { data: objectives = [] } = useObjectives();
  return (
    <>
      <ObjectivesHeader />
      <ListObjectives objectives={objectives} />
    </>
  );
};

export const TeamObjectivesList = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const { data: objectives = [] } = useTeamObjectives(teamId);
  return (
    <>
      <TeamObjectivesHeader />
      <ListObjectives isInTeam objectives={objectives} />
    </>
  );
};
