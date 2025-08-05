"use client";
import { Box, Flex, Text, Wrapper } from "ui";
import { useWorkspaceOverview } from "../hooks/workspace-overview";
import { useAppliedFilters } from "../hooks/filters";
import { OverviewSkeleton } from "./overview-skeleton";

const Card = ({ title, count }: { title: string; count?: number }) => (
  <Wrapper className="px-3 py-3 md:px-5 md:py-4">
    <Flex justify="between">
      <Text className="mb-1.5 text-2xl antialiased" fontWeight="semibold">
        {count}
      </Text>
      <svg
        className="h-5 w-auto"
        fill="none"
        height="24"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="2"
        viewBox="0 0 24 24"
        width="24"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path d="M7 7h10v10" />
        <path d="M7 17 17 7" />
      </svg>
    </Flex>
    <Text className="opacity-80" color="muted">
      {title}
    </Text>
  </Wrapper>
);

export const Overview = () => {
  const filters = useAppliedFilters();

  const { data: overview, isPending } = useWorkspaceOverview(filters);

  if (isPending) {
    return <OverviewSkeleton />;
  }

  const metrics = [
    {
      count: overview?.metrics.totalStories,
      title: "Total Stories",
    },
    {
      count: overview?.metrics.completedStories,
      title: "Completed Stories",
    },
    {
      count: overview?.metrics.activeObjectives,
      title: "Active Objectives",
    },
    {
      count: overview?.metrics.activeSprints,
      title: "Active Sprints",
    },
    {
      count: overview?.metrics.totalTeamMembers,
      title: "Team Members",
    },
  ];

  return (
    <Box className="mb-4 mt-3 grid grid-cols-2 gap-3 md:grid-cols-5 md:gap-4">
      {metrics.map((item) => (
        <Card key={item.title} {...item} />
      ))}
    </Box>
  );
};
