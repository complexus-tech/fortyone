"use client";

import { WarningIcon } from "icons";
import Link from "next/link";
import { Box, Flex, Text, Tooltip } from "ui";
import { useTerminology, useWorkspacePath } from "@/hooks";
import { getStoryPath } from "@/modules/story/utils/story-url";
import type { Objective } from "../types";
import { getObjectiveForecastRiskCopy } from "./objective-forecast-risk-utils";

export const ObjectiveForecastRiskBadge = ({
  objective,
}: {
  objective: Pick<
    Objective,
    | "endDate"
    | "forecastCauseStory"
    | "forecastDaysDelta"
    | "forecastEndDate"
    | "scheduleStatus"
  >;
}) => {
  const { getTermDisplay } = useTerminology();
  const copy = getObjectiveForecastRiskCopy(
    objective,
    getTermDisplay("storyTerm", { capitalize: true }),
  );
  if (!copy) return null;

  return (
    <Tooltip className="max-w-80" title={copy.description}>
      <span className="border-danger/20 bg-danger/5 text-danger inline-flex h-6 shrink-0 items-center gap-1 rounded-md border px-1.5 text-[0.75rem] leading-none font-medium tabular-nums">
        <WarningIcon aria-hidden="true" className="h-3.5 w-auto" />
        {copy.shortLabel}
      </span>
    </Tooltip>
  );
};

export const ObjectiveForecastRiskBanner = ({
  className,
  objective,
}: {
  className?: string;
  objective: Objective;
}) => {
  const { getTermDisplay } = useTerminology();
  const { withWorkspace } = useWorkspacePath();
  const storyTerm = getTermDisplay("storyTerm", { capitalize: true });
  const copy = getObjectiveForecastRiskCopy(objective, storyTerm);
  if (!copy) return null;

  return (
    <Flex
      align="start"
      aria-label="Objective forecast risk"
      className={`border-danger/20 bg-danger/5 rounded-xl border px-4 py-3 ${className ?? ""}`}
      gap={2}
      justify="between"
      role="status"
    >
      <Flex align="start" className="min-w-0" gap={2}>
        <WarningIcon
          aria-hidden="true"
          className="text-danger dark:text-danger mt-0.5 h-5 w-auto shrink-0"
        />
        <Box className="min-w-0">
          <Text color="danger" fontWeight="medium">
            {copy.headline}
          </Text>
          <Text className="mt-0.5 leading-snug" color="muted" fontSize="sm">
            {copy.description}
          </Text>
        </Box>
      </Flex>
      {objective.forecastCauseStory ? (
        <Link
          className="text-danger hover:text-danger/80 focus-visible:ring-ring mt-0.5 shrink-0 rounded-sm text-sm font-medium outline-none focus-visible:ring-2"
          href={withWorkspace(
            getStoryPath({ id: objective.forecastCauseStory.id }),
          )}
        >
          Open {storyTerm.toLowerCase()}
        </Link>
      ) : null}
    </Flex>
  );
};
