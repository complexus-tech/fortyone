import { cn } from "lib";
import { Box, Button, Flex, Menu, Text } from "ui";
import { Check } from "lucide-react";
import type { IssuePriority } from "@/types/issue";
import { PriorityIcon } from "../priority-icon";

export const PrioritiesMenu = ({
  priority,
  isSearchEnabled = true,
  asIcon = true,
}: {
  priority: IssuePriority;
  isSearchEnabled?: boolean;
  asIcon?: boolean;
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
          className={cn("gap-3 px-3", {
            "h-max p-0 hover:bg-transparent focus:bg-transparent dark:hover:bg-transparent dark:focus:bg-transparent":
              asIcon,
          })}
          color="tertiary"
          leftIcon={<PriorityIcon priority={priority} />}
          size={asIcon ? "sm" : "md"}
          variant="naked"
        >
          <span className="sr-only">Change priority</span>
          {asIcon ? null : priority}
        </Button>
      </Menu.Button>
      <Menu.Items align="center" className="w-64">
        {isSearchEnabled ? (
          <>
            <Menu.Group className="px-4">
              <Menu.Input autoFocus placeholder="Change priority..." />
            </Menu.Group>
            <Menu.Separator className="my-2" />
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
                    "relative left-[1px] text-danger": pr === "Urgent",
                  })}
                  priority={pr}
                />
                <Text>{pr}</Text>
              </Box>
              <Flex align="center" gap={2}>
                {pr === priority && (
                  <Check className="h-5 w-auto" strokeWidth={2.1} />
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
