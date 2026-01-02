"use client";
import {
  Dispatch,
  SetStateAction,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react";
import "@tiptap/starter-kit";
import "@tiptap/extension-underline";
import "@tiptap/extension-task-item";
import "@tiptap/extension-task-list";
import "@tiptap/extension-link";
import "@tiptap/extension-heading";
import { Editor } from "@tiptap/react";
import { Flex } from "./flex";
import { Tooltip } from "./tooltip";
import { Button as ButtonBase, ButtonProps } from "./button";
import { Box } from "./box";
import { Text } from "./text";
import { cn } from "lib";
import {
  BoldIcon,
  CheckListIcon,
  CodeBlockIcon,
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

const Button = ({ active, className, ...props }: ButtonProps) => {
  return (
    <ButtonBase
      asIcon
      active={active}
      className={cn(className, "dark:hover:bg-gray/30 hover:bg-gray-300/20", {
        "dark:bg-gray/50 bg-gray-300/30": active,
      })}
      {...props}
    />
  );
};

export const BubbleMenu = ({
  editor,
  isLinkOpen,
  setIsLinkOpen,
}: {
  editor: Editor;
  isLinkOpen: boolean;
  setIsLinkOpen: Dispatch<SetStateAction<boolean>>;
}) => {
  const inputRef = useRef<HTMLInputElement>(null);
  const [url, setURL] = useState(editor.getAttributes("link").href);

  const setLink = useCallback(() => {
    // cancelled
    if (url === null) {
      setIsLinkOpen(false);
      return;
    }
    // empty
    if (url === "") {
      editor.chain().focus().extendMarkRange("link").unsetLink().run();
      setIsLinkOpen(false);
      return;
    }
    // update link
    editor.chain().focus().extendMarkRange("link").setLink({ href: url }).run();
    setIsLinkOpen(false);
  }, [editor, isLinkOpen]);

  useEffect(() => {
    inputRef.current && inputRef.current?.focus();
  });

  return (
    <Box>
      <Flex
        align="center"
        className={cn(
          "w-max rounded-xl border-[0.5px] border-border bg-surface-elevated/90 p-2 shadow-lg shadow-shadow backdrop-blur",
          {
            hidden: isLinkOpen,
          }
        )}
        gap={2}
      >
        <Tooltip title="Bold">
          <Button
            active={editor.isActive("bold")}
            color="tertiary"
            size="sm"
            variant="naked"
            onClick={() => editor.chain().focus().toggleBold().run()}
          >
            <BoldIcon />
          </Button>
        </Tooltip>
        <Tooltip title="Italic">
          <Button
            active={editor.isActive("italic")}
            color="tertiary"
            size="sm"
            variant="naked"
            onClick={() => editor.chain().focus().toggleItalic().run()}
          >
            <ItalicIcon />
          </Button>
        </Tooltip>
        <Tooltip title="Underline">
          <Button
            active={editor.isActive("underline")}
            color="tertiary"
            size="sm"
            variant="naked"
            onClick={() => editor.chain().focus().toggleUnderline().run()}
          >
            <UnderlineIcon />
          </Button>
        </Tooltip>
        <Tooltip title="Strikethrough">
          <Button
            active={editor.isActive("strike")}
            color="tertiary"
            size="sm"
            variant="naked"
            onClick={() => editor.chain().focus().toggleStrike().run()}
          >
            <StrikeThroughIcon />
          </Button>
        </Tooltip>
        <Tooltip title="Link">
          <Button
            active={editor.isActive("link")}
            color="tertiary"
            size="sm"
            variant="naked"
            onClick={() => setIsLinkOpen(true)}
          >
            <LinkIcon />
          </Button>
        </Tooltip>
        <span className="opacity-30">|</span>
        <Tooltip title="Heading 1">
          <Button
            active={editor.isActive("heading", { level: 1 })}
            color="tertiary"
            size="sm"
            variant="naked"
            onClick={() =>
              editor.chain().focus().toggleHeading({ level: 1 }).run()
            }
          >
            <Text as="span" color="muted">
              H1
            </Text>
          </Button>
        </Tooltip>
        <Tooltip title="Heading 2">
          <Button
            active={editor.isActive("heading", { level: 2 })}
            color="tertiary"
            size="sm"
            variant="naked"
            onClick={() =>
              editor.chain().focus().toggleHeading({ level: 2 }).run()
            }
          >
            <Text as="span" color="muted">
              H2
            </Text>
          </Button>
        </Tooltip>
        <Tooltip title="Heading 3">
          <Button
            active={editor.isActive("heading", { level: 3 })}
            color="tertiary"
            size="sm"
            variant="naked"
            onClick={() =>
              editor.chain().focus().toggleHeading({ level: 3 }).run()
            }
          >
            <Text as="span" color="muted">
              H3
            </Text>
          </Button>
        </Tooltip>
        <Tooltip title="Heading 4">
          <Button
            active={editor.isActive("heading", { level: 4 })}
            color="tertiary"
            size="sm"
            variant="naked"
            onClick={() =>
              editor.chain().focus().toggleHeading({ level: 4 }).run()
            }
          >
            <Text as="span" color="muted">
              H4
            </Text>
          </Button>
        </Tooltip>
        <Tooltip title="Paragraph">
          <Button
            active={editor.isActive("paragraph")}
            color="tertiary"
            size="sm"
            variant="naked"
            onClick={() => editor.chain().focus().setParagraph().run()}
          >
            <Text as="span" color="muted">
              P
            </Text>
          </Button>
        </Tooltip>
        <span className="opacity-30">|</span>
        <Tooltip title="Ordered list">
          <Button
            active={editor.isActive("orderedList")}
            color="tertiary"
            size="sm"
            variant="naked"
            onClick={() => editor.chain().focus().toggleOrderedList().run()}
          >
            <OrderedListIcon />
          </Button>
        </Tooltip>
        <Tooltip title="Unordered list">
          <Button
            active={editor.isActive("bulletList")}
            color="tertiary"
            size="sm"
            variant="naked"
            onClick={() => editor.chain().focus().toggleBulletList().run()}
          >
            <UnorderedListIcon />
          </Button>
        </Tooltip>
        <Tooltip title="Check list">
          <Button
            active={editor.isActive("taskList")}
            color="tertiary"
            size="sm"
            variant="naked"
            onClick={() => editor.chain().focus().toggleTaskList().run()}
          >
            <CheckListIcon />
          </Button>
        </Tooltip>
        <Tooltip title="Code">
          <Button
            active={editor.isActive("code")}
            color="tertiary"
            size="sm"
            variant="naked"
            onClick={() => editor.chain().focus().toggleCode().run()}
          >
            <CodeIcon />
          </Button>
        </Tooltip>
        <span className="opacity-30">|</span>
        <Tooltip title="Quote">
          <Button
            active={editor.isActive("blockquote")}
            color="tertiary"
            size="sm"
            variant="naked"
            onClick={() => editor.chain().focus().toggleBlockquote().run()}
          >
            <QuoteIcon />
          </Button>
        </Tooltip>
        <Tooltip title="Code block">
          <Button
            active={editor.isActive("codeBlock")}
            color="tertiary"
            size="sm"
            variant="naked"
            onClick={() => editor.chain().focus().toggleCodeBlock().run()}
          >
            <CodeBlockIcon />
          </Button>
        </Tooltip>
      </Flex>

      <Flex
        align="center"
        className={cn(
          "w-max rounded-[0.6rem] border border-gray-100 bg-white/80 px-2 py-[0.2rem] backdrop-blur dark:border-dark-100 dark:bg-dark-200/80",
          {
            hidden: !isLinkOpen,
          }
        )}
        gap={1}
      >
        <form
          className="flex items-center gap-1"
          onSubmit={(e) => {
            e.preventDefault();
            setLink();
          }}
        >
          <input
            type="url"
            ref={inputRef}
            placeholder="Enter the URL..."
            onChange={(e) => {
              setURL(e.target.value);
            }}
            className="outline-none w-60 bg-transparent text-sm p-2 placeholder:text-gray-250 dark:placeholder:text-gray-200"
            defaultValue={editor.getAttributes("link").href || ""}
          />
          <span className="opacity-30">|</span>
          <Button
            active
            className="border-gray-200"
            size="xs"
            type="button"
            color="tertiary"
            leftIcon={<DeleteIcon className="h-4 relative left-[0.12rem]" />}
            onClick={() => {
              editor.chain().focus().unsetLink().run();
              setIsLinkOpen(false);
            }}
          >
            <span className="sr-only">Delete</span>
          </Button>
        </form>
      </Flex>
    </Box>
  );
};
