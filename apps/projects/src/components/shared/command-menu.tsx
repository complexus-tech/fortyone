"use client";

import { useState } from "react";
import { useHotkeys } from "react-hotkeys-hook";
import { Dialog, Command, Text, Divider, Kbd, Flex } from "ui";
import { PlusIcon } from "icons";
import { useTerminology } from "@/hooks";

export const CommandMenu = () => {
  const { getTermDisplay } = useTerminology();
  const [open, setOpen] = useState(false);

  useHotkeys("mod+k", (e) => {
    e.preventDefault();
    setOpen(true);
  });

  const commands = [
    {
      group: "Actions",
      items: [
        {
          label: `Create new ${getTermDisplay("storyTerm")}`,
          icon: <PlusIcon />,
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>⇧</Kbd>
              <Kbd>n</Kbd>
            </Flex>
          ),
          action: () => {},
        },
      ],
    },
  ];

  return (
    <Dialog onOpenChange={setOpen} open={open}>
      <Dialog.Content className="max-w-3xl" hideClose>
        <Dialog.Header className="sr-only">
          <Dialog.Title className="sr-only">Command Menu</Dialog.Title>
        </Dialog.Header>
        <Dialog.Body className="px-4 pt-2">
          <Command>
            <Command.Input
              className="my-2.5 text-2xl font-normal"
              placeholder="What do you want to do?"
            />
            <Divider className="my-2.5" />
            <Command.Empty className="py-2">
              <Text color="muted" fontSize="xl" fontWeight="normal">
                No results found.
              </Text>
            </Command.Empty>

            {commands.map((command) => (
              <Command.Group
                className="px-0"
                heading={
                  <Text className="mb-1" color="muted">
                    {command.group}
                  </Text>
                }
                key={command.group}
              >
                {command.items.map((item) => (
                  <Command.Item
                    className="justify-between rounded-[0.6rem] px-4 py-3 text-lg antialiased"
                    key={item.label}
                    onSelect={item.action}
                  >
                    <Flex align="center" gap={3}>
                      {item.icon}
                      {item.label}
                    </Flex>
                    {item.shortcut}
                  </Command.Item>
                ))}
              </Command.Group>
            ))}
          </Command>
        </Dialog.Body>
      </Dialog.Content>
    </Dialog>
  );
};
