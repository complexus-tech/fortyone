"use client";

import type { Objective } from "@/modules/objectives/types";
import { ObjectivesHeader } from "./components/header";
import { ListObjectives } from "./components/list-objectives";
import { TeamObjectivesHeader } from "./components/team-header";

export const ObjectivesList = ({ objectives }: { objectives: Objective[] }) => {
  return (
    <>
      <ObjectivesHeader />
      <ListObjectives objectives={objectives} />
    </>
  );
};

export const TeamObjectivesList = ({
  objectives,
}: {
  objectives: Objective[];
}) => {
  return (
    <>
      <TeamObjectivesHeader />
      <ListObjectives isInTeam objectives={objectives} />
    </>
  );
};
