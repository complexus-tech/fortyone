import { Box, Flex, Menu, Text } from "ui";
import type { ReactNode } from "react";
import { CheckIcon } from "icons";
import type { StoryPriority } from "@/modules/stories/types";
import { PriorityIcon } from "../priority-icon";

export const PrioritiesMenu = ({ children }: { children: ReactNode }) => {
  return <Menu>{children}</Menu>;
};

const Items = ({
  priority = "No Priority",
  isSearchEnabled = true,
  setPriority,
}: {
  priority?: StoryPriority;
  setPriority?: (priority: StoryPriority) => void;
  isSearchEnabled?: boolean;
}) => {
  const priorities: StoryPriority[] = [
    "No Priority",
    "Urgent",
    "High",
    "Medium",
    "Low",
  ];
  return (
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
            onClick={() => setPriority?.(pr)}
            className="justify-between"
            key={pr}
          >
            <Box className="grid grid-cols-[24px_auto] items-center">
              <PriorityIcon priority={pr} />
              <Text>{pr}</Text>
            </Box>
            <Flex align="center" gap={2}>
              {pr === priority && (
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

const Trigger = ({ children }: { children: ReactNode }) => (
  <Menu.Button asChild>{children}</Menu.Button>
);

PrioritiesMenu.Trigger = Trigger;
PrioritiesMenu.Items = Items;
