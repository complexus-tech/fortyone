/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

const readSource = (path: string) =>
  readFileSync(join(process.cwd(), path), "utf8");

describe("Maya quota headers", () => {
  it("hides the full-page quota affordance for internal users", () => {
    const source = readSource("src/modules/maya/components/header.tsx");

    expect(source).toContain("useSession");
    expect(source).toContain("session?.user.isInternal");
    expect(source).toContain('tier !== "enterprise" && !isInternalUser');
  });

  it("hides the docked quota affordance for internal users", () => {
    const source = readSource("src/components/ui/chat/chat-header.tsx");

    expect(source).toContain("useSession");
    expect(source).toContain("session?.user.isInternal");
    expect(source).toContain('tier !== "enterprise" && !isInternalUser');
  });
});
