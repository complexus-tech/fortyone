"use client";

import { useParams } from "next/navigation";
import { Text, Button, Box } from "ui";
import { CrownIcon } from "icons";
import { FeatureGuard } from "@/components/ui";
import { useTerminology } from "@/hooks";
import { ObjectivesHeader } from "./components/header";
import { ListObjectives } from "./components/list-objectives";
import { TeamObjectivesHeader } from "./components/team-header";
import { useObjectives, useTeamObjectives } from "./hooks/use-objectives";
import { ObjectivesSkeleton } from "./components/objectives-skeleton";

const Guard = () => {
  const { getTermDisplay } = useTerminology();
  return (
    <Box className="flex h-[80%] items-center justify-center">
      <Box className="flex flex-col items-center">
        <CrownIcon className="h-12 text-warning" strokeWidth={1.3} />
        <Text className="mb-6 mt-8" fontSize="3xl">
          Upgrade your plan
        </Text>
        <Text className="mb-6 max-w-md text-center" color="muted">
          Upgrade your plan to create{" "}
          {getTermDisplay("objectiveTerm", {
            variant: "plural",
          })}
          , unlimited{" "}
          {getTermDisplay("objectiveTerm", {
            variant: "plural",
          })}
          , and unlock more premium features.
        </Text>
        <Button color="warning" href="/settings/workspace/billing">
          Upgrade now
        </Button>
      </Box>
    </Box>
  );
};

export const ObjectivesList = () => {
  const { data: objectives = [], isPending } = useObjectives();

  if (isPending) {
    return <ObjectivesSkeleton />;
  }

  return (
    <>
      <ObjectivesHeader />
      <FeatureGuard fallback={<Guard />} feature="objective">
        <ListObjectives objectives={objectives} />
      </FeatureGuard>
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
      <FeatureGuard fallback={<Guard />} feature="objective">
        <ListObjectives isInTeam objectives={objectives} />
      </FeatureGuard>
    </>
  );
};
