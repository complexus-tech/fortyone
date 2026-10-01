export const TEAM_SETTINGS_TABS = [
  "general",
  "members",
  "workflows",
  "planning",
  "fields",
  "automations",
] as const;

export const TEAM_SETTINGS_QUERY_TABS = [
  ...TEAM_SETTINGS_TABS,
  "sprints",
  "delete",
] as const;

export type TeamSettingsTab = (typeof TEAM_SETTINGS_TABS)[number];
export type TeamSettingsQueryTab = (typeof TEAM_SETTINGS_QUERY_TABS)[number];

const SECTION_TABS = {
  general: "general",
  danger: "general",
  workflows: "workflows",
  complexity: "planning",
  fields: "fields",
  rules: "automations",
  cleanup: "automations",
  sprints: "planning",
  github: "automations",
} as const satisfies Record<string, TeamSettingsTab>;

export type TeamSettingsSection = keyof typeof SECTION_TABS;

type TeamSettingsLocation = {
  tab: TeamSettingsTab;
  section: TeamSettingsSection | null;
};

const LEGACY_LOCATIONS: Partial<
  Record<TeamSettingsQueryTab, TeamSettingsLocation>
> = {
  sprints: { tab: "planning", section: "sprints" },
  delete: { tab: "general", section: "danger" },
};

export const resolveTeamSettingsLocation = (
  tab: TeamSettingsQueryTab,
  section: string | null,
): TeamSettingsLocation => {
  const location = LEGACY_LOCATIONS[tab] ?? {
    tab: tab as TeamSettingsTab,
    section: null,
  };
  if (section && Object.hasOwn(SECTION_TABS, section)) {
    const target = section as TeamSettingsSection;
    if (SECTION_TABS[target] === location.tab) {
      return { tab: location.tab, section: target };
    }
    // Keep links from the combined settings panels pointing at their content.
    if (
      (tab === "workflows" &&
        (target === "fields" || target === "complexity")) ||
      (tab === "automations" && target === "sprints")
    ) {
      return { tab: SECTION_TABS[target], section: target };
    }
  }
  return location;
};

export const getTeamSettingsSectionId = (section: TeamSettingsSection) =>
  `team-settings-${section}`;
