import { BodyContainer } from "@/components/shared/body";
import type { Objective } from "../types";
import { TableHeader } from "./heading";
import { ObjectiveCard } from "./card";

export const ListObjectives = ({
  objectives,
  isInTeam,
}: {
  objectives: Objective[];
  isInTeam?: boolean;
}) => {
  return (
    <BodyContainer className="h-[calc(100vh-3.7rem)]">
      <TableHeader isInTeam={isInTeam} />
      {objectives.map((objective) => (
        <ObjectiveCard key={objective.id} {...objective} isInTeam={isInTeam} />
      ))}
    </BodyContainer>
  );
};
