import { Box, Button, Flex, Menu, Text } from "ui";
import { cn } from "lib";
import { CheckIcon } from "icons";
import type { IssueStatus } from "@/types/issue";
import { IssueStatusIcon } from "@/components/ui";

export const ProjectStatusesMenu = ({
  status,
  isSearchEnabled = true,
  asIcon = true,
}: {
  status: IssueStatus;
  isSearchEnabled?: boolean;
  asIcon?: boolean;
}) => {
  const statuses: IssueStatus[] = [
    "Backlog",
    "In Progress",
    "Done",
    "Paused",
    "Canceled",
  ];
  return (
    <Menu>
      <Menu.Button>
        <Button
          className={cn("gap-2 px-3", {
            "h-max p-0 hover:bg-transparent focus:bg-transparent dark:hover:bg-transparent dark:focus:bg-transparent":
              asIcon,
          })}
          color="tertiary"
          leftIcon={<IssueStatusIcon status={status} />}
          size={asIcon ? "sm" : "sm"}
          variant="naked"
        >
          <span className="sr-only">Change priority</span>
          {asIcon ? null : status}
        </Button>
      </Menu.Button>
      <Menu.Items align="center" className="w-64">
        {isSearchEnabled ? (
          <>
            <Menu.Group className="px-4">
              <Menu.Input autoFocus placeholder="Change project status..." />
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
                <IssueStatusIcon status={st} />
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
    </Menu>
  );
};
