"use client";
import { Box, Button, Flex, Text } from "ui";
import { MoreHorizontalIcon, PlusIcon, TeamIcon } from "icons";
import { Team } from "./team";
import { useTeams } from "@/lib/hooks/teams";

export const Teams = () => {
  const { data: teams = [] } = useTeams();

  return (
    <Box className="mt-4">
      <Flex
        align="center"
        className="group mb-1 h-[2.5rem] select-none rounded-lg pl-3 pr-1"
        justify="between"
        role="button"
        tabIndex={0}
      >
        <Text className="flex items-center gap-2 font-medium">
          <TeamIcon className="h-[1.3rem] w-auto" />
          Teams
        </Text>
        <Button
          color="tertiary"
          className="aspect-square"
          rounded="full"
          leftIcon={<MoreHorizontalIcon className="h-5 w-auto" />}
          size="sm"
          type="button"
          variant="naked"
        >
          <span className="sr-only">Add new team</span>
        </Button>
      </Flex>
      <Flex direction="column" gap={1}>
        {teams.map(({ id, icon, name, color }) => (
          <Team
            icon={
              <Box className="flex w-6 items-center justify-center rounded-xl text-lg">
                {icon || <TeamIcon className="h-5 w-auto" style={{ color }} />}
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
