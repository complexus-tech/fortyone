import type { ReactNode } from "react";
import { Box, Flex, Menu, Text } from "ui";
import { CheckIcon } from "icons";
import type { StoryStatus } from "@/types/story";
import { StoryStatusIcon } from "../story-status-icon";

export const StatusesMenu = ({ children }: { children: ReactNode }) => {
  return <Menu>{children}</Menu>;
};

const Trigger = ({ children }: { children: ReactNode }) => (
  <Menu.Button>{children}</Menu.Button>
);

const Items = ({
  status = "Backlog",
  isSearchEnabled = true,
}: {
  status?: StoryStatus;
  isSearchEnabled?: boolean;
}) => {
  const statuses: StoryStatus[] = [
    "Backlog",
    "Todo",
    "In Progress",
    "Testing",
    "Done",
    "Duplicate",
    "Canceled",
  ];
  return (
    <Menu.Items align="center" className="w-64">
      {isSearchEnabled ? (
        <>
          <Menu.Group className="px-4">
            <Menu.Input autoFocus placeholder="Change status..." />
          </Menu.Group>
          <Menu.Separator className="my-2" />
        </>
      ) : null}
      <Menu.Group>
        {statuses.map((st, idx) => (
          <Menu.Item
            active={st === status}
            className="justify-between"
            key={st}
          >
            <Box className="grid grid-cols-[24px_auto] items-center">
              <StoryStatusIcon status={st} />
              <Text>{st}</Text>
            </Box>
            <Flex align="center" gap={2}>
              {st === status && (
                <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
              )}
              <Text color="muted">{idx}</Text>
            </Flex>
          </Menu.Item>
        ))}
      </Menu.Group>
    </Menu.Items>
  );
};

StatusesMenu.Trigger = Trigger;
StatusesMenu.Items = Items;
