"use client";
import dynamic from "next/dynamic";
import { useObjectiveProgress } from "../hooks/objective-progress";
import type { HealthDistributionItem } from "../types";
import { useAppliedFilters } from "../hooks/filters";
import { ObjectiveHealthSkeleton } from "./objective-health-skeleton";

const HEALTH_COLORS = {
  "On Track": "#22c55e",
  "At Risk": "#eab308",
  "Off Track": "#EA6060",
  "Not Started": "#6b7280",
  Completed: "#10b981",
  Blocked: "#f97316",
  Planning: "#6366F1",
  Review: "#8b5cf6",
};

const EMPTY_HEALTH_DISTRIBUTION: HealthDistributionItem[] = [];

const HealthDistributionChart = dynamic(
  () =>
    import("./health-distribution-chart").then(
      (module) => module.HealthDistributionChart,
    ),
  {
    loading: () => <ObjectiveHealthSkeleton />,
    ssr: false,
  },
);

export const ObjectiveHealth = () => {
  const filters = useAppliedFilters();
  const { data: objectiveProgress, isPending } = useObjectiveProgress(filters);
  const chartData =
    objectiveProgress?.healthDistribution ?? EMPTY_HEALTH_DISTRIBUTION;

  if (isPending) {
    return <ObjectiveHealthSkeleton />;
  }

  return (
    <HealthDistributionChart
      data={chartData}
      description="Health status distribution of objectives."
      heading="Objective health"
      itemLabel="objectives"
      statusColors={HEALTH_COLORS}
      totalLabel="Total objectives"
    />
  );
};
