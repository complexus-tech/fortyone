"use client";

import Link from "next/link";
import { useEffect, useRef } from "react";
import { Box, Button, Divider, Flex, Text } from "ui";
import { ChevronRightIcon, CloseIcon, ObjectiveIcon } from "icons";
import { useWorkspacePath } from "@/hooks";
import { useObjectiveAnalytics } from "@/modules/objectives/hooks/objective-analytics";
import { useKeyResults } from "@/modules/objectives/hooks/use-key-results";
import { useObjective } from "@/modules/objectives/hooks/use-objective";
import type { KeyResult, Objective } from "@/modules/objectives/types";
import { ObjectiveForecastRiskBanner } from "@/modules/objectives/components/objective-forecast-risk";
import { ProgressChart } from "@/modules/objectives/stories/progress-chart";
import { ObjectiveDetailsProperties } from "./objective-details-properties";
import { ObjectiveDetailsKeyResults } from "./objective-details-key-results";

const EMPTY_PROGRESS_DATA: [] = [];

export const RoadmapObjectiveDetails = ({
  objective: initialObjective,
  onClose,
  onKeyResultSelect,
}: {
  objective: Objective;
  onClose: () => void;
  onKeyResultSelect?: (keyResult: KeyResult) => void;
}) => {
  const { withWorkspace } = useWorkspacePath();
  const { data: fetchedObjective } = useObjective(initialObjective.id);
  const objective = fetchedObjective ?? initialObjective;
  const { data: keyResults = [] } = useKeyResults(initialObjective.id);
  const { data: analytics } = useObjectiveAnalytics(initialObjective.id);
  const objectiveHref = withWorkspace(
    `/teams/${objective.teamId}/objectives/${objective.id}`,
  );
  const panelRef = useRef<HTMLDivElement>(null);
  const panelId = `objective-details-${initialObjective.id}`;

  useEffect(() => {
    const panel = panelRef.current;
    const trigger = document.activeElement;
    panel?.focus({ preventScroll: true });

    return () => {
      if (
        trigger instanceof HTMLElement &&
        trigger.isConnected &&
        (panel?.contains(document.activeElement) ||
          document.activeElement === document.body)
      ) {
        trigger.focus({ preventScroll: true });
      }
    };
  }, [initialObjective.id]);

  return (
    <Box
      aria-labelledby={`${panelId}-title`}
      className="border-border-strong bg-surface-elevated absolute top-14 right-3 bottom-4 isolate z-50 w-[calc(100%-1.5rem)] overflow-y-auto rounded-xl border-[0.5px] shadow-xl outline-none md:top-[1.625rem] md:right-6 md:bottom-[4.875rem] md:w-[34rem]"
      id={panelId}
      onKeyDown={(event) => {
        if (event.key === "Escape" && !event.defaultPrevented) {
          event.preventDefault();
          onClose();
        }
      }}
      ref={panelRef}
      role="region"
      tabIndex={-1}
    >
      <Flex
        align="center"
        className="border-border bg-surface-elevated sticky top-0 z-10 min-h-16 gap-3 border-b-[0.5px] px-5 py-4"
        justify="between"
      >
        <Link
          className="group focus-visible:ring-primary flex min-w-0 flex-1 items-center gap-2 rounded-sm outline-none focus-visible:ring-1"
          href={objectiveHref}
        >
          <ObjectiveIcon
            className="size-4 shrink-0"
            style={{ color: objective.color }}
          />
          <Flex align="center" className="min-w-0" gap={2}>
            <Text
              as="h2"
              className="line-clamp-2 leading-5"
              fontWeight="semibold"
              id={`${panelId}-title`}
            >
              {objective.name}
            </Text>
            <ChevronRightIcon className="text-text-muted h-4 shrink-0 transition-transform group-hover:translate-x-0.5" />
          </Flex>
        </Link>
        <Button
          aria-label="Close objective details"
          color="tertiary"
          leftIcon={<CloseIcon className="h-4" strokeWidth={3} />}
          onClick={onClose}
          size="sm"
          variant="naked"
        />
      </Flex>

      <Box className="px-5 pt-4 pb-6">
        {objective.shortSummary ? (
          <Text className="mb-5 line-clamp-3 leading-6" color="muted">
            {objective.shortSummary}
          </Text>
        ) : null}

        <ObjectiveForecastRiskBanner className="mb-5" objective={objective} />

        <ObjectiveDetailsProperties objective={objective} />

        <Divider className="my-5" />

        <Text className="mb-3">Progress</Text>
        <ProgressChart
          progressData={analytics?.progressChart ?? EMPTY_PROGRESS_DATA}
        />
        <Divider className="my-5" />

        <ObjectiveDetailsKeyResults
          keyResults={keyResults}
          onSelect={onKeyResultSelect}
        />
      </Box>
    </Box>
  );
};
