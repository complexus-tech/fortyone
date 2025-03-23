"use client";
import { Box, Button, Dialog, Flex, Menu, Text } from "ui";
import { CommandIcon, DocsIcon, EmailIcon, HelpIcon, PlusIcon } from "icons";
import { useState } from "react";
import { InviteMembersDialog } from "@/components/ui";
import { Header } from "./header";
import { Navigation } from "./navigation";
import { Teams } from "./teams";


const KeyboardShortcuts = ({ isOpen, setIsOpen }: { isOpen: boolean; setIsOpen: (isOpen: boolean) => void }) => {
  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <Dialog.Content  className="mb-auto mt-auto"
          // overlayClassName="justify-end pr-[2.5vh]"
          >
        <Dialog.Title>Keyboard Shortcuts</Dialog.Title>
        <Dialog.Description>
          Use the following shortcuts to navigate through the app.
        </Dialog.Description>
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
              <Button color="tertiary" variant="naked" asIcon rounded="full">
                <HelpIcon className="h-6" />
              </Button>
            </Menu.Button>
            <Menu.Items align="end">
              <Menu.Group>
                <Menu.Item onClick={() => setIsKeyboardShortcutsOpen(true)}>
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
      <KeyboardShortcuts isOpen setIsOpen={setIsKeyboardShortcutsOpen} />
    </Box>
  );
};
