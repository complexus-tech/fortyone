"use client";

import { ObjectiveCard } from "@/components/ui/objective/card";
import { Objective } from "@/modules/objectives/types";
import { BodyContainer } from "@/components/shared/body";
import { ObjectivesHeader } from "./components/header";
import { Heading } from "./components/heading";

export const ObjectivesList = ({ objectives }: { objectives: Objective[] }) => {
  return (
    <>
      <ObjectivesHeader />
      <Heading />
      <BodyContainer className="h-[calc(100vh-7.2rem)]">
        {objectives.map((objective) => (
          <ObjectiveCard key={objective.id} {...objective} />
        ))}
      </BodyContainer>
    </>
  );
};
