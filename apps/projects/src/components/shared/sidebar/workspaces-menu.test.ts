/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

const readSource = (path: string) =>
  readFileSync(join(process.cwd(), path), "utf8");

describe("WorkspacesMenu", () => {
  it("guards the current workspace menu item while workspace data is unavailable", () => {
    const source = readSource("src/components/shared/sidebar/workspaces-menu.tsx");

    expect(source).not.toContain("workspace!.id");
    expect(source).not.toContain("workspace!.slug");
    expect(source).toContain("disabled={!workspace}");
    expect(source).toContain("Loading workspace");
  });
});
