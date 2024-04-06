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
import { Editor } from "@tiptap/react";
import { Flex } from "../Flex/Flex";
import { Tooltip } from "../Tooltip/Tooltip";
import { Button } from "../Button/Button";
import { Box } from "../Box/Box";
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
      {/* <Flex
        align="center"
        className={cn(
          "w-max rounded-lg border border-gray-100 bg-white/80 px-2 py-1 backdrop-blur dark:border-dark-100 dark:bg-dark-200/80",
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
            <BoldIcon className="h-5 w-auto" />
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
            <ItalicIcon className="h-5 w-auto" />
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
            <UnderlineIcon className="h-5 w-auto" />
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
            <StrikeThroughIcon className="h-5 w-auto" />
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
            <LinkIcon className="h-5 w-auto" />
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
            <OrderedListIcon className="h-5 w-auto" />
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
            <UnorderedListIcon className="h-5 w-auto" />
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
            <CheckListIcon className="h-5 w-auto" />
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
            <CodeIcon className="h-5 w-auto" />
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
            <QuoteIcon className="h-5 w-auto" />
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
            <CodeBlockIcon className="h-5 w-auto" />
          </Button>
        </Tooltip>
      </Flex> */}

      <Flex
        align="center"
        className={cn(
          "w-max rounded-lg border border-gray-100 bg-white/80 px-2 py-[0.2rem] backdrop-blur dark:border-dark-100 dark:bg-dark-200/80",
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
            leftIcon={
              <DeleteIcon className="h-4 w-auto relative left-[0.12rem]" />
            }
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
