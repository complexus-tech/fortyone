/* global describe, expect, it -- Jest globals are provided by the projects test runner. */
import {
  buildSettingsNavigation,
  isSettingsItemActive,
} from "./settings-navigation";

const withWorkspace = (path: string) => `/acme${path}`;

describe("settings navigation", () => {
  it("builds the four admin groups in their intended order", () => {
    const navigation = buildSettingsNavigation({
      userRole: "admin",
      hasCustomTerminology: true,
      hasInvitations: true,
      objectiveTitle: "Goals",
      withWorkspace,
    });

    expect(navigation.map(({ category }) => category)).toEqual([
      "Account",
      "Workspace",
      "Administration",
      "Features",
    ]);
    expect(navigation[1]?.items.map(({ title }) => title)).toEqual([
      "General",
      "Terminology",
      "API",
      "Integrations",
    ]);
    expect(navigation[2]?.items.map(({ title }) => title)).toEqual([
      "Members",
      "Billing & plans",
      "Imports",
    ]);
    expect(navigation[3]?.items.map(({ title }) => title)).toEqual([
      "Labels",
      "Goals",
      "Teams",
      "Feedback",
    ]);
    expect(navigation[0]?.items.at(-1)?.title).toBe("Invitations");
    expect(navigation[0]?.items).toEqual(
      expect.arrayContaining([
        {
          href: "/acme/settings/account/google-drive",
          title: "Google Drive",
        },
      ]),
    );
  });

  it.each(["member", "guest"] as const)(
    "keeps Integrations available to a %s without exposing Administration",
    (userRole) => {
      const navigation = buildSettingsNavigation({
        userRole,
        hasCustomTerminology: true,
        hasInvitations: false,
        objectiveTitle: "Objectives",
        withWorkspace,
      });

      expect(navigation.map(({ category }) => category)).toEqual([
        "Account",
        "Workspace",
      ]);
      expect(navigation[1]?.items.map(({ title }) => title)).toEqual([
        "API",
        "Integrations",
      ]);
      expect(
        navigation.flatMap(({ items }) => items).map(({ title }) => title),
      ).not.toContain("Imports");
      expect(
        navigation.flatMap(({ items }) => items).map(({ title }) => title),
      ).toContain("Google Drive");
    },
  );

  it("matches nested integration and import settings routes", () => {
    const navigation = buildSettingsNavigation({
      userRole: "admin",
      hasCustomTerminology: false,
      hasInvitations: false,
      objectiveTitle: "Objectives",
      withWorkspace,
    });
    const items = navigation.flatMap(
      ({ items: categoryItems }) => categoryItems,
    );
    const integrations = items.find(({ title }) => title === "Integrations");
    const imports = items.find(({ title }) => title === "Imports");

    if (!integrations || !imports) {
      throw new Error("Expected integration and import settings items");
    }

    expect(
      isSettingsItemActive("/acme/settings/integrations/slack", integrations),
    ).toBe(true);
    expect(
      isSettingsItemActive(
        "/acme/settings/workspace/integrations/github",
        integrations,
      ),
    ).toBe(true);
    expect(
      isSettingsItemActive("/acme/settings/workspace/imports/run-1", imports),
    ).toBe(true);
  });
});
