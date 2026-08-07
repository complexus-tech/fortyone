"use client";

import type {
  Dispatch,
  MouseEvent as ReactMouseEvent,
  SetStateAction,
} from "react";
import { useCallback, useEffect, useRef } from "react";
import "@tiptap/starter-kit";
import "@tiptap/extension-underline";
import "@tiptap/extension-task-item";
import "@tiptap/extension-task-list";
import "@tiptap/extension-link";
import "@tiptap/extension-heading";
import type { Editor } from "@tiptap/react";
import {
  BoldIcon,
  CheckIcon,
  CheckListIcon,
  ArrowDown2Icon,
  CodeIcon,
  DeleteIcon,
  ItalicIcon,
  LinkIcon,
  OrderedListIcon,
  QuoteIcon,
  StrikeThroughIcon,
  UnderlineIcon,
  UnorderedListIcon,
} from "icons";
import { cn } from "lib";
import { Box } from "./box";
import { Button as ButtonBase, type ButtonProps } from "./button";
import { Flex } from "./flex";
import { Menu } from "./menu";
import { Text } from "./text";
import { Tooltip } from "./tooltip";

const Button = ({ active, className, ...props }: ButtonProps) => (
  <ButtonBase
    active={active}
    asIcon
    className={cn("hover:bg-state-hover", className, {
      "bg-state-active": active,
    })}
    type="button"
    {...props}
  />
);

const textStyles = [
  { label: "Regular text", level: null },
  { label: "Heading 1", level: 1 },
  { label: "Heading 2", level: 2 },
  { label: "Heading 3", level: 3 },
  { label: "Heading 4", level: 4 },
] as const;

export type BubbleMenuPanel = "text" | "list" | "link" | null;

type ToolbarMenuProps = {
  activeMenu: BubbleMenuPanel;
  editor: Editor;
  setActiveMenu: Dispatch<SetStateAction<BubbleMenuPanel>>;
};

const preserveEditorSelection = (event: ReactMouseEvent) => {
  event.preventDefault();
};

const TextStyleMenu = ({
  activeMenu,
  editor,
  setActiveMenu,
}: ToolbarMenuProps) => (
  <Menu
    onOpenChange={(open) => {
      setActiveMenu((current) =>
        open ? "text" : current === "text" ? null : current,
      );
    }}
    open={activeMenu === "text"}
  >
    <Menu.Button>
      <ButtonBase
        aria-label="Text style"
        className="h-8 gap-1 px-2"
        color="tertiary"
        onMouseDown={preserveEditorSelection}
        rightIcon={<ArrowDown2Icon className="h-3" />}
        size="sm"
        type="button"
        variant="naked"
      >
        <span className="text-base font-medium">Aa</span>
      </ButtonBase>
    </Menu.Button>
    <Menu.Items
      align="start"
      className="border-border-strong bg-surface-elevated min-w-64 border-[0.5px] p-1.5 shadow-xl dark:bg-surface-elevated"
      onCloseAutoFocus={(event) => {
        event.preventDefault();
      }}
      onEscapeKeyDown={() => {
        editor.commands.focus();
      }}
      portal={false}
    >
      <Menu.Group className="px-0">
        {textStyles.map(({ label, level }) => {
          const active = level
            ? editor.isActive("heading", { level })
            : editor.isActive("paragraph");
          return (
            <Menu.Item
              className="hover:bg-state-hover focus-visible:bg-state-hover flex w-full items-center justify-between rounded-md px-2 py-2 text-left outline-none"
              key={label}
              onSelect={() => {
                if (level) {
                  editor.chain().focus().setHeading({ level }).run();
                } else {
                  editor.chain().focus().setParagraph().run();
                }
              }}
            >
              <Text
                as="span"
                className={cn({
                  "text-xl font-semibold": level === 1,
                  "text-lg font-semibold": level === 2,
                  "font-semibold": level === 3,
                  "text-sm font-semibold": level === 4,
                })}
              >
                {label}
              </Text>
              {active ? <CheckIcon className="h-4" /> : null}
            </Menu.Item>
          );
        })}
      </Menu.Group>
    </Menu.Items>
  </Menu>
);

const ListStyleMenu = ({
  activeMenu,
  editor,
  setActiveMenu,
}: ToolbarMenuProps) => {
  const listStyles = [
    {
      active: editor.isActive("bulletList"),
      icon: UnorderedListIcon,
      label: "Bulleted list",
      onSelect: () => editor.chain().focus().toggleBulletList().run(),
    },
    {
      active: editor.isActive("orderedList"),
      icon: OrderedListIcon,
      label: "Numbered list",
      onSelect: () => editor.chain().focus().toggleOrderedList().run(),
    },
    {
      active: editor.isActive("taskList"),
      icon: CheckListIcon,
      label: "Checklist",
      onSelect: () => editor.chain().focus().toggleTaskList().run(),
    },
  ];

  return (
    <Menu
      onOpenChange={(open) => {
        setActiveMenu((current) =>
          open ? "list" : current === "list" ? null : current,
        );
      }}
      open={activeMenu === "list"}
    >
      <Menu.Button>
        <ButtonBase
          aria-label="List style"
          className="h-8 gap-1 px-2"
          color="tertiary"
          onMouseDown={preserveEditorSelection}
          size="sm"
          type="button"
          variant="naked"
        >
          <UnorderedListIcon className="h-4" />
          <ArrowDown2Icon className="h-2.5" />
        </ButtonBase>
      </Menu.Button>
      <Menu.Items
        align="end"
        className="border-border-strong bg-surface-elevated min-w-52 border-[0.5px] p-1.5 shadow-xl dark:bg-surface-elevated"
        onCloseAutoFocus={(event) => {
          event.preventDefault();
        }}
        onEscapeKeyDown={() => {
          editor.commands.focus();
        }}
        portal={false}
      >
        <Menu.Group className="px-0">
          {listStyles.map(({ active, icon: ListIcon, label, onSelect }) => (
            <Menu.Item
              active={active}
              className={cn(
                "hover:bg-state-hover focus-visible:bg-state-hover flex w-full items-center gap-2 rounded-md px-2 py-2 text-left outline-none",
                { "bg-state-active": active },
              )}
              key={label}
              onSelect={() => {
                onSelect();
              }}
            >
              <ListIcon className="h-4" />
              {label}
            </Menu.Item>
          ))}
        </Menu.Group>
      </Menu.Items>
    </Menu>
  );
};

const getLinkHref = (editor: Editor): string => {
  const attributes = editor.getAttributes("link") as { href?: unknown };
  return typeof attributes.href === "string" ? attributes.href : "";
};

export const BubbleMenu = ({
  activeMenu,
  editor,
  setActiveMenu,
}: {
  activeMenu: BubbleMenuPanel;
  editor: Editor;
  setActiveMenu: Dispatch<SetStateAction<BubbleMenuPanel>>;
}) => {
  const inputRef = useRef<HTMLInputElement>(null);
  const isLinkOpen = activeMenu === "link";

  const setLink = useCallback(() => {
    const url = inputRef.current?.value.trim() ?? "";
    if (url === "") {
      editor.chain().focus().extendMarkRange("link").unsetLink().run();
      setActiveMenu(null);
      return;
    }
    editor.chain().focus().extendMarkRange("link").setLink({ href: url }).run();
    setActiveMenu(null);
  }, [editor, setActiveMenu]);

  useEffect(() => {
    if (!isLinkOpen || !inputRef.current) return;
    inputRef.current.value = getLinkHref(editor);
    inputRef.current.focus();
  }, [editor, isLinkOpen]);

  return (
    <Box>
      <Flex
        align="center"
        className={cn(
          "border-border-strong bg-surface-elevated w-max rounded-xl border-[0.5px] p-1.5 shadow-xl",
          { hidden: isLinkOpen },
        )}
        gap={1}
      >
        <TextStyleMenu
          activeMenu={activeMenu}
          editor={editor}
          setActiveMenu={setActiveMenu}
        />
        <span className="bg-border-strong mx-1 h-5 w-px" />
        <Tooltip title="Bold">
          <Button
            active={editor.isActive("bold")}
            color="tertiary"
            onClick={() => editor.chain().focus().toggleBold().run()}
            size="sm"
            variant="naked"
          >
            <BoldIcon />
          </Button>
        </Tooltip>
        <Tooltip title="Italic">
          <Button
            active={editor.isActive("italic")}
            color="tertiary"
            onClick={() => editor.chain().focus().toggleItalic().run()}
            size="sm"
            variant="naked"
          >
            <ItalicIcon />
          </Button>
        </Tooltip>
        <Tooltip title="Strikethrough">
          <Button
            active={editor.isActive("strike")}
            color="tertiary"
            onClick={() => editor.chain().focus().toggleStrike().run()}
            size="sm"
            variant="naked"
          >
            <StrikeThroughIcon />
          </Button>
        </Tooltip>
        <Tooltip title="Underline">
          <Button
            active={editor.isActive("underline")}
            color="tertiary"
            onClick={() => editor.chain().focus().toggleUnderline().run()}
            size="sm"
            variant="naked"
          >
            <UnderlineIcon />
          </Button>
        </Tooltip>
        <Tooltip title="Link">
          <Button
            active={editor.isActive("link")}
            color="tertiary"
            onClick={() => {
              setActiveMenu("link");
            }}
            size="sm"
            variant="naked"
          >
            <LinkIcon />
          </Button>
        </Tooltip>
        <Tooltip title="Quote">
          <Button
            active={editor.isActive("blockquote")}
            color="tertiary"
            onClick={() => editor.chain().focus().toggleBlockquote().run()}
            size="sm"
            variant="naked"
          >
            <QuoteIcon />
          </Button>
        </Tooltip>
        <Tooltip title="Inline code">
          <Button
            active={editor.isActive("code")}
            color="tertiary"
            onClick={() => editor.chain().focus().toggleCode().run()}
            size="sm"
            variant="naked"
          >
            <CodeIcon />
          </Button>
        </Tooltip>
        <ListStyleMenu
          activeMenu={activeMenu}
          editor={editor}
          setActiveMenu={setActiveMenu}
        />
      </Flex>

      <Flex
        align="center"
        className={cn(
          "border-border-strong bg-surface-elevated w-max rounded-xl border-[0.5px] px-2 py-1 shadow-xl",
          { hidden: !isLinkOpen },
        )}
        gap={1}
      >
        <Box className="flex items-center gap-1">
          <input
            aria-label="Link URL"
            className="placeholder:text-text-muted w-60 bg-transparent p-2 text-sm outline-none"
            defaultValue={getLinkHref(editor)}
            onKeyDown={(event) => {
              if (event.key === "Enter") setLink();
            }}
            placeholder="Enter the URL..."
            ref={inputRef}
            type="url"
          />
          <Button
            active
            color="tertiary"
            onClick={setLink}
            size="xs"
            type="button"
          >
            Save
          </Button>
          <span className="bg-border-strong mx-1 h-5 w-px" />
          <Button
            active
            className="border-border"
            color="tertiary"
            leftIcon={<DeleteIcon className="relative left-[0.12rem] h-4" />}
            onClick={() => {
              editor.chain().focus().unsetLink().run();
              setActiveMenu(null);
            }}
            size="xs"
            type="button"
          >
            <span className="sr-only">Delete</span>
          </Button>
        </Box>
      </Flex>
    </Box>
  );
};
