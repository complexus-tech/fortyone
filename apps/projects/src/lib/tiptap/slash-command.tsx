"use client";

import type { ReactNode } from "react";
import {
  forwardRef,
  useEffect,
  useImperativeHandle,
  useRef,
  useState,
} from "react";
import { Extension, type Editor } from "@tiptap/core";
import { PluginKey } from "@tiptap/pm/state";
import { ReactRenderer } from "@tiptap/react";
import Suggestion, {
  exitSuggestion,
  type SuggestionKeyDownProps,
  type SuggestionProps,
} from "@tiptap/suggestion";
import tippy, { type Instance } from "tippy.js";
import { Box, Text } from "ui";
import {
  CheckListIcon,
  CodeXmlIcon,
  ImageIcon,
  LayoutTable01Icon,
  ListIcon,
  MinusIcon,
  OrderedListIcon,
  QuoteIcon,
  UnorderedListIcon,
} from "icons";
import { cn } from "lib";
import { getRichTextOverlayRoot } from "./rich-text-overlay";

type SlashCommandItem = {
  icon: ReactNode;
  id: string;
  label: string;
  group: "Text" | "Lists" | "Insert";
  command: (editor: Editor) => void;
};

type SlashCommandOptions = {
  onMediaRequest: ((editor: Editor) => void) | null;
};

type SlashCommandListProps = {
  command: (item: SlashCommandItem) => void;
  items: SlashCommandItem[];
  query: string;
};

export type SlashCommandListRef = {
  onKeyDown: (event: KeyboardEvent) => boolean;
};

const SLASH_COMMAND_PLUGIN_KEY = new PluginKey("slashCommand");

type ActiveSlashCommandMenu = {
  destroy: () => void;
  owner: symbol;
};

let activeSlashCommandMenu: ActiveSlashCommandMenu | null = null;

export const shouldShowSlashCommand = (
  editor: Pick<Editor, "isDestroyed" | "isFocused">,
) => !editor.isDestroyed && editor.isFocused;

export const hasVisibleSlashCommandAnchor = (
  clientRect: DOMRect | null | undefined,
): clientRect is DOMRect =>
  Boolean(clientRect && (clientRect.width !== 0 || clientRect.height !== 0));

export const SlashCommandList = forwardRef<
  SlashCommandListRef,
  SlashCommandListProps
>(({ command, items, query }, ref) => {
  const [selection, setSelection] = useState({ index: 0, query });
  const selectedItemRef = useRef<HTMLButtonElement | null>(null);
  const selectedIndex = selection.query === query ? selection.index : 0;

  useEffect(() => {
    selectedItemRef.current?.scrollIntoView({ block: "nearest" });
  }, [items.length, query, selectedIndex]);

  const select = (index: number) => {
    if (items.length === 0) return;
    command(items[index % items.length]);
  };

  useImperativeHandle(ref, () => ({
    onKeyDown: (event) => {
      if (event.key === "ArrowUp") {
        setSelection({
          index:
            items.length === 0
              ? 0
              : (selectedIndex + items.length - 1) % items.length,
          query,
        });
        return true;
      }
      if (event.key === "ArrowDown") {
        setSelection({
          index: items.length === 0 ? 0 : (selectedIndex + 1) % items.length,
          query,
        });
        return true;
      }
      if (event.key === "Enter") {
        select(selectedIndex);
        return true;
      }
      return false;
    },
  }));

  if (items.length === 0) {
    return (
      <Box className="not-prose border-border/70 bg-surface-elevated shadow-shadow dark:bg-surface-elevated/80 w-max max-w-[calc(100vw-1.5rem)] min-w-56 rounded-xl border-[0.5px] px-3 py-4 text-[13px] font-medium shadow-xl backdrop-blur-md 2xl:text-[13.5px]">
        <Text color="muted">No matching commands</Text>
      </Box>
    );
  }

  let previousGroup: SlashCommandItem["group"] | null = null;
  return (
    <Box
      className="not-prose border-border/70 bg-surface-elevated shadow-shadow dark:bg-surface-elevated/80 max-h-[min(26rem,calc(100dvh-1.5rem))] w-max max-w-[calc(100vw-1.5rem)] min-w-56 overflow-y-auto rounded-xl border-[0.5px] py-2 text-[13px] font-medium shadow-xl backdrop-blur-md 2xl:text-[13.5px]"
      data-slash-command-menu=""
      role="menu"
    >
      {items.map((item, index) => {
        const showDivider =
          previousGroup !== null && item.group !== previousGroup;
        previousGroup = item.group;
        return (
          <Box key={item.id}>
            {showDivider ? (
              <Box className="border-border dark:border-border-strong/80 my-1.5 border-b-[0.5px]" />
            ) : null}
            <Box className="px-1.5">
              <button
                className={cn(
                  "hover:bg-accent focus-visible:bg-accent flex w-full cursor-pointer items-center gap-1.5 rounded-md px-2 py-1.5 text-left outline-none select-none",
                  { "bg-accent": selectedIndex === index },
                )}
                onClick={() => {
                  select(index);
                }}
                onMouseDown={(event) => {
                  event.preventDefault();
                }}
                onMouseEnter={() => {
                  setSelection({ index, query });
                }}
                ref={selectedIndex === index ? selectedItemRef : undefined}
                role="menuitem"
                type="button"
              >
                <span className="text-text-muted flex size-5 shrink-0 items-center justify-center">
                  {item.icon}
                </span>
                <Text as="span" className="min-w-0 truncate">
                  {item.label}
                </Text>
              </button>
            </Box>
          </Box>
        );
      })}
    </Box>
  );
});

SlashCommandList.displayName = "SlashCommandList";

export const getSlashCommandItems = (
  editor: Editor,
  onMediaRequest: SlashCommandOptions["onMediaRequest"],
): SlashCommandItem[] => [
  {
    id: "paragraph",
    label: "Regular text",
    icon: <ListIcon className="size-5" />,
    group: "Text",
    command: () => editor.chain().focus().setParagraph().run(),
  },
  {
    id: "heading-1",
    label: "Heading 1",
    icon: (
      <Text as="span" fontWeight="bold">
        H1
      </Text>
    ),
    group: "Text",
    command: () => editor.chain().focus().setHeading({ level: 1 }).run(),
  },
  {
    id: "heading-2",
    label: "Heading 2",
    icon: (
      <Text as="span" fontWeight="bold">
        H2
      </Text>
    ),
    group: "Text",
    command: () => editor.chain().focus().setHeading({ level: 2 }).run(),
  },
  {
    id: "heading-3",
    label: "Heading 3",
    icon: (
      <Text as="span" fontWeight="bold">
        H3
      </Text>
    ),
    group: "Text",
    command: () => editor.chain().focus().setHeading({ level: 3 }).run(),
  },
  {
    id: "quote",
    label: "Quote",
    icon: <QuoteIcon className="size-5" />,
    group: "Text",
    command: () => editor.chain().focus().setBlockquote().run(),
  },
  {
    id: "bullet-list",
    label: "Bulleted list",
    icon: <UnorderedListIcon className="size-5" />,
    group: "Lists",
    command: () => editor.chain().focus().toggleBulletList().run(),
  },
  {
    id: "ordered-list",
    label: "Numbered list",
    icon: <OrderedListIcon className="size-5" />,
    group: "Lists",
    command: () => editor.chain().focus().toggleOrderedList().run(),
  },
  {
    id: "task-list",
    label: "Checklist",
    icon: <CheckListIcon className="size-5" />,
    group: "Lists",
    command: () => editor.chain().focus().toggleTaskList().run(),
  },
  ...(onMediaRequest
    ? [
        {
          id: "media",
          label: "Insert media...",
          icon: <ImageIcon className="size-5" />,
          group: "Insert" as const,
          command: (currentEditor: Editor) => {
            onMediaRequest(currentEditor);
          },
        },
      ]
    : []),
  {
    id: "table",
    label: "Table",
    icon: <LayoutTable01Icon className="size-5" />,
    group: "Insert",
    command: () =>
      editor
        .chain()
        .focus()
        .insertTable({ rows: 3, cols: 3, withHeaderRow: true })
        .run(),
  },
  {
    id: "code-block",
    label: "Code block",
    icon: <CodeXmlIcon className="size-5" />,
    group: "Insert",
    command: () => editor.chain().focus().setCodeBlock().run(),
  },
  {
    id: "divider",
    label: "Divider",
    icon: <MinusIcon className="size-5" />,
    group: "Insert",
    command: () => editor.chain().focus().setHorizontalRule().run(),
  },
];

const renderSlashCommand = () => {
  const owner = Symbol("slash-command-menu");
  let component: ReactRenderer<SlashCommandListRef> | null = null;
  let popup: Instance | null = null;
  let referenceClientRect: DOMRect | null = null;

  const destroy = () => {
    const popupToDestroy = popup;
    const componentToDestroy = component;
    popup = null;
    component = null;
    referenceClientRect = null;
    if (activeSlashCommandMenu?.owner === owner) {
      activeSlashCommandMenu = null;
    }
    popupToDestroy?.destroy();
    componentToDestroy?.destroy();
  };

  return {
    onStart: (props: SuggestionProps<SlashCommandItem, SlashCommandItem>) => {
      destroy();
      if (!shouldShowSlashCommand(props.editor)) return;

      const nextReferenceClientRect = props.clientRect?.();
      if (!hasVisibleSlashCommandAnchor(nextReferenceClientRect)) {
        return;
      }
      activeSlashCommandMenu?.destroy();
      referenceClientRect = nextReferenceClientRect;

      const nextComponent = new ReactRenderer(SlashCommandList, {
        props: props as unknown as Record<string, unknown>,
        editor: props.editor,
      });
      component = nextComponent;
      const overlayRoot = getRichTextOverlayRoot(props.editor);
      popup = tippy(document.body, {
        appendTo: () => overlayRoot,
        content: nextComponent.element,
        getReferenceClientRect: () =>
          referenceClientRect ?? nextReferenceClientRect,
        interactive: true,
        placement: "bottom-start",
        showOnCreate: true,
        trigger: "manual",
      });
      activeSlashCommandMenu = { destroy, owner };
    },
    onUpdate: (props: SuggestionProps<SlashCommandItem, SlashCommandItem>) => {
      component?.updateProps(props as unknown as Record<string, unknown>);
      const nextReferenceClientRect = props.clientRect?.();
      if (
        !shouldShowSlashCommand(props.editor) ||
        !hasVisibleSlashCommandAnchor(nextReferenceClientRect)
      ) {
        destroy();
        return;
      }
      referenceClientRect = nextReferenceClientRect;
      popup?.setProps({
        getReferenceClientRect: () =>
          referenceClientRect ?? nextReferenceClientRect,
      });
    },
    onKeyDown: (props: SuggestionKeyDownProps) => {
      if (props.event.key === "Escape") {
        exitSuggestion(props.view, SLASH_COMMAND_PLUGIN_KEY);
        return true;
      }
      return component?.ref?.onKeyDown(props.event) ?? false;
    },
    onExit: destroy,
  };
};

export const SlashCommand = Extension.create<SlashCommandOptions>({
  name: "slashCommand",
  addOptions() {
    return {
      onMediaRequest: null,
    };
  },
  addProseMirrorPlugins() {
    return [
      Suggestion({
        editor: this.editor,
        pluginKey: SLASH_COMMAND_PLUGIN_KEY,
        char: "/",
        allowSpaces: true,
        startOfLine: true,
        allow: ({ editor }) => shouldShowSlashCommand(editor),
        items: ({ query }) => {
          const items = getSlashCommandItems(
            this.editor,
            this.options.onMediaRequest,
          );
          const normalizedQuery = query.trim().toLowerCase();
          return normalizedQuery
            ? items.filter((item) =>
                item.label.toLowerCase().includes(normalizedQuery),
              )
            : items;
        },
        command: ({ editor, range, props }) => {
          editor.chain().focus().deleteRange(range).run();
          props.command(editor);
        },
        render: renderSlashCommand,
      }),
    ];
  },
});
