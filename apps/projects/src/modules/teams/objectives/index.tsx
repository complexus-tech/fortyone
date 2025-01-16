"use client";

import { ObjectiveCard } from "@/components/ui/objective/card";
import type { Objective } from "@/modules/objectives/types";
import { BodyContainer } from "@/components/shared/body";
import { ObjectivesHeader } from "./components/header";

export const ObjectivesList = ({ objectives }: { objectives: Objective[] }) => {
  return (
    <>
      <ObjectivesHeader />
      <BodyContainer className="h-[calc(100vh-3.7rem)]">
        {objectives.map((objective) => (
          <ObjectiveCard key={objective.id} {...objective} />
        ))}
      </BodyContainer>
    </>
  );
};
