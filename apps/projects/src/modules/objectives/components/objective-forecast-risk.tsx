"use client";

import { LinkIcon, WarningIcon } from "icons";
import { cn } from "lib";
import Link from "next/link";
import { Tooltip } from "ui";
import { SignalBannerRow } from "@/components/ui/signal-banner-row";
import { useTerminology, useWorkspacePath } from "@/hooks";
import { getStoryPath } from "@/modules/story/utils/story-url";
import type { Objective } from "../types";
import { getObjectiveForecastRiskCopy } from "./objective-forecast-risk-utils";

export const ObjectiveForecastRiskBadge = ({
  className,
  objective,
}: {
  className?: string;
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
      <span
        className={cn(
          "border-primary/20 bg-primary/5 text-primary inline-flex h-6 shrink-0 items-center gap-1 rounded-md border px-1.5 text-[0.75rem] leading-none font-medium tabular-nums",
          className,
        )}
      >
        <WarningIcon aria-hidden="true" className="text-primary h-3.5 w-auto" />
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

  const action = objective.forecastCauseStory ? (
    <Link
      aria-label={`Open ${storyTerm.toLowerCase()} driving this forecast`}
      className="text-primary hover:text-primary/80 focus-visible:ring-ring grid size-7 place-items-center rounded-md transition outline-none focus-visible:ring-2"
      href={withWorkspace(
        getStoryPath({ id: objective.forecastCauseStory.id }),
      )}
      title={`Open ${storyTerm.toLowerCase()}`}
    >
      <LinkIcon aria-hidden="true" className="text-current" />
    </Link>
  ) : null;

  return (
    <SignalBannerRow
      actions={action}
      ariaLabel="Objective forecast risk"
      className={className}
      icon={
        <WarningIcon aria-hidden="true" className="text-primary h-5 shrink-0" />
      }
      title={`${copy.headline} · ${copy.description}`}
    >
      {copy.headline}
      <span aria-hidden="true"> · </span>
      {copy.description}
    </SignalBannerRow>
  );
};
