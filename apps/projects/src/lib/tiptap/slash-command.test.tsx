/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { Editor } from "@tiptap/core";
import { fireEvent, render, screen } from "@testing-library/react";
import {
  getSlashCommandItems,
  hasVisibleSlashCommandAnchor,
  SlashCommandList,
  shouldShowSlashCommand,
} from "./slash-command";

jest.mock("ui", () => ({ Box: "div", Text: "span" }));
jest.mock("icons", () => ({
  CheckListIcon: "span",
  CodeXmlIcon: "span",
  ImageIcon: "span",
  LayoutTable01Icon: "span",
  ListIcon: "span",
  MinusIcon: "span",
  OrderedListIcon: "span",
  QuoteIcon: "span",
  UnorderedListIcon: "span",
}));
jest.mock("lib", () => ({
  cn: (...values: (Record<string, boolean> | string | undefined)[]) =>
    values
      .flatMap((value) =>
        typeof value === "string"
          ? [value]
          : Object.entries(value ?? {})
              .filter(([, enabled]) => enabled)
              .map(([className]) => className),
      )
      .join(" "),
}));

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

  it("keeps editor focus until a clicked media command opens its picker", () => {
    const onMediaRequest = jest.fn();
    const items = getSlashCommandItems(editor, onMediaRequest).filter(
      ({ id }) => id === "media",
    );
    const command = jest.fn((item: (typeof items)[number]) => {
      item.command(editor);
    });

    render(<SlashCommandList command={command} items={items} query="" />);

    const mediaButton = screen.getByRole("menuitem", {
      name: "Insert media...",
    });
    const mouseDown = new MouseEvent("mousedown", {
      bubbles: true,
      cancelable: true,
    });
    mediaButton.dispatchEvent(mouseDown);
    fireEvent.click(mediaButton);

    expect(mouseDown.defaultPrevented).toBe(true);
    expect(command).toHaveBeenCalledWith(items[0]);
    expect(onMediaRequest).toHaveBeenCalledWith(editor);
  });
});
