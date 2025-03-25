"use client";

import { useState } from "react";
import { useHotkeys } from "react-hotkeys-hook";
import { Dialog, Command, Text, Divider, Kbd, Flex, Box } from "ui";
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
import nProgress from "nprogress";
import { useRouter, usePathname } from "next/navigation";
import { useTheme } from "next-themes";
import { useTerminology, useAnalytics } from "@/hooks";
import { KeyboardShortcuts } from "@/components/shared/keyboard-shortcuts";
import {
  NewObjectiveDialog,
  NewStoryDialog,
  InviteMembersDialog,
} from "@/components/ui";
import { NewSprintDialog } from "@/components/ui/new-sprint-dialog";
import { logOut } from "@/components/shared/sidebar/actions";

const clearAllStorage = () => {
  document.cookie.split(";").forEach((c) => {
    document.cookie = c
      .replace(/^ +/, "")
      .replace(/=.*/, `=;expires=${new Date().toUTCString()};path=/`);
  });

  if ("caches" in window) {
    caches.keys().then((names) => {
      names.forEach((name) => {
        caches.delete(name);
      });
    });
  }
};

export const CommandMenu = () => {
  const { analytics } = useAnalytics();
  const { getTermDisplay } = useTerminology();
  const [isStoryOpen, setIsStoryOpen] = useState(false);
  const [isSprintsOpen, setIsSprintsOpen] = useState(false);
  const [isObjectivesOpen, setIsObjectivesOpen] = useState(false);
  const [isInviteMembersOpen, setIsInviteMembersOpen] = useState(false);
  const [open, setOpen] = useState(false);
  const router = useRouter();
  const pathname = usePathname();
  const { resolvedTheme: theme, setTheme } = useTheme();
  const [isKeyboardShortcutsOpen, setIsKeyboardShortcutsOpen] = useState(false);

  const toggleTheme = () => {
    setTheme(theme === "dark" ? "light" : "dark");
  };

  const handleLogout = async () => {
    try {
      await logOut();
      analytics.logout(true);
    } finally {
      clearAllStorage();
      window.location.href = "https://www.complexus.app?signedOut=true";
    }
  };

  useHotkeys("mod+k", (e) => {
    e.preventDefault();
    setOpen((prev) => !prev);
  });

  useHotkeys("g+i", () => {
    if (pathname !== "/notifications") {
      nProgress.start();
      router.push("/notifications");
    }
  });
  useHotkeys("g+m", () => {
    if (pathname !== "/my-work") {
      nProgress.start();
      router.push("/my-work");
    }
  });

  useHotkeys("g+s", () => {
    if (pathname !== "/summary") {
      nProgress.start();
      router.push("/summary");
    }
  });
  useHotkeys("g+o", () => {
    if (pathname !== "/objectives") {
      nProgress.start();
      router.push("/objectives");
    }
  });

  useHotkeys("alt+shift+s", () => {
    if (pathname !== "/settings") {
      nProgress.start();
      router.push("/settings");
    }
  });

  useHotkeys("alt+shift+t", () => {});

  useHotkeys("mod+/", () => {
    setIsKeyboardShortcutsOpen((prev) => !prev);
  });
  useHotkeys("mod+i", () => {
    setIsInviteMembersOpen((prev) => !prev);
    setOpen(false);
  });

  useHotkeys("/", () => {
    if (pathname !== "search") {
      nProgress.start();
      router.push("/search");
    }
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
          action: () => {
            setIsStoryOpen(true);
            setOpen(false);
          },
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
          action: () => {
            setIsObjectivesOpen(true);
            setOpen(false);
          },
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
          action: () => {
            setIsSprintsOpen(true);
            setOpen(false);
          },
        },
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
            setOpen(false);
          },
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
          action: () => {
            setOpen(false);
            if (pathname !== "/my-work") {
              nProgress.start();
              router.push("/my-work");
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
            setOpen(false);
            if (pathname !== "/notifications") {
              nProgress.start();
              router.push("/notifications");
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
            setOpen(false);
            if (pathname !== "/summary") {
              nProgress.start();
              router.push("/summary");
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
            setOpen(false);
            if (pathname !== "/objectives") {
              nProgress.start();
              router.push("/objectives");
            }
          },
        },
        {
          label: "Search",
          icon: <SearchIcon className="h-[1.15rem]" />,
          shortcut: <Kbd>/</Kbd>,
          action: () => {
            setOpen(false);
            if (pathname !== "search") {
              nProgress.start();
              router.push("/search");
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
            setOpen(false);
            if (pathname !== "/settings") {
              nProgress.start();
              router.push("/settings");
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
            setOpen(false);
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
            setOpen(false);
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
            setOpen(false);
            await handleLogout();
          },
        },
      ],
    },
  ];

  return (
    <>
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
              <Box className="max-h-[35rem] overflow-y-auto px-3 pt-2">
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
      <InviteMembersDialog
        isOpen={isInviteMembersOpen}
        setIsOpen={setIsInviteMembersOpen}
      />
    </>
  );
};
