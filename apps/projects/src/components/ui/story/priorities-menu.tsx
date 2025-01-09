"use client";
import { Box, Command, Flex, Popover, Text, Divider } from "ui";
import type { ReactNode} from "react";
import { createContext, useContext, useState } from "react";
import { CheckIcon } from "icons";
import type { StoryPriority } from "@/modules/stories/types";
import { PriorityIcon } from "../priority-icon";

const PriorityContext = createContext<{
  open: boolean;
  setOpen: (open: boolean) => void;
}>({
  open: false,
  setOpen: () => {},
});

export const usePriorityMenu = () => {
  const { open, setOpen } = useContext(PriorityContext);
  return { open, setOpen };
};

const Menu = ({ children }: { children: ReactNode }) => {
  const { open, setOpen } = usePriorityMenu();
  return (
    <Popover onOpenChange={setOpen} open={open}>
      {children}
    </Popover>
  );
};

export const PrioritiesMenu = ({ children }: { children: ReactNode }) => {
  const [open, setOpen] = useState(false);
  return (
    <PriorityContext.Provider value={{ open, setOpen }}>
      <Menu>{children}</Menu>
    </PriorityContext.Provider>
  );
};

const Items = ({
  priority = "No Priority",
  setPriority,
}: {
  priority?: StoryPriority;
  setPriority: (priority: StoryPriority) => void;
}) => {
  const priorities: StoryPriority[] = [
    "No Priority",
    "Low",
    "Medium",
    "High",
    "Urgent",
  ];
  const { setOpen } = usePriorityMenu();
  return (
    <Popover.Content align="center" className="w-64">
      <Command>
        <Command.Input autoFocus placeholder="Change priority..." />
        <Divider className="my-2" />
        <Command.Empty className="py-2">
          <Text color="muted">No priority found.</Text>
        </Command.Empty>
        <Command.Group>
          {priorities.map((pr, idx) => (
            <Command.Item
              active={pr === priority}
              className="justify-between"
              key={pr}
              onSelect={() => {
                if (pr !== priority) {
                  setPriority(pr);
                }
                setOpen(false);
              }}
              value={pr}
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
            </Command.Item>
          ))}
        </Command.Group>
      </Command>
    </Popover.Content>
  );
};

const Trigger = ({ children }: { children: ReactNode }) => (
  <Popover.Trigger asChild>{children}</Popover.Trigger>
);

PrioritiesMenu.Trigger = Trigger;
PrioritiesMenu.Items = Items;
