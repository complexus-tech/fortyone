"use client";
import { Box, Button, Text } from "ui";
import { EmailIcon, PlusIcon } from "icons";
import { useState } from "react";
import { InviteMembersDialog } from "@/components/ui";
import { Header } from "./header";
import { Navigation } from "./navigation";
import { Teams } from "./teams";

export const Sidebar = () => {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <Box className="flex h-screen flex-col justify-between bg-gray-50/40 px-4 pb-6 dark:bg-black">
      <Box>
        <Header />
        <Navigation />
        <Teams />
      </Box>

      <Box>
        <Box className="rounded-xl bg-white p-4 shadow-lg shadow-gray-100 dark:bg-dark-200 dark:shadow-none">
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
            leftIcon={<EmailIcon className="h-4" />}
            size="sm"
          >
            Contact developer
          </Button>
        </Box>
        <button
          className="mt-3 flex items-center gap-2 px-1"
          onClick={() => {
            setIsOpen(true);
          }}
          type="button"
        >
          <PlusIcon />
          Invite members
        </button>
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
    </Box>
  );
};
