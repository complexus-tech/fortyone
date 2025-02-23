import { Box, Flex, Text, Button, Menu, TimeAgo } from "ui";
import {
  DeleteIcon,
  LogoutIcon,
  MoreHorizontalIcon,
  SettingsIcon,
} from "icons";
import type { Team } from "@/modules/teams/types";
import { RowWrapper, TeamColor } from "@/components/ui";

export const WorkspaceTeam = ({ name, color, code, createdAt }: Team) => {
  return (
    <RowWrapper className="px-6 last-of-type:border-b-0">
      <Flex align="center" gap={3}>
        <Box className="flex size-8 items-center justify-center rounded-lg bg-gray-100/80 dark:bg-dark-100/80">
          <TeamColor color={color} />
        </Box>
        <Box>
          <Text className="font-medium">{name}</Text>
          <Text className="text-[0.95rem]" color="muted">
            {code}
          </Text>
        </Box>
      </Flex>
      <Flex align="center" gap={2}>
        <Text as="span" className="text-[0.9rem]" color="muted">
          <TimeAgo timestamp={createdAt} />
        </Text>
        <Menu>
          <Menu.Button>
            <Button
              aria-label="More options"
              asIcon
              color="tertiary"
              rounded="full"
              variant="naked"
            >
              <MoreHorizontalIcon className="h-4" />
            </Button>
          </Menu.Button>
          <Menu.Items align="end">
            <Menu.Group>
              <Menu.Item>
                <SettingsIcon />
                Settings
              </Menu.Item>
              <Menu.Item>
                <LogoutIcon />
                Leave team
              </Menu.Item>
            </Menu.Group>
            <Menu.Separator />
            <Menu.Group>
              <Menu.Item>
                <DeleteIcon />
                Delete team
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
    </RowWrapper>
  );
};
