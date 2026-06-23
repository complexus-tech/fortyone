/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

const readSource = (path: string) =>
  readFileSync(join(process.cwd(), path), "utf8");

describe("ChatContent", () => {
  it("keeps AI message limits scoped to non-internal users", () => {
    const source = readSource("src/components/ui/chat/content.tsx");

    expect(source).toContain("useSession");
    expect(source).toContain("session?.user.isInternal");
    expect(source).toContain("shouldShowMayaMessageLimit");
    expect(source).toContain("isLiveVoiceVisible={isInternalUser}");
    expect(source).toContain("liveVoiceDisabled={needsUpgrade}");
  });
});
