"use client";
import { Kbd, Box, Dialog, Flex, Text } from "ui";
import type { ReactNode } from "react";
import { useTerminology } from "@/hooks";

type Shortcuts = {
  name: string;
  items: {
    name: string;
    shortcut: ReactNode;
  }[];
};

export const KeyboardShortcuts = ({
  isOpen,
  setIsOpen,
}: {
  isOpen: boolean;
  setIsOpen: (isOpen: boolean) => void;
}) => {
  const { getTermDisplay } = useTerminology();
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
          name: "Show keyboard shortcuts",
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>⌘</Kbd>
              <Kbd>/</Kbd>
            </Flex>
          ),
        },
        {
          name: "Open AI chat",
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>⇧</Kbd>
              <Kbd>M</Kbd>
            </Flex>
          ),
        },
        {
          name: "Search",
          shortcut: <Kbd>/</Kbd>,
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
              <Kbd>⌥</Kbd>
              <Kbd>⇧</Kbd>
              <Kbd>s</Kbd>
            </Flex>
          ),
        },
        {
          name: "Log out",
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>⌥</Kbd>
              <Kbd>⇧</Kbd>
              <Kbd>l</Kbd>
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
          name: `Go to my ${getTermDisplay("storyTerm", { variant: "plural" })}`,
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
          name: `Go to ${getTermDisplay("objectiveTerm", { variant: "plural" })}`,
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
          name: `Create new ${getTermDisplay("storyTerm")}`,
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>⇧</Kbd>
              <Kbd>n</Kbd>
            </Flex>
          ),
        },
        {
          name: `Create new ${getTermDisplay("objectiveTerm")}`,
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>⇧</Kbd>
              <Kbd>o</Kbd>
            </Flex>
          ),
        },
        {
          name: `Create new ${getTermDisplay("sprintTerm")}`,
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>⇧</Kbd>
              <Kbd>s</Kbd>
            </Flex>
          ),
        },
      ],
    },
    {
      name: "View Controls",
      items: [
        {
          name: "View filters",
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>v</Kbd>
              <Kbd>f</Kbd>
            </Flex>
          ),
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
          name: "Kanban view",
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>v</Kbd>
              <Kbd>k</Kbd>
            </Flex>
          ),
        },
        {
          name: "Toggle theme",
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>⌥</Kbd>
              <Kbd>⇧</Kbd>
              <Kbd>t</Kbd>
            </Flex>
          ),
        },
      ],
    },
    {
      name: getTermDisplay("storyTerm", {
        capitalize: true,
      }),
      items: [
        {
          name: `Delete selected ${getTermDisplay("storyTerm")}`,
          shortcut: <Kbd>Delete</Kbd>,
        },
        {
          name: "Change assignee",
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
        {
          name: `Create child ${getTermDisplay("storyTerm")}`,
          shortcut: <Kbd>c</Kbd>,
        },
        {
          name: `Previous ${getTermDisplay("storyTerm")} (in dialog)`,
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>↑</Kbd>
              <Text color="muted">or</Text>
              <Kbd>←</Kbd>
            </Flex>
          ),
        },
        {
          name: `Next ${getTermDisplay("storyTerm")} (in dialog)`,
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>↓</Kbd>
              <Text color="muted">or</Text>
              <Kbd>→</Kbd>
            </Flex>
          ),
        },
      ],
    },
  ];
  return (
    <Dialog onOpenChange={setIsOpen} open={isOpen}>
      <Dialog.Content
        className="md:mb-auto md:mt-auto"
        overlayClassName="justify-end pr-[1vh]"
        size="sm"
      >
        <Dialog.Title className="px-6 py-4 text-lg">
          Keyboard Shortcuts
        </Dialog.Title>
        <Dialog.Body className="h-[91dvh] max-h-[91dvh] pt-2">
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
