import { cn } from "lib";
import { TbCheck } from "react-icons/tb";
import { Box, Button, Flex, Menu, Text } from "ui";
import { PriorityIcon } from "../priority-icon";
import type { IssuePriority } from "@/types/issue";

export const PrioritiesMenu = ({
  priority,
  isSearchEnabled,
}: {
  priority: IssuePriority;
  isSearchEnabled?: boolean;
}) => {
  const priorities: IssuePriority[] = [
    "No Priority",
    "Urgent",
    "High",
    "Medium",
    "Low",
  ];
  return (
    <Menu>
      <Menu.Button>
        <Button
          className="h-max p-0 hover:bg-transparent focus:bg-transparent dark:hover:bg-transparent dark:focus:bg-transparent"
          color="tertiary"
          leftIcon={<PriorityIcon priority={priority} />}
          size="sm"
          variant="naked"
        >
          <span className="sr-only">Change priority</span>
        </Button>
      </Menu.Button>
      <Menu.Items align="center" className="w-64">
        {isSearchEnabled ? (
          <>
            <Menu.Group className="mb-2 px-4">
              <Menu.Input autoFocus placeholder="Change priority" />
            </Menu.Group>
            <Menu.Separator />
          </>
        ) : null}
        <Menu.Group>
          {priorities.map((pr, idx) => (
            <Menu.Item
              active={pr === priority}
              className="justify-between"
              key={pr}
            >
              <Box className="grid grid-cols-[24px_auto] items-center">
                <PriorityIcon
                  className={cn({
                    "relative left-[1px] text-gray": pr === "Urgent",
                  })}
                  priority={pr}
                />
                <Text>{pr}</Text>
              </Box>
              <Flex align="center" gap={2}>
                {pr === priority && (
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
