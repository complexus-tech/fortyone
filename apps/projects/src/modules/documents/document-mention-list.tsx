import type { SuggestionProps } from "@tiptap/suggestion";
import { forwardRef, useImperativeHandle, useState } from "react";
import { Avatar } from "ui";
import { ObjectiveIcon, StoryIcon } from "icons";
import { cn } from "lib";
import { MayaAvatar } from "@/components/ui/maya-avatar";

export type DocumentMentionItem = {
  id: string;
  label: string;
  kind: "user" | "agent" | "story" | "objective" | "message";
  avatar?: string | null;
  displayId?: string;
};

export type DocumentMentionListRef = {
  onKeyDown: (event: KeyboardEvent) => boolean;
};

export const DocumentMentionList = forwardRef<
  DocumentMentionListRef,
  Pick<
    SuggestionProps<DocumentMentionItem, DocumentMentionItem>,
    "items" | "command" | "query"
  >
>(({ items, command, query }, ref) => {
  const [selection, setSelection] = useState({ query, index: 0 });
  const selectable = items.filter((item) => item.kind !== "message");
  const activeIndex =
    selection.query === query
      ? Math.min(selection.index, Math.max(0, selectable.length - 1))
      : 0;
  const select = (index: number) => {
    const item = selectable.at(index);
    if (item) command(item);
  };
  useImperativeHandle(ref, () => ({
    onKeyDown: (event) => {
      if (event.key === "Enter") {
        select(activeIndex);
        return true;
      }
      if (event.key !== "ArrowUp" && event.key !== "ArrowDown") return false;
      if (selectable.length)
        setSelection({
          query,
          index:
            (activeIndex +
              (event.key === "ArrowUp" ? -1 : 1) +
              selectable.length) %
            selectable.length,
        });
      return true;
    },
  }));

  return (
    <div
      aria-label="Mention or link"
      className="border-border bg-surface-elevated pointer-events-auto max-h-80 w-80 max-w-[calc(100vw-2rem)] overflow-y-auto rounded-xl border p-1.5 shadow-lg"
      data-document-mention-picker
      role="listbox"
    >
      {selectable.length ? (
        selectable.map((item, index) => (
          <button
            aria-selected={index === activeIndex}
            className={cn(
              "text-foreground flex w-full items-center gap-2.5 rounded-lg px-2.5 py-2 text-left text-base outline-none",
              index === activeIndex && "bg-state-hover",
            )}
            key={`${item.kind}:${item.id}`}
            onClick={() => {
              select(index);
            }}
            onMouseDown={(event) => {
              event.preventDefault();
            }}
            onMouseEnter={() => {
              setSelection({ query, index });
            }}
            role="option"
            type="button"
          >
            {item.kind === "agent" ? (
              <MayaAvatar size="sm" src={item.avatar} />
            ) : null}
            {item.kind === "user" ? (
              <Avatar name={item.label} size="sm" src={item.avatar} />
            ) : null}
            {item.kind === "story" ? (
              <StoryIcon className="text-primary size-6 shrink-0 p-0.5" />
            ) : null}
            {item.kind === "objective" ? (
              <ObjectiveIcon className="text-secondary size-6 shrink-0 p-0.5" />
            ) : null}
            <span className="min-w-0 flex-1 truncate font-medium">
              {item.displayId ? (
                <span className="text-text-muted mr-2 font-normal">
                  {item.displayId}
                </span>
              ) : null}
              {item.label}
              {item.kind === "agent" ? (
                <span className="text-text-muted font-normal"> (agent)</span>
              ) : null}
            </span>
          </button>
        ))
      ) : (
        <p className="text-text-muted px-3 py-3 text-base" role="status">
          {items.find((item) => item.kind === "message")?.label ??
            "No matches found"}
        </p>
      )}
    </div>
  );
});
DocumentMentionList.displayName = "DocumentMentionList";
