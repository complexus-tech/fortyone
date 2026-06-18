/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

const readSource = (path: string) =>
  readFileSync(join(process.cwd(), path), "utf8");

describe("ChatContent", () => {
  it("only enables realtime voice for internal users within the AI limit", () => {
    const source = readSource("src/components/ui/chat/content.tsx");

    expect(source).toContain("useSession");
    expect(source).toContain("session?.user.isInternal");
    expect(source).toContain("isLiveVoiceVisible={isInternalUser}");
    expect(source).toContain("liveVoiceDisabled={needsUpgrade}");
  });
});
