"use client";
import dynamic from "next/dynamic";
import { useSprintAnalytics } from "../hooks/sprint-analytics";
import type { SprintHealthItem } from "../types";
import { useAppliedFilters } from "../hooks/filters";
import { SprintHealthSkeleton } from "./sprint-health-skeleton";

const SPRINT_COLORS = {
  Active: "#22c55e",
  Planning: "#6366F1",
  Complete: "#10b981",
  "On Hold": "#eab308",
  Cancelled: "#EA6060",
  Draft: "#6b7280",
  Review: "#8b5cf6",
  Archived: "#374151",
};

const EMPTY_SPRINT_HEALTH: SprintHealthItem[] = [];

const HealthDistributionChart = dynamic(
  () =>
    import("./health-distribution-chart").then(
      (module) => module.HealthDistributionChart,
    ),
  {
    loading: () => <SprintHealthSkeleton />,
    ssr: false,
  },
);

export const SprintHealth = () => {
  const filters = useAppliedFilters();
  const { data: sprintAnalytics, isPending } = useSprintAnalytics(filters);
  const chartData = sprintAnalytics?.sprintHealth ?? EMPTY_SPRINT_HEALTH;

  if (isPending) {
    return <SprintHealthSkeleton />;
  }

  return (
    <HealthDistributionChart
      data={chartData}
      description="Status distribution of active sprints."
      heading="Sprint health"
      itemLabel="sprints"
      statusColors={SPRINT_COLORS}
      totalLabel="Total sprints"
    />
  );
};
