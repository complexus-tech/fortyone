/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

const readSource = (path: string) =>
  readFileSync(join(process.cwd(), path), "utf8");

describe("ChatMessage", () => {
  it("uses story-list surface colors for user prompts instead of inverse colors", () => {
    const source = readSource("src/components/ui/chat/chat-message.tsx");

    expect(source).toContain("bg-state-hover/80 dark:bg-surface-muted/95");
    expect(source).not.toContain("text-foreground-inverse");
    expect(source).not.toContain("bg-background-inverse rounded-tr-md");
  });

  it("does not force text colors in chat message markdown", () => {
    const source = readSource("src/styles/global.css");

    expect(source).not.toMatch(/\.chat-tables[\s\S]*text-foreground/);
  });

  it("lets the chat composer inherit theme text color", () => {
    const source = readSource("src/components/ui/chat/chat-input.tsx");

    expect(source).not.toContain("dark:text-white");
  });
});
