import { TbCheck } from "react-icons/tb";
import { Box, Button, Flex, Menu, Text } from "ui";
import { IssueStatusIcon } from "../issue-status-icon";
import type { IssueStatus } from "@/types/issue";

export const StatusesMenu = ({
  status,
  isSearchEnabled,
}: {
  status: IssueStatus;
  isSearchEnabled?: boolean;
}) => {
  const statuses: IssueStatus[] = [
    "Backlog",
    "Todo",
    "In Progress",
    "Testing",
    "Done",
    "Duplicate",
    "Canceled",
  ];
  return (
    <Menu>
      <Menu.Button>
        <Button
          className="h-max p-0 hover:bg-transparent focus:bg-transparent dark:hover:bg-transparent dark:focus:bg-transparent"
          color="tertiary"
          leftIcon={<IssueStatusIcon status={status} />}
          size="sm"
          variant="naked"
        >
          <span className="sr-only">Change status</span>
        </Button>
      </Menu.Button>
      <Menu.Items align="center" className="w-64">
        {isSearchEnabled ? (
          <>
            <Menu.Group className="mb-2 px-4">
              <Menu.Input autoFocus placeholder="Change status" />
            </Menu.Group>
            <Menu.Separator />
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
                <IssueStatusIcon status={st} />
                <Text>{st}</Text>
              </Box>
              <Flex align="center" gap={2}>
                {st === status && (
                  <TbCheck className="h-5 w-auto" strokeWidth={2.1} />
                )}
                <Text color="muted">{idx}</Text>
              </Flex>
            </Menu.Item>
          ))}
        </Menu.Group>
      </Menu.Items>
    </Menu>
  );
};
