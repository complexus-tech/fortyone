"use client";
import { Box, Flex, Text, Button } from "ui";
import { MoreHorizontalIcon, TeamIcon } from "icons";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useUserRole } from "@/hooks";
import { TeamsMenu } from "@/components/ui/teams-menu";
import { Team } from "./team";

export const Teams = () => {
  const { data: teams = [] } = useTeams();
  const { userRole } = useUserRole();

  return (
    <Box className="mt-5">
      <Flex align="center" className="mb-2.5" justify="between">
        <Text
          className="flex items-center gap-1 pl-2.5 font-medium"
          color="muted"
        >
          <TeamIcon className="h-[1.2rem]" />
          Your Teams
        </Text>
        {userRole !== "guest" && (
          <TeamsMenu>
            <TeamsMenu.Trigger>
              <Button
                asIcon
                color="tertiary"
                leftIcon={<MoreHorizontalIcon />}
                rounded="full"
                size="sm"
                variant="naked"
              >
                <span className="sr-only">Manage Teams</span>
              </Button>
            </TeamsMenu.Trigger>
            <TeamsMenu.Items />
          </TeamsMenu>
        )}
      </Flex>
      <Flex direction="column" gap={1}>
        {teams.map(({ id, name, color }) => (
          <Team color={color} id={id} key={id} name={name} />
        ))}
      </Flex>
    </Box>
  );
};
