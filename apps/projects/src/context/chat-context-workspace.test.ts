/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

const workspaceLayoutSource = readFileSync(
  join(process.cwd(), "src/app/[workspaceSlug]/layout.tsx"),
  "utf8",
);

describe("workspace Maya chat lifecycle", () => {
  it("remounts the chat provider when the active workspace changes", () => {
    expect(workspaceLayoutSource).toContain(
      "<ChatProvider key={workspaceSlug}>",
    );
  });
});
