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
    expect(popupSource).toContain("w-[min(460px,calc(100vw-36px))]");
    expect(popupSource).toContain("h-[min(760px,calc(100dvh-64px))]");
    expect(popupSource).toContain("border-border/80 dark:border-border");
    expect(popupSource).toContain("border-[0.5px]");
    expect(popupSource).toContain("rounded-2xl");
    expect(popupSource).toContain("bg-white/80");
    expect(popupSource).toContain("dark:bg-surface-elevated/90");
    expect(popupSource).toContain("<ChatContent isPopup />");
  });

  it("keeps the popup header open to content with restrained top spacing", () => {
    const contentSource = readSource("src/components/ui/chat/content.tsx");
    const messagesSource = readSource(
      "src/components/ui/chat/chat-messages.tsx",
    );
    const suggestionsSource = readSource(
      "src/components/ui/chat/suggested-prompts.tsx",
    );

    expect(contentSource).toContain('"h-19 px-5 py-3": isPopup');
    expect(contentSource).not.toContain("border-b-[0.5px]");
    expect(messagesSource).toContain('"px-[18px] pt-4 pb-5": isPopup');
    expect(suggestionsSource).toContain('className="px-[18px] pt-8 pb-[18px]"');
  });

  it("uses a subtle prompt border inside the popup", () => {
    const source = readSource("src/components/ui/chat/chat-input.tsx");

    expect(source).toContain(
      "border-[0.5px] border-black/[0.07] bg-black/[0.035]",
    );
    expect(source).toContain("dark:border-white/[0.07]");
  });

  it("keeps the popup avatar perfectly circular", () => {
    const source = readSource("src/components/ui/chat/chat-header.tsx");

    expect(source).toContain('className="h-10 w-10"');
    expect(source).toContain('rounded="full"');
    expect(source).not.toContain('className="size-10 rounded-full"');
  });

  it("matches Kanban card typography in the popup prompt rows", () => {
    const source = readSource("src/components/ui/chat/suggested-prompts.tsx");

    expect(source).toContain("Hi, {name}! Ask me anything!");
    expect(source).toContain("What should I focus on today?");
    expect(source).toContain("min-h-[52px]");
    expect(source).toContain("text-foreground");
    expect(source).toContain("text-[1.1rem] leading-[1.4rem]");
    expect(source).toContain("border-b");
  });

  it("does not collapse workspace controls while the popup is open", () => {
    const objectiveHeader = readSource(
      "src/modules/objectives/stories/header.tsx",
    );
    const objectiveOverview = readSource(
      "src/modules/objectives/stories/overview/index.tsx",
    );
    const sprintHeader = readSource("src/modules/sprints/stories/header.tsx");
    const sprintStories = readSource(
      "src/modules/sprints/stories/list-stories.tsx",
    );
    const storyPage = readSource("src/modules/story/index.tsx");

    expect(objectiveHeader).not.toContain("isChatOpen");
    expect(objectiveOverview).not.toContain("isChatOpen");
    expect(sprintHeader).not.toContain("isChatOpen");
    expect(sprintStories).not.toContain("isChatOpen");
    expect(storyPage).not.toContain("isChatOpen");
  });
});
