/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { resolveTeamSettingsLocation } from "./navigation";

describe("team settings locations", () => {
  it.each([
    ["sprints", "planning", "sprints"],
    ["delete", "general", "danger"],
  ] as const)(
    "preserves the legacy %s bookmark's destination",
    (tab, expectedTab, section) => {
      expect(resolveTeamSettingsLocation(tab, null)).toEqual({
        tab: expectedTab,
        section,
      });
    },
  );

  it.each([
    ["general", "danger"],
    ["planning", "complexity"],
    ["planning", "sprints"],
    ["fields", "fields"],
    ["automations", "github"],
  ] as const)("resolves %s with its %s section", (tab, section) => {
    expect(resolveTeamSettingsLocation(tab, section)).toEqual({
      tab,
      section,
    });
  });

  it("ignores sections from other tabs and unrecognized section names", () => {
    expect(resolveTeamSettingsLocation("members", "fields")).toEqual({
      tab: "members",
      section: null,
    });
    expect(resolveTeamSettingsLocation("automations", "constructor")).toEqual({
      tab: "automations",
      section: null,
    });
    expect(resolveTeamSettingsLocation("general", "unknown")).toEqual({
      tab: "general",
      section: null,
    });
  });

  it.each([
    ["workflows", "fields", "fields"],
    ["workflows", "complexity", "planning"],
    ["automations", "sprints", "planning"],
  ] as const)(
    "redirects the former %s %s section to its current %s tab",
    (tab, section, expectedTab) => {
      expect(resolveTeamSettingsLocation(tab, section)).toEqual({
        tab: expectedTab,
        section,
      });
    },
  );

  it("opens Custom fields directly through its existing bookmarked tab", () => {
    expect(resolveTeamSettingsLocation("fields", null)).toEqual({
      tab: "fields",
      section: null,
    });
  });

  it("retains the legacy destination when its explicit section belongs elsewhere", () => {
    expect(resolveTeamSettingsLocation("sprints", "fields")).toEqual({
      tab: "planning",
      section: "sprints",
    });
  });
});
