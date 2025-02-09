"use client";
import { Box, Flex, Text } from "ui";
import { TeamIcon } from "icons";
import { useTeams } from "@/modules/teams/hooks/teams";
import { Team } from "./team";

export const Teams = () => {
  const { data: teams = [] } = useTeams();

  return (
    <Box className="mt-5">
      <Text
        className="mb-2.5 flex items-center gap-1 pl-2.5 font-medium opacity-70"
        color="muted"
      >
        <TeamIcon className="h-[1.2rem]" />
        Your Teams
      </Text>
      <Flex direction="column" gap={1}>
        {teams.map(({ id, name, color }) => (
          <Team color={color} id={id} key={id} name={name} />
        ))}
      </Flex>
    </Box>
  );
};
