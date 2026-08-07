/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { Editor } from "@tiptap/core";
import {
  getSlashCommandItems,
  hasVisibleSlashCommandAnchor,
  shouldShowSlashCommand,
} from "./slash-command";

const editor = {} as Editor;

describe("slash command items", () => {
  it("exposes the complete rich-text command set", () => {
    const itemIds = getSlashCommandItems(editor, () => undefined).map(
      ({ id }) => id,
    );

    expect(itemIds).toEqual([
      "paragraph",
      "heading-1",
      "heading-2",
      "heading-3",
      "quote",
      "bullet-list",
      "ordered-list",
      "task-list",
      "media",
      "table",
      "code-block",
      "divider",
    ]);
  });

  it("does not advertise media when a surface has no upload adapter", () => {
    const itemIds = getSlashCommandItems(editor, null).map(({ id }) => id);

    expect(itemIds).not.toContain("media");
    expect(itemIds).toContain("table");
  });

  it("only opens from the focused, mounted editor", () => {
    expect(
      shouldShowSlashCommand({ isDestroyed: false, isFocused: true }),
    ).toBe(true);
    expect(
      shouldShowSlashCommand({ isDestroyed: false, isFocused: false }),
    ).toBe(false);
    expect(shouldShowSlashCommand({ isDestroyed: true, isFocused: true })).toBe(
      false,
    );
  });

  it("rejects the zero-sized anchor produced by a hidden editor", () => {
    expect(
      hasVisibleSlashCommandAnchor({ height: 0, width: 0 } as DOMRect),
    ).toBe(false);
    expect(
      hasVisibleSlashCommandAnchor({ height: 20, width: 0 } as DOMRect),
    ).toBe(true);
  });
});
