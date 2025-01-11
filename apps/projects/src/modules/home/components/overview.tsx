"use client";
import { Box, Flex, Text, Wrapper } from "ui";
import { useSummary } from "@/lib/hooks/summary";

const Card = ({ title, count }: { title: string; count?: number }) => (
  <Wrapper className="px-5">
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
  const timeOfDay = () => {
    const hour = new Date().getHours();
    if (hour < 12) return "morning";
    if (hour < 18) return "afternoon";
    return "evening";
  };

  const { data: summary } = useSummary();
  const overview = [
    {
      count: summary?.closed,
      title: "Stories closed",
    },
    {
      count: summary?.overdue,
      title: "Stories overdue",
    },
    {
      count: summary?.inProgress,
      title: "Stories in progress",
    },
    {
      count: summary?.created,
      title: "Created by you",
    },
    {
      count: summary?.assigned,
      title: "Assigned to you",
    },
  ];
  return (
    <Box>
      <Text as="h2" fontSize="3xl" fontWeight="medium">
        Good {timeOfDay()}, Joseph.
      </Text>
      <Text color="muted">
        Here&rsquo;s what&rsquo;s happening with your stories.
      </Text>
      <Box className="my-4 grid grid-cols-5 gap-4">
        {overview.map((item) => (
          <Card key={item.title} {...item} />
        ))}
      </Box>
    </Box>
  );
};
