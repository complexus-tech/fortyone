"use client";

import { useParams } from "next/navigation";
import { ObjectivesHeader } from "./components/header";
import { ListObjectives } from "./components/list-objectives";
import { TeamObjectivesHeader } from "./components/team-header";
import { useObjectives, useTeamObjectives } from "./hooks/use-objectives";
import { ObjectivesSkeleton } from "./components/objectives-skeleton";

export const ObjectivesList = () => {
  const { data: objectives = [], isPending } = useObjectives();

  return (
    <>
      <ObjectivesHeader />
      {isPending ? (
        <ObjectivesSkeleton />
      ) : (
        <ListObjectives objectives={objectives} />
      )}
    </>
  );
};

export const TeamObjectivesList = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const { data: objectives = [], isPending } = useTeamObjectives(teamId);

  return (
    <>
      <TeamObjectivesHeader />
      {isPending ? (
        <ObjectivesSkeleton isInTeam />
      ) : (
        <ListObjectives isInTeam objectives={objectives} />
      )}
    </>
  );
};
