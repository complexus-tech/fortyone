"use client";
import { Box, Flex, Text } from "ui";
import { Team } from "./team";
import { useTeams } from "@/modules/teams/hooks/teams";
import { TeamIcon } from "icons";

export const Teams = () => {
  const { data: teams = [] } = useTeams();

  return (
    <Box className="mt-4">
      <Text
        className="mb-2.5 flex items-center gap-1 pl-2.5 font-medium"
        color="muted"
      >
        <TeamIcon className="h-[1.2rem] w-auto text-gray dark:text-gray-300" />
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
            color={color}
          />
        ))}
      </Flex>
    </Box>
  );
};
