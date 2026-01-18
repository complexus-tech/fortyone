import { Dialog, Command, Text, Divider, Flex, Box, Kbd } from "ui";
import {
  PlusIcon,
  NotificationsIcon,
  ObjectiveIcon,
  SunIcon,
  HelpIcon,
  LogoutIcon,
  SearchIcon,
  SettingsIcon,
  UserIcon,
  DashboardIcon,
  MoonIcon,
  UsersAddIcon,
} from "icons";
import { useState } from "react";
import { useRouter, usePathname } from "next/navigation";
import { useTheme } from "next-themes";
import { useTerminology, useAnalytics, useUserRole, useWorkspacePath } from "@/hooks";
import { logOut } from "@/components/shared/sidebar/actions";
import { KeyboardShortcuts } from "@/components/shared/keyboard-shortcuts";
import {
  NewObjectiveDialog,
  NewStoryDialog,
  InviteMembersDialog,
} from "@/components/ui";
import { NewSprintDialog } from "@/components/ui/new-sprint-dialog";
import { clearAllStorage } from "./sidebar/utils";

export const CommandBar = ({
  isOpen,
  setIsOpen,
}: {
  isOpen: boolean;
  setIsOpen: (isOpen: boolean) => void;
}) => {
  const { userRole } = useUserRole();
  const { analytics } = useAnalytics();
  const { getTermDisplay } = useTerminology();
  const [isStoryOpen, setIsStoryOpen] = useState(false);
  const [isSprintsOpen, setIsSprintsOpen] = useState(false);
  const [isObjectivesOpen, setIsObjectivesOpen] = useState(false);
  const [isInviteMembersOpen, setIsInviteMembersOpen] = useState(false);
  const router = useRouter();
  const pathname = usePathname();
  const { resolvedTheme: theme, setTheme } = useTheme();
  const [isKeyboardShortcutsOpen, setIsKeyboardShortcutsOpen] = useState(false);
  const { withWorkspace } = useWorkspacePath();

  const toggleTheme = () => {
    setTheme(theme === "dark" ? "light" : "dark");
  };

  const handleLogout = async () => {
    try {
      await logOut();
      analytics.logout(true);
    } finally {
      clearAllStorage();
      window.location.href = "/?signedOut=true";
    }
  };

  const commands = [
    {
      group: "Quick Actions",
      items: [
        {
          label: `New ${getTermDisplay("storyTerm")}`,
          disabled: userRole === "guest",
          icon: <PlusIcon />,
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>⇧</Kbd>
              <Kbd>n</Kbd>
            </Flex>
          ),
          action: () => {
            setIsStoryOpen(true);
            setIsOpen(false);
          },
        },
        {
          label: `New ${getTermDisplay("objectiveTerm")}`,
          disabled: userRole === "guest",
          icon: <PlusIcon />,
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>⇧</Kbd>
              <Kbd>o</Kbd>
            </Flex>
          ),
          action: () => {
            setIsObjectivesOpen(true);
            setIsOpen(false);
          },
        },
        ...(userRole === "admin"
          ? [
              {
                label: "Invite Members",
                icon: <UsersAddIcon />,
                shortcut: (
                  <Flex align="center" gap={1}>
                    <Kbd>⌘</Kbd>
                    <Kbd>i</Kbd>
                  </Flex>
                ),
                action: () => {
                  setIsInviteMembersOpen(true);
                  setIsOpen(false);
                },
              },
            ]
          : []),
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
          action: () => {
            setIsOpen(false);
            if (pathname !== withWorkspace("/my-work")) {
              router.push(withWorkspace("/my-work"));
            }
          },
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
          action: () => {
            setIsOpen(false);
            if (pathname !== withWorkspace("/notifications")) {
              router.push(withWorkspace("/notifications"));
            }
          },
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
          action: () => {
            setIsOpen(false);
            if (pathname !== withWorkspace("/summary")) {
              router.push(withWorkspace("/summary"));
            }
          },
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
          action: () => {
            setIsOpen(false);
            if (pathname !== withWorkspace("/objectives")) {
              router.push(withWorkspace("/objectives"));
            }
          },
        },
        {
          label: "Search",
          icon: <SearchIcon className="h-[1.15rem]" />,
          shortcut: <Kbd>/</Kbd>,
          action: () => {
            setIsOpen(false);
            if (pathname !== withWorkspace("/search")) {
              router.push(withWorkspace("/search"));
            }
          },
        },
      ],
    },
    {
      group: "Settings & Display",
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
          action: () => {
            setIsOpen(false);
            if (pathname !== withWorkspace("/settings")) {
              router.push(withWorkspace("/settings"));
            }
          },
        },
        {
          label: `Toggle ${theme === "dark" ? "light" : "dark"} mode`,
          icon: theme === "dark" ? <SunIcon /> : <MoonIcon />,
          shortcut: (
            <Flex align="center" gap={1}>
              <Kbd>⌥</Kbd>
              <Kbd>⇧</Kbd>
              <Kbd>t</Kbd>
            </Flex>
          ),
          action: () => {
            toggleTheme();
            setIsOpen(false);
          },
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
          action: () => {
            setIsKeyboardShortcutsOpen((prev) => !prev);
            setIsOpen(false);
          },
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
          action: async () => {
            setIsOpen(false);
            await handleLogout();
          },
        },
      ],
    },
  ];
  return (
    <>
      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content className="max-w-3xl" hideClose>
          <Dialog.Header className="sr-only">
            <Dialog.Title className="sr-only">Command Menu</Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="px-0 pt-2 pb-0">
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
              <Box className="max-h-140 overflow-y-auto px-3 pt-2">
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
                        className="justify-between rounded-[0.6rem] p-3 text-[1.1rem] opacity-85"
                        disabled={item.disabled}
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
      <KeyboardShortcuts
        isOpen={isKeyboardShortcutsOpen}
        setIsOpen={setIsKeyboardShortcutsOpen}
      />
      <NewStoryDialog isOpen={isStoryOpen} setIsOpen={setIsStoryOpen} />
      <NewSprintDialog isOpen={isSprintsOpen} setIsOpen={setIsSprintsOpen} />
      <NewObjectiveDialog
        isOpen={isObjectivesOpen}
        setIsOpen={setIsObjectivesOpen}
      />
      {userRole === "admin" && (
        <InviteMembersDialog
          isOpen={isInviteMembersOpen}
          setIsOpen={setIsInviteMembersOpen}
        />
      )}
    </>
  );
};
