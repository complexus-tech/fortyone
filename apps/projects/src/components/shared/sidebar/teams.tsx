"use client";
import { Box, Flex, Text, Button, Menu } from "ui";
import { MoreHorizontalIcon, TeamIcon } from "icons";
import { useRouter } from "next/navigation";
import nProgress from "nprogress";
import { useTeams } from "@/modules/teams/hooks/teams";
import { Team } from "./team";

export const Teams = () => {
  const router = useRouter();
  const { data: teams = [] } = useTeams();

  return (
    <Box className="mt-5">
      <Flex align="center" className="group mb-2.5" justify="between">
        <Text
          className="flex items-center gap-1 pl-2.5 font-medium opacity-70"
          color="muted"
        >
          <TeamIcon className="h-[1.2rem]" />
          Your Teams
        </Text>

        <Menu>
          <Menu.Button>
            <Button
              asIcon
              className="pointer-events-none opacity-0 transition duration-200 ease-linear group-hover:pointer-events-auto group-hover:opacity-100"
              color="tertiary"
              leftIcon={<MoreHorizontalIcon />}
              rounded="full"
              size="sm"
              variant="naked"
            >
              <span className="sr-only">Add Team</span>
            </Button>
          </Menu.Button>
          <Menu.Items>
            <Menu.Group>
              <Menu.Item
                onSelect={() => {
                  nProgress.start();
                  router.push("/settings/workspace/teams");
                }}
              >
                <TeamIcon /> Manage Teams
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
      <Flex direction="column" gap={1}>
        {teams.map(({ id, name, color }) => (
          <Team color={color} id={id} key={id} name={name} />
        ))}
      </Flex>
    </Box>
  );
};
