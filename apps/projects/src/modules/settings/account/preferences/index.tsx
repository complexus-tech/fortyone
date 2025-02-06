"use client";

import { Box, Text } from "ui";
import { useUserRole } from "@/hooks";
import { Theming } from "./components/theming";
import { Notifications } from "./components/notifications";
import { Automations } from "./components/automations";

export const PreferencesSettings = () => {
  const { userRole } = useUserRole();

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Preferences
      </Text>
      <Notifications />
      <Theming />
      {userRole !== "guest" && <Automations />}
    </Box>
  );
};
