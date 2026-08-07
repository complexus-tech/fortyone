"use client";

import type {
  Dispatch,
  MouseEvent as ReactMouseEvent,
  ReactNode,
  SetStateAction,
} from "react";
import { useCallback, useEffect, useRef } from "react";
import "@tiptap/starter-kit";
import "@tiptap/extension-underline";
import "@tiptap/extension-task-item";
import "@tiptap/extension-task-list";
import "@tiptap/extension-link";
import "@tiptap/extension-heading";
import { useEditorState, type Editor } from "@tiptap/react";
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
    className={cn(
      "hover:bg-state-active focus-visible:bg-state-active dark:hover:bg-state-active dark:focus-visible:bg-state-active",
      className,
      {
        "bg-state-active text-foreground dark:bg-state-active dark:text-foreground [&_svg]:!text-foreground dark:[&_svg]:!text-foreground":
          active,
      },
    )}
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

export type BubbleMenuCreateAction = {
  id: string;
  icon?: ReactNode;
  label: string;
  onSelect: (selectedText: string) => void;
};

export type BubbleMenuPanel = "text" | "create" | "link" | null;

type ToolbarMenuProps = {
  activeMenu: BubbleMenuPanel;
  editor: Editor;
  setActiveMenu: Dispatch<SetStateAction<BubbleMenuPanel>>;
};

const getBubbleMenuEditorState = (editor: Editor) => ({
  blockquote: editor.isActive("blockquote"),
  bold: editor.isActive("bold"),
  bulletList: editor.isActive("bulletList"),
  code: editor.isActive("code"),
  headings: {
    1: editor.isActive("heading", { level: 1 }),
    2: editor.isActive("heading", { level: 2 }),
    3: editor.isActive("heading", { level: 3 }),
    4: editor.isActive("heading", { level: 4 }),
  },
  italic: editor.isActive("italic"),
  link: editor.isActive("link"),
  orderedList: editor.isActive("orderedList"),
  paragraph: editor.isActive("paragraph"),
  strike: editor.isActive("strike"),
  taskList: editor.isActive("taskList"),
  underline: editor.isActive("underline"),
});

type BubbleMenuEditorState = ReturnType<typeof getBubbleMenuEditorState>;

const preserveEditorSelection = (event: ReactMouseEvent) => {
  event.preventDefault();
};

const TextStyleMenu = ({
  activeMenu,
  editor,
  editorState,
  setActiveMenu,
}: ToolbarMenuProps & { editorState: BubbleMenuEditorState }) => (
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
        className="hover:bg-state-active focus-visible:bg-state-active h-8 gap-1 px-2"
        color="tertiary"
        onMouseDown={preserveEditorSelection}
        rightIcon={<ArrowDown2Icon className="h-3.5" />}
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
            ? editorState.headings[level]
            : editorState.paragraph;
          return (
            <Menu.Item
              className="hover:bg-state-active focus-visible:bg-state-active flex w-full items-center justify-between rounded-md px-2 py-2 text-left outline-none"
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

const getSelectedText = (editor: Editor) => {
  const { from, to } = editor.state.selection;
  return editor.state.doc.textBetween(from, to, "\n\n").trim();
};

const CreateMenu = ({
  actions,
  activeMenu,
  editor,
  setActiveMenu,
}: ToolbarMenuProps & {
  actions: readonly BubbleMenuCreateAction[];
}) => {
  const selectedTextRef = useRef("");

  if (actions.length === 0) return null;

  return (
    <Menu
      onOpenChange={(open) => {
        if (open) selectedTextRef.current = getSelectedText(editor);
        setActiveMenu((current) =>
          open ? "create" : current === "create" ? null : current,
        );
      }}
      open={activeMenu === "create"}
    >
      <Menu.Button>
        <ButtonBase
          aria-label="Create from selection"
          className={cn(
            "hover:bg-state-active focus-visible:bg-state-active h-8 gap-1 px-2",
            { "bg-state-active": activeMenu === "create" },
          )}
          color="tertiary"
          onMouseDown={preserveEditorSelection}
          rightIcon={<ArrowDown2Icon className="h-3.5" />}
          size="sm"
          type="button"
          variant="naked"
        >
          Create
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
          {actions.map((action) => (
            <Menu.Item
              className="hover:bg-state-active focus-visible:bg-state-active flex w-full items-center gap-2 rounded-md px-2 py-2 text-left outline-none"
              key={action.id}
              onSelect={() => {
                action.onSelect(selectedTextRef.current);
              }}
            >
              {action.icon}
              {action.label}
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
  createActions = [],
  editor,
  setActiveMenu,
}: {
  activeMenu: BubbleMenuPanel;
  createActions?: readonly BubbleMenuCreateAction[];
  editor: Editor;
  setActiveMenu: Dispatch<SetStateAction<BubbleMenuPanel>>;
}) => {
  const editorState = useEditorState({
    editor,
    selector: ({ editor: currentEditor }) =>
      getBubbleMenuEditorState(currentEditor),
  });
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
          editorState={editorState}
          setActiveMenu={setActiveMenu}
        />
        <span className="bg-border-strong mx-1 h-5 w-px" />
        <Tooltip title="Bold">
          <Button
            active={editorState.bold}
            aria-label="Bold"
            aria-pressed={editorState.bold}
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
            active={editorState.italic}
            aria-label="Italic"
            aria-pressed={editorState.italic}
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
            active={editorState.strike}
            aria-label="Strikethrough"
            aria-pressed={editorState.strike}
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
            active={editorState.underline}
            aria-label="Underline"
            aria-pressed={editorState.underline}
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
            active={editorState.link}
            aria-label="Link"
            aria-pressed={editorState.link}
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
            active={editorState.blockquote}
            aria-label="Quote"
            aria-pressed={editorState.blockquote}
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
            active={editorState.code}
            aria-label="Inline code"
            aria-pressed={editorState.code}
            color="tertiary"
            onClick={() => editor.chain().focus().toggleCode().run()}
            size="sm"
            variant="naked"
          >
            <CodeIcon />
          </Button>
        </Tooltip>
        <Tooltip title="Bulleted list">
          <Button
            active={editorState.bulletList}
            aria-label="Bulleted list"
            aria-pressed={editorState.bulletList}
            color="tertiary"
            onClick={() => editor.chain().focus().toggleBulletList().run()}
            size="sm"
            variant="naked"
          >
            <UnorderedListIcon strokeWidth={2} />
          </Button>
        </Tooltip>
        <Tooltip title="Numbered list">
          <Button
            active={editorState.orderedList}
            aria-label="Numbered list"
            aria-pressed={editorState.orderedList}
            color="tertiary"
            onClick={() => editor.chain().focus().toggleOrderedList().run()}
            size="sm"
            variant="naked"
          >
            <OrderedListIcon strokeWidth={2} />
          </Button>
        </Tooltip>
        <Tooltip title="Checklist">
          <Button
            active={editorState.taskList}
            aria-label="Checklist"
            aria-pressed={editorState.taskList}
            color="tertiary"
            onClick={() => editor.chain().focus().toggleTaskList().run()}
            size="sm"
            variant="naked"
          >
            <CheckListIcon strokeWidth={2} />
          </Button>
        </Tooltip>
        <CreateMenu
          actions={createActions}
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
