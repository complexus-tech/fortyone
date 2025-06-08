"use client";
import { Avatar, Box, Flex, Text, Wrapper } from "ui";
import { useState, useEffect } from "react";
import { useTeamPerformance } from "../hooks/team-performance";
import type { MemberContributionItem } from "../types";
import { MemberContributionsSkeleton } from "./member-contributions-skeleton";

export const MemberContributions = () => {
  const { data: teamPerformance, isPending } = useTeamPerformance();
  const [contributionsData, setContributionsData] = useState<
    MemberContributionItem[]
  >([]);

  useEffect(() => {
    if (teamPerformance?.memberContributions.length) {
      setContributionsData(teamPerformance.memberContributions.slice(0, 8));
    }
  }, [teamPerformance]);

  if (isPending) {
    return <MemberContributionsSkeleton />;
  }

  const maxCompleted = Math.max(
    ...contributionsData.map((item) => item.completed),
  );

  return (
    <Wrapper>
      <Box className="mb-6">
        <Text className="mb-1" fontSize="lg">
          Member contributions
        </Text>
        <Text color="muted">Top contributors by completed work.</Text>
      </Box>

      <Box className="space-y-3">
        {contributionsData.map((member) => (
          <Flex align="center" gap={3} key={member.userId}>
            <Avatar name={member.username} size="sm" src={member.avatarUrl} />
            <Box className="flex-1">
              <Flex align="center" className="mb-1" justify="between">
                <Text className="truncate font-medium">{member.username}</Text>
                <Text color="muted" fontSize="sm">
                  {member.completed}
                </Text>
              </Flex>
              <Box className="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-100">
                <Box
                  className="from-blue-500 to-purple-600 h-full rounded-full bg-gradient-to-r transition-all duration-300"
                  style={{
                    width: `${(member.completed / maxCompleted) * 100}%`,
                  }}
                />
              </Box>
            </Box>
          </Flex>
        ))}
      </Box>

      {contributionsData.length === 0 && (
        <Box className="text-gray-400 flex h-[180px] items-center justify-center">
          <Text>No contribution data available</Text>
        </Box>
      )}
    </Wrapper>
  );
};
