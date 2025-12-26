"use client";
import { Avatar, Box, Flex, Text, Wrapper } from "ui";
import { useState, useEffect } from "react";
import { useTeamPerformance } from "../hooks/team-performance";
import type { MemberContributionItem } from "../types";

export const MemberContributions = () => {
  const { data: teamPerformance, isPending } = useTeamPerformance();
  const [contributionsData, setContributionsData] = useState<
    MemberContributionItem[]
  >([]);

  useEffect(() => {
    if (teamPerformance?.memberContributions.length) {
      const sortedData = teamPerformance.memberContributions
        .slice(0, 6)
        .sort((a, b) => b.completed - a.completed);
      setContributionsData(sortedData);
    }
  }, [teamPerformance]);

  if (isPending) {
    return (
      <Wrapper>
        <Box className="mb-4">
          <Text className="mb-1" fontSize="lg">
            Team contributions
          </Text>
          <Text color="muted" fontSize="sm">
            Performance overview.
          </Text>
        </Box>
        <Box className="space-y-2">
          {Array.from({ length: 4 }).map((_, index) => (
            <Box
              className="h-10 animate-pulse rounded bg-gray-200 dark:bg-dark-100"
              key={index}
            />
          ))}
        </Box>
      </Wrapper>
    );
  }

  return (
    <Wrapper>
      <Box className="mb-4">
        <Text className="mb-1" fontSize="lg">
          Team contributions
        </Text>
        <Text color="muted" fontSize="sm">
          Performance overview.
        </Text>
      </Box>

      {contributionsData.length > 0 ? (
        <Box className="overflow-hidden rounded-[0.6rem] border border-gray-200 dark:border-dark-50">
          {/* Compact Rows */}
          <Box className="bg-white dark:bg-dark-200">
            {contributionsData.map((member, index) => {
              const completionRate =
                member.assigned > 0
                  ? (member.completed / member.assigned) * 100
                  : 0;

              return (
                <Box
                  className={`px-3 py-2.5 ${
                    index !== contributionsData.length - 1
                      ? "border-b border-gray-100 dark:border-dark-50"
                      : ""
                  }`}
                  key={member.userId}
                >
                  <Flex align="center" className="mb-2" gap={2}>
                    <Avatar
                      name={member.username}
                      size="xs"
                      src={member.avatarUrl}
                    />
                    <Text className="truncate text-sm font-medium">
                      {member.username}
                    </Text>
                  </Flex>

                  <Flex className="mb-1" justify="between">
                    <Text color="muted" fontSize="xs">
                      {member.completed}/{member.assigned}
                    </Text>
                    <Text className="font-medium" fontSize="xs">
                      {Math.round(completionRate)}%
                    </Text>
                  </Flex>

                  <Box className="h-1 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-100">
                    <Box
                      className="from-green-500 to-green-600 h-full rounded-full bg-linear-to-r"
                      style={{
                        width: `${Math.min(completionRate, 100)}%`,
                      }}
                    />
                  </Box>
                </Box>
              );
            })}
          </Box>
        </Box>
      ) : (
        <Box className="text-gray-400 flex h-[100px] items-center justify-center">
          <Text fontSize="sm">No data available</Text>
        </Box>
      )}
    </Wrapper>
  );
};
