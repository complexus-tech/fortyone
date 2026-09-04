/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

const readSource = (path: string) =>
  readFileSync(join(process.cwd(), path), "utf8");

describe("MayaChat", () => {
  it("keeps AI message limits scoped to non-internal users", () => {
    const source = readSource("src/modules/maya/components/index.tsx");
    const availabilitySource = readSource(
      "src/modules/maya/hooks/use-maya-message-availability.ts",
    );

    expect(source).toContain("useMayaMessageAvailability");
    expect(availabilitySource).toContain("session?.user.isInternal");
    expect(availabilitySource).toContain("shouldShowMayaMessageLimit");
    expect(source).not.toContain("isLiveVoiceVisible");
    expect(source).toContain("liveVoiceDisabled={needsUpgrade}");
  });

  it("uses a wider empty-state stage without widening conversations", () => {
    const source = readSource("src/modules/maya/components/index.tsx");
    const messagesSource = readSource(
      "src/components/ui/chat/chat-messages.tsx",
    );

    expect(source).toContain(
      '"relative flex h-full min-h-0 flex-col overflow-hidden"',
    );
    expect(source).toContain(
      'className="mx-auto flex w-full max-w-3xl shrink-0 flex-col"',
    );
    expect(source).toContain("<ChatMessages\n              isOnPage");
    expect(source).toContain("isEmptyState = displayMessages.length === 0");
    expect(source).toContain("max-w-5xl");
    expect(source).toContain("max-w-4xl");
    expect(messagesSource).toContain(
      '"mx-auto w-full max-w-3xl px-6 pt-20 pb-5": isOnPage',
    );
    expect(messagesSource).toContain('"hide-scrollbar": !isOnPage');
    expect(source).not.toContain('<BodyContainer className="mx-auto');
  });

  it("keeps reusable suggestions keyboard operable", () => {
    const chatSource = readSource("src/modules/maya/components/index.tsx");
    const suggestionsSource = readSource(
      "src/components/ui/chat/suggested-prompts.tsx",
    );

    expect(chatSource).toContain("<SuggestedPrompts");
    expect(suggestionsSource).toContain("<button\n            className={cn(");
    expect(suggestionsSource).toContain('type="button"');
    expect(suggestionsSource).not.toContain("tabIndex={0}");
    expect(suggestionsSource).toContain('label: "Plan project"');
    expect(suggestionsSource).toContain('label: "Sprint summary"');
    expect(suggestionsSource).toContain('label: "Status report"');
    expect(suggestionsSource).toContain('label: "Active work"');
    expect(suggestionsSource).toContain("size-[1.1875rem]");
    expect(suggestionsSource).toContain("min-h-[6.25rem]");
  });

  it("keeps the empty-state treatment restrained and personalized", () => {
    const source = readSource("src/modules/maya/components/index.tsx");
    const composerStyles = readSource(
      "src/components/ui/chat/chat-input.module.css",
    );
    const pageStyles = readSource(
      "src/modules/maya/components/index.module.css",
    );
    const composerSource = readSource("src/components/ui/chat/chat-input.tsx");

    expect(source).toContain("useProfile");
    expect(source).toContain("Hi, {firstName}! Ask Maya anything.");
    expect(source.replace(/\s+/g, " ")).toContain(
      "Plan what&apos;s next, find what matters, or move work forward.",
    );
    expect(source).not.toContain("AiIcon");
    expect(composerStyles).toContain("padding: 6px");
    expect(composerStyles).toContain("padding: 3px");
    expect(composerStyles).toContain("min-height: 9rem");
    expect(composerStyles).toContain("border-radius: 2.75rem");
    expect(composerStyles).toContain("corner-shape: squircle");
    expect(composerStyles).not.toContain("box-shadow");
    expect(composerStyles).not.toContain("backdrop-filter");
    expect(composerSource).toContain('color="tertiary"');
    expect(composerSource).toContain("isHighlighted");
    expect(composerSource).toContain(
      'className="bg-state-hover dark:bg-state-hover gap-1"',
    );
    expect(composerSource).toContain(
      "[styles.dockedFrame]: isOnPage && !isEmptyState",
    );
    expect(composerSource).toContain(
      "[styles.dockedSurface]: isOnPage && !isEmptyState",
    );
    expect(composerSource).toContain('rounded="full"');
    expect(composerSource).not.toContain("Maya can make mistakes");
    expect(composerSource).toContain("min-h-[5.75rem]");
    expect(source).toContain("{composer}\n                <SuggestedPrompts");
    expect(pageStyles).toContain("var(--color-secondary)");
    expect(pageStyles).toContain("var(--color-info)");
    expect(pageStyles).toContain("var(--color-primary)");
  });
});
