"use client";

import { createContext, useContext, useState, type ReactNode } from "react";
import { CheckIcon, StopIcon, TimeScheduleIcon } from "icons";
import { Box, Command, Flex, Popover, Text } from "ui";

const AutoSchedulingContext = createContext<{
  setOpen: (open: boolean) => void;
}>({ setOpen: () => {} });

export const AutoSchedulingMenu = ({ children }: { children: ReactNode }) => {
  const [open, setOpen] = useState(false);

  return (
    <AutoSchedulingContext.Provider value={{ setOpen }}>
      <Popover onOpenChange={setOpen} open={open}>
        {children}
      </Popover>
    </AutoSchedulingContext.Provider>
  );
};

const Trigger = ({ children }: { children: ReactNode }) => (
  <Popover.Trigger asChild>{children}</Popover.Trigger>
);

const Items = ({
  autoSchedulingEnabled,
  autoSchedulingLocked = false,
  setAutoSchedulingEnabled,
}: {
  autoSchedulingEnabled: boolean;
  autoSchedulingLocked?: boolean;
  setAutoSchedulingEnabled: (enabled: boolean) => void;
}) => {
  const { setOpen } = useContext(AutoSchedulingContext);
  const options = [
    {
      enabled: true,
      icon: <TimeScheduleIcon className="text-success h-4 w-auto" />,
      label: "On",
    },
    {
      enabled: false,
      icon: <StopIcon className="text-text-muted h-4 w-auto" />,
      label: "Off",
    },
  ];

  return (
    <Popover.Content align="center" className="w-40">
      <Command>
        <Command.Group>
          {options.map((option) => {
            const isSelected = option.enabled === autoSchedulingEnabled;
            const isDisabled = !option.enabled && autoSchedulingLocked;

            return (
              <Command.Item
                active={isSelected}
                className="justify-between"
                disabled={isDisabled}
                key={option.label}
                onSelect={() => {
                  if (isDisabled) return;

                  if (!isSelected) {
                    setAutoSchedulingEnabled(option.enabled);
                  }
                  setOpen(false);
                }}
                value={option.label}
              >
                <Box className="grid grid-cols-[24px_auto] items-center">
                  {option.icon}
                  <Text fontWeight="medium">{option.label}</Text>
                </Box>
                <Flex align="center" gap={2}>
                  {isSelected ? (
                    <CheckIcon
                      className="h-5 w-auto shrink-0"
                      strokeWidth={2.1}
                    />
                  ) : null}
                </Flex>
              </Command.Item>
            );
          })}
        </Command.Group>
      </Command>
    </Popover.Content>
  );
};

AutoSchedulingMenu.Trigger = Trigger;
AutoSchedulingMenu.Items = Items;
