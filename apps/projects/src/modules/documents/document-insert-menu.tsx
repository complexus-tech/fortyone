"use client";

import type { Editor } from "@tiptap/react";
import { FloatingMenu } from "@tiptap/react/menus";
import { PlusIcon } from "icons";
import { Button, Tooltip } from "ui";

export function DocumentInsertMenu({
  editor,
  scrollContainer,
}: {
  editor: Editor;
  scrollContainer: HTMLElement | null;
}) {
  return (
    <FloatingMenu
      appendTo={() => window.document.body}
      className="z-30"
      editor={editor}
      options={{
        strategy: "fixed",
        placement: "left",
        offset: 8,
        scrollTarget: scrollContainer ?? undefined,
        flip: false,
        shift: {
          boundary: scrollContainer ?? "clippingAncestors",
          padding: { top: 88, right: 8, bottom: 16, left: 8 },
        },
      }}
    >
      <Tooltip className="text-base" title="Insert a block">
        <Button
          aria-label="Insert a block"
          asIcon
          className="text-base"
          color="tertiary"
          onClick={() => {
            editor.chain().focus().insertContent("/").run();
          }}
          onMouseDown={(event) => {
            event.preventDefault();
          }}
          size="sm"
          variant="naked"
        >
          <PlusIcon strokeWidth={2} />
        </Button>
      </Tooltip>
    </FloatingMenu>
  );
}
