"use client";

import { useParams } from "next/navigation";
import { ObjectivesHeader } from "./components/header";
import { ListObjectives } from "./components/list-objectives";
import { TeamObjectivesHeader } from "./components/team-header";
import { useObjectives, useTeamObjectives } from "./hooks/use-objectives";
import { ObjectivesSkeleton } from "./components/objectives-skeleton";

export const ObjectivesList = () => {
  const { data: objectives = [], isPending } = useObjectives();

  if (isPending) {
    return <ObjectivesSkeleton />;
  }

  return (
    <>
      <ObjectivesHeader />
      <ListObjectives objectives={objectives} />
    </>
  );
};

export const TeamObjectivesList = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const { data: objectives = [], isPending } = useTeamObjectives(teamId);

  if (isPending) {
    return <ObjectivesSkeleton isInTeam />;
  }

  return (
    <>
      <TeamObjectivesHeader />
      <ListObjectives isInTeam objectives={objectives} />
    </>
  );
};
