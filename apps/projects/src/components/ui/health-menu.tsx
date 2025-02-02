"use client";

import { Box, Command, Flex, Popover, Text, Divider } from "ui";
import type { ReactNode } from "react";
import { createContext, useContext, useState } from "react";
import { CheckIcon } from "icons";
import type { ObjectiveHealth } from "@/modules/objectives/types";
import { ObjectiveHealthIcon } from "./objective-health-icon";

const HealthContext = createContext<{
  open: boolean;
  setOpen: (open: boolean) => void;
}>({
  open: false,
  setOpen: () => {},
});

export const useHealthMenu = () => {
  const { open, setOpen } = useContext(HealthContext);
  return { open, setOpen };
};

const Menu = ({ children }: { children: ReactNode }) => {
  const { open, setOpen } = useHealthMenu();
  return (
    <Popover onOpenChange={setOpen} open={open}>
      {children}
    </Popover>
  );
};

export const HealthMenu = ({ children }: { children: ReactNode }) => {
  const [open, setOpen] = useState(false);
  return (
    <HealthContext.Provider value={{ open, setOpen }}>
      <Menu>{children}</Menu>
    </HealthContext.Provider>
  );
};

const Items = ({
  health,
  setHealth,
}: {
  health?: ObjectiveHealth;
  setHealth: (health: ObjectiveHealth) => void;
}) => {
  const healthOptions: ObjectiveHealth[] = ["On Track", "At Risk", "Off Track"];
  const { setOpen } = useHealthMenu();

  return (
    <Popover.Content align="center" className="w-64">
      <Command>
        <Command.Input autoFocus placeholder="Change health status..." />
        <Divider className="my-2" />
        <Command.Empty className="py-2">
          <Text color="muted">No health status found.</Text>
        </Command.Empty>
        <Command.Group>
          {healthOptions.map((status) => (
            <Command.Item
              active={status === health}
              className="justify-between"
              key={status ?? "no-status"}
              onSelect={() => {
                if (status !== health) {
                  setHealth(status);
                }
                setOpen(false);
              }}
              value={status ?? "No Status"}
            >
              <Box className="grid grid-cols-[24px_auto] items-center">
                <ObjectiveHealthIcon health={status} />
                <Text>{status ?? "No Status"}</Text>
              </Box>
              <Flex align="center" gap={2}>
                {status === health && (
                  <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                )}
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

HealthMenu.Trigger = Trigger;
HealthMenu.Items = Items;
