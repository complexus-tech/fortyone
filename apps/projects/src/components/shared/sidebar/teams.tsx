"use client";
import { Box, Button, Flex, Text } from "ui";
import { MoreHorizontalIcon, PlusIcon, TeamIcon } from "icons";
import { Team } from "./team";
import { useTeams } from "@/lib/hooks/teams";

export const Teams = () => {
  const { data: teams = [] } = useTeams();

  return (
    <Box className="mt-4">
      <Text className="mb-2 pl-2.5 font-medium" color="muted">
        Your Teams
      </Text>
      <Flex direction="column" gap={1}>
        {teams.map(({ id, icon, name, color }) => (
          <Team
            icon={
              <Box className="flex w-6 items-center justify-center rounded-xl text-lg">
                {icon || "🌟"}
              </Box>
            }
            id={id}
            key={id}
            name={name}
          />
        ))}
      </Flex>
    </Box>
  );
};
