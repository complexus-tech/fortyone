/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

const readSource = (path: string) =>
  readFileSync(join(process.cwd(), path), "utf8");

describe("TeamsMenu", () => {
  it("uses the shared nested menu pattern for overflow teams", () => {
    const source = readSource("src/components/ui/teams-menu.tsx");

    expect(source).toContain("<DropdownMenu.SubMenu");
    expect(source).toContain("<DropdownMenu.SubTrigger");
    expect(source).toContain("<DropdownMenu.SubItems");
    expect(source).toContain("Search your teams");
    expect(source).toContain("Your Teams");
    expect(source).not.toContain("useJoinedTeamsInfinite");
  });

  it("keeps joined-team membership and pin actions distinct", () => {
    const source = readSource("src/components/ui/teams-menu.tsx");

    expect(source).toMatch(/aria-label=\{`Leave \$\{team\.name\}`\}/);
    expect(source).toMatch(/aria-label=\{`Pin \$\{team\.name\}`\}/);
    expect(source).toContain('<Tooltip title="Pin">');
    expect(source).toContain('setTeam(team.id, "leave")');
    expect(source).toContain("onPinTeam(team.id)");
  });

  it("lists joinable teams directly without the old heading", () => {
    const source = readSource("src/components/ui/teams-menu.tsx");

    expect(source).toContain("usePublicTeamsInfinite");
    expect(source).toContain("Join team");
    expect(source).not.toContain("Join a team");
    expect(source.indexOf("Manage Teams")).toBeLessThan(
      source.lastIndexOf("<YourTeamsSubMenu"),
    );
  });

  it("renders one compact placeholder set while public teams load", () => {
    const source = readSource("src/components/ui/teams-menu.tsx");
    const initialLoaderUsages = source.match(
      /rows=\{INITIAL_TEAM_MENU_SKELETON_ROWS\}/g,
    );

    expect(initialLoaderUsages).toHaveLength(1);
  });
});
