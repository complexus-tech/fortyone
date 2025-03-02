"use client";
import { Box, Button, Text } from "ui";
import { Header } from "./header";
import { Navigation } from "./navigation";
import { Teams } from "./teams";

export const Sidebar = () => {
  return (
    <Box className="flex h-screen flex-col justify-between bg-gray-50/60 px-4 pb-4 dark:bg-black">
      <Box>
        <Header />
        <Navigation />
        <Teams />
      </Box>

      <Box className="rounded-xl bg-white p-4 shadow dark:bg-dark-300">
        <Text fontWeight="medium">System Under Development</Text>
        <Text className="mt-2.5" color="muted">
          This is a preview version. Some features may be limited or
          unavailable.
        </Text>
        <Button
          className="mt-3 px-3"
          color="tertiary"
          href="mailto:joseph@complexus.app"
          size="sm"
        >
          Contact developer
        </Button>
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
    </Box>
  );
};
