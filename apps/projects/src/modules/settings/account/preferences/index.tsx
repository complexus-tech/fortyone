"use client";

import { Box, Text } from "ui";
import { useUserRole } from "@/hooks";
import { Theming } from "./components/theming";
import { Timezone } from "./components/timezone";
import { Automations } from "./components/automations";

export const PreferencesSettings = () => {
  const { userRole } = useUserRole();

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-medium">
        Preferences
      </Text>
      <Theming />
      <Timezone />
      {userRole !== "guest" && <Automations />}
    </Box>
  );
};
