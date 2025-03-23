"use client";
import { Badge, Box, Button, Dialog, Flex, Menu, Text } from "ui";
import { CommandIcon, DocsIcon, EmailIcon, HelpIcon, PlusIcon } from "icons";
import type { ReactNode } from "react";
import { useState } from "react";
import { InviteMembersDialog } from "@/components/ui";
import { Header } from "./header";
import { Navigation } from "./navigation";
import { Teams } from "./teams";

type Shortcuts = {
  name: string;
  items: {
    name: string;
    shortcut: ReactNode;
  }[];
};

const Kbd = ({ children }: { children: ReactNode }) => (
  <Badge
    className="h-6 min-w-6 rounded px-1 uppercase dark:border-dark-50"
    color="tertiary"
    size="sm"
  >
    {children}
  </Badge>
);

const shortcuts: Shortcuts[] = [
  {
    name: "General",
    items: [
      {
        name: "Open command menu",
        shortcut: (
          <Flex align="center" gap={1}>
            <Kbd>⌘</Kbd>
            <Kbd>K</Kbd>
          </Flex>
        ),
      },
      {
        name: "Search",
        shortcut: <Kbd>/</Kbd>,
      },
      {
        name: "Show keyboard shortcut help",
        shortcut: <Kbd>?</Kbd>,
      },
      {
        name: "Undo",
        shortcut: (
          <Flex align="center" gap={1}>
            <Kbd>⌘</Kbd>
            <Kbd>Z</Kbd>
          </Flex>
        ),
      },
      {
        name: "Redo",
        shortcut: (
          <Flex align="center" gap={1}>
            <Kbd>⌘</Kbd>
            <Kbd>⇧</Kbd>
            <Kbd>Z</Kbd>
          </Flex>
        ),
      },
      {
        name: "Go to settings",
        shortcut: (
          <Flex align="center" gap={1}>
            <Kbd>⌘</Kbd>
            <Kbd>⇧</Kbd>
            <Kbd>s</Kbd>
          </Flex>
        ),
      },
      {
        name: "Log out",
        shortcut: (
          <Flex align="center" gap={1}>
            <Kbd>⌘</Kbd>
            <Kbd>⇧</Kbd>
            <Kbd>q</Kbd>
          </Flex>
        ),
      },
    ],
  },
  {
    name: "Navigation",
    items: [
      {
        name: "Go to inbox",
        shortcut: (
          <Flex align="center" gap={1}>
            <Kbd>g</Kbd>
            <Kbd>i</Kbd>
          </Flex>
        ),
      },
      {
        name: "Go to my work",
        shortcut: (
          <Flex align="center" gap={1}>
            <Kbd>g</Kbd>
            <Kbd>m</Kbd>
          </Flex>
        ),
      },
      {
        name: "Go to summary",
        shortcut: (
          <Flex align="center" gap={1}>
            <Kbd>g</Kbd>
            <Kbd>s</Kbd>
          </Flex>
        ),
      },
      {
        name: "Go to projects",
        shortcut: (
          <Flex align="center" gap={1}>
            <Kbd>g</Kbd>
            <Kbd>o</Kbd>
          </Flex>
        ),
      },
    ],
  },
  {
    name: "Creation",
    items: [
      {
        name: "Create new item",
        shortcut: (
          <Flex align="center" gap={1}>
            <Kbd>⇧</Kbd>
            <Kbd>n</Kbd>
          </Flex>
        ),
      },
      {
        name: "Create new Objective/Project/Goal",
        shortcut: (
          <Flex align="center" gap={1}>
            <Kbd>⇧</Kbd>
            <Kbd>o</Kbd>
          </Flex>
        ),
      },
      {
        name: "Create new Sprint/Cycle/Iteration",
        shortcut: (
          <Flex align="center" gap={1}>
            <Kbd>⇧</Kbd>
            <Kbd>s</Kbd>
          </Flex>
        ),
      },
      {
        name: "Create new Key Result/Milestone/Focus Area",
        shortcut: (
          <Flex align="center" gap={1}>
            <Kbd>⇧</Kbd>
            <Kbd>k</Kbd>
          </Flex>
        ),
      },
    ],
  },
  {
    name: "Work Item Management",
    items: [
      {
        name: "Delete selected item",
        shortcut: <Kbd>Delete</Kbd>,
      },
      {
        name: "Comment on selected item",
        shortcut: <Kbd>c</Kbd>,
      },
      {
        name: "Assign selected item",
        shortcut: <Kbd>a</Kbd>,
      },
      {
        name: "Add label",
        shortcut: <Kbd>l</Kbd>,
      },
      {
        name: "Change status",
        shortcut: <Kbd>s</Kbd>,
      },
      {
        name: "Change priority",
        shortcut: <Kbd>p</Kbd>,
      },
      {
        name: "Set due date",
        shortcut: <Kbd>d</Kbd>,
      },
    ],
  },
  {
    name: "View Controls",
    items: [
      {
        name: "View filters",
        shortcut: <Kbd>f</Kbd>,
      },
      {
        name: "List view",
        shortcut: (
          <Flex align="center" gap={1}>
            <Kbd>v</Kbd>
            <Kbd>l</Kbd>
          </Flex>
        ),
      },
      {
        name: "Board view",
        shortcut: (
          <Flex align="center" gap={1}>
            <Kbd>v</Kbd>
            <Kbd>b</Kbd>
          </Flex>
        ),
      },
    ],
  },
];

const KeyboardShortcuts = ({
  isOpen,
  setIsOpen,
}: {
  isOpen: boolean;
  setIsOpen: (isOpen: boolean) => void;
}) => {
  return (
    <Dialog onOpenChange={setIsOpen} open={isOpen}>
      <Dialog.Content
        className="mb-auto mt-auto"
        overlayClassName="justify-end pr-[2vh]"
        size="sm"
      >
        <Dialog.Title className="px-6 py-4 text-lg">
          Keyboard Shortcuts
        </Dialog.Title>
        <Dialog.Body className="h-[91vh] max-h-[91vh] pt-2">
          {shortcuts.map((shortcut) => (
            <Box className="mb-7" key={shortcut.name}>
              <Text className="mb-3" fontWeight="medium">
                {shortcut.name}
              </Text>
              <Box className="space-y-3">
                {shortcut.items.map((item) => (
                  <Flex
                    align="center"
                    gap={3}
                    justify="between"
                    key={item.name}
                  >
                    <Text color="muted">{item.name}</Text>
                    <Box>{item.shortcut}</Box>
                  </Flex>
                ))}
              </Box>
            </Box>
          ))}
        </Dialog.Body>
      </Dialog.Content>
    </Dialog>
  );
};

export const Sidebar = () => {
  const [isOpen, setIsOpen] = useState(false);
  const [isKeyboardShortcutsOpen, setIsKeyboardShortcutsOpen] = useState(false);

  return (
    <Box className="flex h-screen flex-col justify-between bg-gray-50/60 px-4 pb-6 dark:bg-black">
      <Box>
        <Header />
        <Navigation />
        <Teams />
      </Box>

      <Box>
        <Box className="rounded-xl border-[0.5px] border-gray-100 bg-white p-4 shadow-lg shadow-gray-100 dark:border-dark-50 dark:bg-dark-300 dark:shadow-none">
          <Text color="gradient" fontWeight="medium">
            System Under Development
          </Text>
          <Text className="mt-2.5" color="muted">
            This is a preview version. Some features may be limited or
            unavailable.
          </Text>
          <Button
            className="mt-3"
            color="tertiary"
            href="mailto:joseph@complexus.app"
            leftIcon={<EmailIcon />}
            size="sm"
          >
            Contact developer
          </Button>
        </Box>
        <Flex align="center" className="mt-3 gap-3" justify="between">
          <button
            className="flex items-center gap-2 px-1"
            onClick={() => {
              setIsOpen(true);
            }}
            type="button"
          >
            <PlusIcon />
            Invite members
          </button>
          <Menu>
            <Menu.Button>
              <Button asIcon color="tertiary" rounded="full" variant="naked">
                <HelpIcon className="h-6" />
              </Button>
            </Menu.Button>
            <Menu.Items align="end">
              <Menu.Group>
                <Menu.Item
                  onClick={() => {
                    setIsKeyboardShortcutsOpen(true);
                  }}
                >
                  <CommandIcon />
                  Keyboard shortcuts
                </Menu.Item>
                <Menu.Item disabled>
                  <EmailIcon />
                  Contact support
                </Menu.Item>
                <Menu.Item disabled>
                  <DocsIcon />
                  Documentation
                </Menu.Item>
              </Menu.Group>
            </Menu.Items>
          </Menu>
        </Flex>
      </Box>

      {/* <Box className="rounded-xl bg-white p-4 shadow dark:bg-dark-300">
        <Text fontWeight="medium">You&apos;re on the free plan</Text>
        <Text className="mt-2.5" color="muted">
          You can upgrade to a paid plan to get more features.
        </Text>
        <Button className="mt-3 px-3" color="tertiary" size="sm">
          Upgrade to Pro
        </Button>
      </Box> */}

      <InviteMembersDialog isOpen={isOpen} setIsOpen={setIsOpen} />
      <KeyboardShortcuts
        isOpen={isKeyboardShortcutsOpen}
        setIsOpen={setIsKeyboardShortcutsOpen}
      />
    </Box>
  );
};
