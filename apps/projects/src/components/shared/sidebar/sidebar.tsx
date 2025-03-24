"use client";
import { Box, Button, Flex, Menu, Text } from "ui";
import { CommandIcon, DocsIcon, EmailIcon, HelpIcon, PlusIcon } from "icons";
import { useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import { useHotkeys } from "react-hotkeys-hook";
import nProgress from "nprogress";
import { useTheme } from "next-themes";
import { InviteMembersDialog } from "@/components/ui";
import { KeyboardShortcuts } from "../keyboard-shortcuts";
import { CommandMenu } from "../command-menu";
import { Header } from "./header";
import { Navigation } from "./navigation";
import { Teams } from "./teams";

export const Sidebar = () => {
  const [isOpen, setIsOpen] = useState(false);
  const [isKeyboardShortcutsOpen, setIsKeyboardShortcutsOpen] = useState(false);
  const router = useRouter();
  const pathname = usePathname();
  const { theme, setTheme } = useTheme();
  // Navigation shortcuts
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

  useHotkeys("alt+shift+t", () => {
    setTheme(theme === "dark" ? "light" : "dark");
  });

  useHotkeys("mod+/", () => {
    setIsKeyboardShortcutsOpen((prev) => !prev);
  });
  useHotkeys("/", () => {
    if (pathname !== "search") {
      nProgress.start();
      router.push("/search");
    }
  });

  return (
    <Box className="flex h-screen flex-col justify-between bg-gray-50/60 px-4 pb-6 dark:bg-black">
      <Box>
        <Header />
        <Navigation />
        <Teams />
      </Box>
      <Box>
        <Box className="rounded-xl border-[0.5px] border-gray-200/60 bg-white p-4 shadow-lg shadow-gray-100 dark:border-dark-50 dark:bg-dark-300 dark:shadow-none">
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
      <CommandMenu />
    </Box>
  );
};
