"use client";

import type { ReactNode } from "react";
import { PreferencesIcon, SystemIcon } from "icons";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import { Box, Tabs, Text } from "ui";
import { useUserRole } from "@/hooks";
import { Theming } from "./components/theming";
import { Timezone } from "./components/timezone";
import { Automations } from "./components/automations";
import { PersonalWorkSchedule } from "./components/work-schedule";

const PREFERENCE_TABS = ["general", "desktop"] as const;

const GeneralPreferences = () => {
  const { userRole } = useUserRole();

  return (
    <>
      <Theming />
      <Timezone />
      <PersonalWorkSchedule />
      {userRole !== "guest" && <Automations />}
    </>
  );
};

const PreferenceTabs = ({
  desktopSettings,
}: {
  desktopSettings: ReactNode;
}) => {
  const [tab, setTab] = useQueryState(
    "tab",
    parseAsStringLiteral(PREFERENCE_TABS)
      .withDefault("general")
      .withOptions({ history: "push" }),
  );

  return (
    <Tabs
      onValueChange={(value) => {
        if (value === "general" || value === "desktop") void setTab(value);
      }}
      value={tab}
    >
      <Box className="overflow-x-auto">
        <Tabs.List
          aria-label="Preference settings"
          className="mx-0 flex-nowrap md:mx-0"
        >
          <Tabs.Tab
            leftIcon={<PreferencesIcon className="h-[1.1rem]" />}
            value="general"
          >
            General
          </Tabs.Tab>
          <Tabs.Tab
            leftIcon={<SystemIcon className="h-[1.1rem]" />}
            value="desktop"
          >
            Desktop
          </Tabs.Tab>
        </Tabs.List>
      </Box>
      <Box className="mt-5">
        <Tabs.Panel className="[&>div:first-child]:mt-0" value="general">
          <GeneralPreferences />
        </Tabs.Panel>
        <Tabs.Panel value="desktop">{desktopSettings}</Tabs.Panel>
      </Box>
    </Tabs>
  );
};

export const PreferencesSettings = ({
  desktopSettings,
}: {
  desktopSettings?: ReactNode;
}) => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-medium">
        Preferences
      </Text>
      {desktopSettings ? (
        <PreferenceTabs desktopSettings={desktopSettings} />
      ) : (
        <GeneralPreferences />
      )}
    </Box>
  );
};
