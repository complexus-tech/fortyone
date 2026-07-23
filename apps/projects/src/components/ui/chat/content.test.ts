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
    expect(source).not.toContain("isLiveVoiceVisible");
    expect(source).toContain("liveVoiceDisabled={needsUpgrade}");
  });

  it("opens Maya in a fixed popup without resizing the workspace", () => {
    const chatSource = readSource("src/components/ui/chat/index.tsx");
    const layoutSource = readSource(
      "src/components/ui/chat/workspace-chat-layout.tsx",
    );
    const popupSource = readSource("src/components/ui/chat/rail.tsx");

    expect(chatSource).toContain("<ChatButton");
    expect(chatSource).toContain("<ChatRail />");
    expect(layoutSource).not.toContain("ResizablePanel");
    expect(popupSource).toContain("w-[min(420px,calc(100vw-36px))]");
    expect(popupSource).toContain("h-[min(760px,calc(100dvh-64px))]");
    expect(popupSource).toContain("border-border");
    expect(popupSource).toContain("<ChatContent isPopup />");
  });

  it("uses compact prompt rows in the popup empty state", () => {
    const source = readSource("src/components/ui/chat/suggested-prompts.tsx");

    expect(source).toContain("Hi, {name}! Ask me anything!");
    expect(source).toContain("What should I focus on today?");
    expect(source).toContain("min-h-[52px]");
    expect(source).toContain("border-b");
  });
});
