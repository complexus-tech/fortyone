"use client";
import { Box, Button, Flex, Menu, Text } from "ui";
import { CommandIcon, DocsIcon, EmailIcon, HelpIcon, PlusIcon } from "icons";
import { useState } from "react";
import { InviteMembersDialog } from "@/components/ui";
import { KeyboardShortcuts } from "@/components/shared/keyboard-shortcuts";
import { useSubscription } from "@/lib/hooks/subscriptions/subscription";
import { CommandMenu } from "../command-menu";
import { Header } from "./header";
import { Navigation } from "./navigation";
import { Teams } from "./teams";

export const Sidebar = () => {
  const { data: subscription } = useSubscription();

  const plan = subscription?.tier;
  const status = subscription?.status;

  const [isOpen, setIsOpen] = useState(false);
  const [isKeyboardShortcutsOpen, setIsKeyboardShortcutsOpen] = useState(false);

  return (
    <Box className="flex h-dvh flex-col justify-between bg-gray-50/60 px-4 pb-6 dark:bg-[#000000]/45">
      <Box>
        <Header />
        <Navigation />
        <Teams />
      </Box>
      <Box>
        {(plan === "free" || status !== "active") && (
          <Box className="rounded-xl border-[0.5px] border-gray-200/60 bg-white p-4 shadow-lg shadow-gray-100 dark:border-dark-50 dark:bg-dark-300 dark:shadow-none">
            <Text fontWeight="medium">You&apos;re on the free plan</Text>
            <Text className="mt-2" color="muted">
              Upgrade to a paid plan to get more features.
            </Text>
            <Button
              className="mt-3 px-3"
              color="tertiary"
              href="/settings/workspace/billing"
              size="sm"
            >
              Upgrade plan
            </Button>
          </Box>
        )}

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
              <Button
                asIcon
                className="border-[0.5px]"
                color="tertiary"
                rounded="full"
              >
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
      <KeyboardShortcuts
        isOpen={isKeyboardShortcutsOpen}
        setIsOpen={setIsKeyboardShortcutsOpen}
      />
      <InviteMembersDialog isOpen={isOpen} setIsOpen={setIsOpen} />
      <CommandMenu />
    </Box>
  );
};
