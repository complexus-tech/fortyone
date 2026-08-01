"use client";

import Link from "next/link";
import { Box, Button, Divider, Flex, Text } from "ui";
import { ChevronRightIcon, CloseIcon, ObjectiveIcon } from "icons";
import { useWorkspacePath } from "@/hooks";
import { useObjectiveAnalytics } from "@/modules/objectives/hooks/objective-analytics";
import { useKeyResults } from "@/modules/objectives/hooks/use-key-results";
import { useObjective } from "@/modules/objectives/hooks/use-objective";
import type { Objective } from "@/modules/objectives/types";
import { ProgressChart } from "@/modules/objectives/stories/progress-chart";
import { ObjectiveDetailsProperties } from "./objective-details-properties";
import { ObjectiveDetailsKeyResults } from "./objective-details-key-results";

const EMPTY_PROGRESS_DATA: [] = [];

export const RoadmapObjectiveDetails = ({
  objective: initialObjective,
  onClose,
}: {
  objective: Objective;
  onClose: () => void;
}) => {
  const { withWorkspace } = useWorkspacePath();
  const { data: fetchedObjective } = useObjective(initialObjective.id);
  const objective = fetchedObjective ?? initialObjective;
  const { data: keyResults = [] } = useKeyResults(initialObjective.id);
  const { data: analytics } = useObjectiveAnalytics(initialObjective.id);
  const objectiveHref = withWorkspace(
    `/teams/${objective.teamId}/objectives/${objective.id}`,
  );

  return (
    <Box className="border-border/70 dark:border-border dark:bg-surface absolute top-12 right-3 bottom-6 isolate z-40 w-[calc(100%-1.5rem)] overflow-y-auto rounded-xl border bg-white shadow-xl md:w-[34rem]">
      <Flex
        align="center"
        className="border-border/70 dark:border-border dark:bg-surface/80 sticky top-0 z-10 min-h-16 gap-6 border-b-[0.5px] bg-white/80 px-6 backdrop-blur-2xl"
        justify="between"
      >
        <Link
          className="group flex min-w-0 flex-1 items-center gap-3"
          href={objectiveHref}
        >
          <ObjectiveIcon className="text-text-muted h-5 shrink-0" />
          <Flex align="center" className="min-w-0" gap={1}>
            <Text className="truncate" fontSize="lg" fontWeight="semibold">
              {objective.name}
            </Text>
            <ChevronRightIcon className="text-text-muted h-4 shrink-0 transition-transform group-hover:translate-x-0.5" />
          </Flex>
        </Link>
        <Button
          aria-label="Close objective details"
          className="-mr-2"
          color="tertiary"
          leftIcon={<CloseIcon className="h-4" strokeWidth={3} />}
          onClick={onClose}
          size="sm"
          variant="naked"
        />
      </Flex>

      <Box className="px-6 pt-5 pb-24">
        {objective.shortSummary ? (
          <Text className="mb-5 line-clamp-3 leading-6" color="muted">
            {objective.shortSummary}
          </Text>
        ) : null}

        <ObjectiveDetailsProperties objective={objective} />

        <Divider className="my-5" />

        <Text className="mb-3">Progress</Text>
        <ProgressChart
          progressData={analytics?.progressChart ?? EMPTY_PROGRESS_DATA}
        />
        <Divider className="my-5" />

        <ObjectiveDetailsKeyResults keyResults={keyResults} />
      </Box>
    </Box>
  );
};
