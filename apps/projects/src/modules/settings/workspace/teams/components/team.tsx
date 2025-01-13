import { Box, Flex, Text, Button, Avatar, Badge, Menu } from "ui";
import { DeleteIcon, EditIcon, LogoutIcon, MoreHorizontalIcon } from "icons";
import type { Team } from "@/modules/teams/types";
import { RowWrapper, TeamColor } from "@/components/ui";

export const WorkspaceTeam = ({ name, color }: Team) => {
  return (
    <RowWrapper className="px-6">
      <Flex align="center" gap={3}>
        <Box className="flex size-8 items-center justify-center rounded-lg dark:bg-dark-100/80">
          <TeamColor color={color} />
        </Box>
        <Box>
          <Text className="font-medium">{name}</Text>
          <Text color="muted">8 members</Text>
        </Box>
      </Flex>
      <Flex align="center" gap={2}>
        <Flex className="-space-x-2">
          <Avatar
            className="ring-1 ring-white dark:ring-dark-100"
            name="John Doe"
            size="sm"
          />
          <Avatar
            className="ring-1 ring-white dark:ring-dark-100"
            name="Jane Smith"
            size="sm"
          />
          <Badge className="ml-2" color="tertiary">
            +6
          </Badge>
        </Flex>
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
                <EditIcon />
                Edit team
              </Menu.Item>
              <Menu.Item>
                <LogoutIcon />
                Leave team
              </Menu.Item>
            </Menu.Group>
            <Menu.Separator />
            <Menu.Group>
              <Menu.Item className="text-primary">
                <DeleteIcon className="text-primary dark:text-primary" />
                Delete team
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
    </RowWrapper>
  );
};
