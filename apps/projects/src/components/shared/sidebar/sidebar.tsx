"use client";
import { Box, Button, Flex, Menu, Text, Tooltip } from "ui";
import { CommandIcon, DocsIcon, EmailIcon, HelpIcon, PlusIcon } from "icons";
import { useState } from "react";
import { InviteMembersDialog } from "@/components/ui";
import { KeyboardShortcuts } from "@/components/shared/keyboard-shortcuts";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { useUserRole } from "@/hooks";
import { CommandMenu } from "../command-menu";
import { Header } from "./header";
import { Navigation } from "./navigation";
import { Teams } from "./teams";
import { ProfileMenu } from "./profile-menu";

export const Sidebar = () => {
  const [isOpen, setIsOpen] = useState(false);
  const [isKeyboardShortcutsOpen, setIsKeyboardShortcutsOpen] = useState(false);
  const { tier, trialDaysRemaining } = useSubscriptionFeatures();
  const { userRole } = useUserRole();

  return (
    <Box className="flex h-dvh flex-col justify-between bg-gray-50/80 pb-6 dark:bg-[#000000]/45">
      <Box className="px-4">
        <Header />
        <Navigation />
        <Teams />
      </Box>
      <Box>
        <Box className="mb-2.5 px-3.5">
          {tier === "free" && (
            <Box className="rounded-xl border-[0.5px] border-gray-200/60 bg-white p-4 shadow-lg shadow-gray-100 dark:border-dark-50 dark:bg-dark-300 dark:shadow-none">
              <Text fontWeight="medium">You&apos;re on the free plan</Text>
              <Text className="mt-2" color="muted">
                {userRole === "admin"
                  ? "Upgrade to a paid plan to get more features."
                  : "Ask your admin to upgrade to a paid plan to get more features."}
              </Text>
              {userRole === "admin" && (
                <Button
                  className="mt-3 px-3"
                  color="tertiary"
                  href="/settings/workspace/billing"
                  prefetch
                  size="sm"
                >
                  Upgrade plan
                </Button>
              )}
            </Box>
          )}

          {tier === "trial" && (
            <Tooltip
              className="ml-2 max-w-56 py-3"
              title={`${trialDaysRemaining} days left in your trial. ${userRole === "admin" ? "Upgrade" : "Ask your admin to upgrade"} to a paid plan to get more premium features.`}
            >
              <span>
                <Button
                  className="mt-3 border-opacity-15 bg-opacity-10 px-3 text-primary dark:bg-opacity-15"
                  href={
                    userRole === "admin"
                      ? "/settings/workspace/billing"
                      : undefined
                  }
                  prefetch
                  rounded="lg"
                  size="sm"
                >
                  {trialDaysRemaining} day{trialDaysRemaining !== 1 ? "s" : ""}{" "}
                  left in trial
                </Button>
              </span>
            </Tooltip>
          )}

          <Flex align="center" className="mt-3 gap-3" justify="between">
            {userRole === "admin" ? (
              <button
                className="flex items-center gap-2 px-1"
                data-invite-button
                onClick={() => {
                  setIsOpen(true);
                }}
                type="button"
              >
                <PlusIcon />
                Invite members
              </button>
            ) : null}

            <Menu>
              <Menu.Button>
                <Button
                  asIcon
                  className="border-[0.5px]"
                  color="tertiary"
                  data-help-button
                  rounded="full"
                  variant="naked"
                >
                  <HelpIcon className="h-6" />
                </Button>
              </Menu.Button>
              <Menu.Items align="end">
                <Menu.Group>
                  <Menu.Item
                    onSelect={() => {
                      setIsKeyboardShortcutsOpen(true);
                    }}
                  >
                    <CommandIcon />
                    Keyboard shortcuts
                  </Menu.Item>
                  <Menu.Item
                    onSelect={() => {
                      window.open("mailto:support@complexus.app", "_blank");
                    }}
                  >
                    <EmailIcon />
                    Contact support
                  </Menu.Item>
                  <Menu.Item
                    onSelect={() => {
                      window.open("https://docs.complexus.app", "_blank");
                    }}
                  >
                    <DocsIcon />
                    Documentation
                  </Menu.Item>
                </Menu.Group>
              </Menu.Items>
            </Menu>
          </Flex>
        </Box>
        <ProfileMenu />
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
