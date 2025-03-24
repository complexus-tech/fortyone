"use client";

import { useState } from "react";
import { useHotkeys } from "react-hotkeys-hook";
import { Dialog, Command, Text, Divider, Kbd, Flex, Box } from "ui";
import {
  PlusIcon,
  NotificationsIcon,
  ObjectiveIcon,
  ListIcon,
  FilterIcon,
  SunIcon,
  HelpIcon,
  LogoutIcon,
  SearchIcon,
  SettingsIcon,
  UserIcon,
  DashboardIcon,
  KanbanIcon,
} from "icons";
import { useTerminology } from "@/hooks";

export const CommandMenu = () => {
  const { getTermDisplay } = useTerminology();
  const [open, setOpen] = useState(false);

  useHotkeys("mod+k", (e) => {
    e.preventDefault();
    setOpen((prev) => !prev);
  });

  const commands = [
    {
      group: "Quick Actions",
      items: [
        {
          label: `New ${getTermDisplay("storyTerm")}`,
          icon: <PlusIcon />,
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>⇧</Kbd>
              <Kbd>n</Kbd>
            </Flex>
          ),
          action: () => {},
        },
        {
          label: `New ${getTermDisplay("objectiveTerm")}`,
          icon: <PlusIcon />,
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>⇧</Kbd>
              <Kbd>o</Kbd>
            </Flex>
          ),
          action: () => {},
        },
        {
          label: `New ${getTermDisplay("sprintTerm")}`,
          icon: <PlusIcon />,
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>⇧</Kbd>
              <Kbd>s</Kbd>
            </Flex>
          ),
          action: () => {},
        },
      ],
    },
    {
      group: "Go To",
      items: [
        {
          label: `My ${getTermDisplay("storyTerm", { variant: "plural" })}`,
          icon: <UserIcon />,
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>g</Kbd>
              <Kbd>m</Kbd>
            </Flex>
          ),
          action: () => {},
        },
        {
          label: "Inbox",
          icon: <NotificationsIcon />,
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>g</Kbd>
              <Kbd>i</Kbd>
            </Flex>
          ),
          action: () => {},
        },
        {
          label: "Summary",
          icon: <DashboardIcon />,
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>g</Kbd>
              <Kbd>s</Kbd>
            </Flex>
          ),
          action: () => {},
        },
        {
          label: getTermDisplay("objectiveTerm", {
            variant: "plural",
            capitalize: true,
          }),
          icon: <ObjectiveIcon />,
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>g</Kbd>
              <Kbd>o</Kbd>
            </Flex>
          ),
          action: () => {},
        },
        {
          label: "Search",
          icon: <SearchIcon className="h-[1.15rem]" />,
          shortcut: <Kbd>/</Kbd>,
          action: () => {},
        },
      ],
    },
    {
      group: "Display",
      items: [
        {
          label: "List View",
          icon: <ListIcon className="h-[1.35rem]" />,
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>v</Kbd>
              <Kbd>l</Kbd>
            </Flex>
          ),
          action: () => {},
        },
        {
          label: "Kanban View",
          icon: <KanbanIcon />,
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>v</Kbd>
              <Kbd>b</Kbd>
            </Flex>
          ),
          action: () => {},
        },
        {
          label: "Show Filters",
          icon: <FilterIcon />,
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>v</Kbd>
              <Kbd>f</Kbd>
            </Flex>
          ),
          action: () => {},
        },
        {
          label: "Toggle Theme",
          icon: <SunIcon />,
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>⌥</Kbd>
              <Kbd>⇧</Kbd>
              <Kbd>t</Kbd>
            </Flex>
          ),
          action: () => {},
        },
      ],
    },
    {
      group: "Settings & Help",
      items: [
        {
          label: "Settings",
          icon: <SettingsIcon />,
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>⌥</Kbd>
              <Kbd>⇧</Kbd>
              <Kbd>s</Kbd>
            </Flex>
          ),
          action: () => {},
        },
        {
          label: "Keyboard Shortcuts",
          icon: <HelpIcon />,
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>⌘</Kbd>
              <Kbd>/</Kbd>
            </Flex>
          ),
          action: () => {},
        },
        {
          label: "Log Out",
          icon: <LogoutIcon />,
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>⌥</Kbd>
              <Kbd>⇧</Kbd>
              <Kbd>l</Kbd>
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
        <Dialog.Body className="px-0 pb-0 pt-2">
          <Command>
            <Command.Input
              className="my-2.5 text-2xl antialiased"
              icon={null}
              placeholder="What do you want to do?"
            />
            <Divider className="my-2.5" />
            <Command.Empty className="py-2">
              <Text color="muted" fontSize="xl" fontWeight="normal">
                No results found.
              </Text>
            </Command.Empty>
            <Box className="max-h-[35rem] overflow-y-auto px-4 pt-2">
              {commands.map((command) => (
                <Command.Group
                  className="mb-4 px-0"
                  heading={
                    <Text
                      className="mb-1.5 pl-3 dark:antialiased"
                      color="muted"
                    >
                      {command.group}
                    </Text>
                  }
                  key={command.group}
                >
                  {command.items.map((item) => (
                    <Command.Item
                      className="justify-between rounded-[0.6rem] p-3 text-lg opacity-85"
                      key={item.label}
                      onSelect={item.action}
                    >
                      <Flex
                        align="center"
                        className="font-medium antialiased"
                        gap={3}
                      >
                        {item.icon}
                        {item.label}
                      </Flex>
                      {item.shortcut}
                    </Command.Item>
                  ))}
                </Command.Group>
              ))}
            </Box>
          </Command>
        </Dialog.Body>
      </Dialog.Content>
    </Dialog>
  );
};
