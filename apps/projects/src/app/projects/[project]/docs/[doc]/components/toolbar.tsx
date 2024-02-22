import { Button, Container, Flex, Tooltip, Menu } from "ui";
import type { Editor } from "@tiptap/react";
import { ToggleNode } from "./node";
import {
  BoldIcon,
  CheckListIcon,
  CodeBlockIcon,
  CodeIcon,
  DeleteIcon,
  DuplicateIcon,
  EditIcon,
  FileLockedIcon,
  ItalicIcon,
  LinkIcon,
  MoreVerticalIcon,
  OrderedListIcon,
  PrintIcon,
  QuoteIcon,
  RedoIcon,
  StrikeThroughIcon,
  UnderlineIcon,
  UndoIcon,
  UnorderedListIcon,
} from "icons";

export const Toolbar = ({ editor }: { editor: Editor | null }) => {
  return (
    <Container className="sticky top-0 flex items-center justify-between border-b border-gray-50 bg-white/80  backdrop-blur dark:border-dark-200 dark:bg-dark-300/80">
      <Flex align="center" className="py-1.5" gap={2}>
        <Tooltip hidden={!editor?.can().undo()} title="Undo">
          <Button
            color="tertiary"
            disabled={!editor?.can().undo()}
            onClick={() => editor?.commands.undo()}
            size="sm"
            variant="naked"
          >
            <UndoIcon className="h-5 w-auto" />
          </Button>
        </Tooltip>
        <Tooltip hidden={!editor?.can().redo()} title="Redo">
          <Button
            active={editor?.isActive("code")}
            color="tertiary"
            disabled={!editor?.can().redo()}
            onClick={() => editor?.commands.redo()}
            size="sm"
            variant="naked"
          >
            <RedoIcon className="h-5 w-auto" />
          </Button>
        </Tooltip>
        <ToggleNode editor={editor} />
        <Tooltip title="Bold">
          <Button
            active={editor?.isActive("bold")}
            color="tertiary"
            onClick={() => editor?.chain().focus().toggleBold().run()}
            size="sm"
            variant="naked"
          >
            <BoldIcon className="h-5 w-auto" />
          </Button>
        </Tooltip>
        <Tooltip title="Italic">
          <Button
            active={editor?.isActive("italic")}
            color="tertiary"
            onClick={() => editor?.chain().focus().toggleItalic().run()}
            size="sm"
            variant="naked"
          >
            <ItalicIcon className="h-5 w-auto" />
          </Button>
        </Tooltip>
        <Tooltip title="Underline">
          <Button
            active={editor?.isActive("underline")}
            color="tertiary"
            onClick={() => editor?.chain().focus().toggleUnderline().run()}
            size="sm"
            variant="naked"
          >
            <UnderlineIcon className="h-5 w-auto" />
          </Button>
        </Tooltip>
        <Tooltip title="Strikethrough">
          <Button
            active={editor?.isActive("strike")}
            color="tertiary"
            onClick={() => editor?.chain().focus().toggleStrike().run()}
            size="sm"
            variant="naked"
          >
            <StrikeThroughIcon className="h-5 w-auto" />
          </Button>
        </Tooltip>
        <Tooltip title="Link">
          <Button
            active={editor?.isActive("link")}
            color="tertiary"
            size="sm"
            variant="naked"
          >
            <LinkIcon className="h-5 w-auto" />
          </Button>
        </Tooltip>
        <span className="opacity-30">|</span>
        <Tooltip title="Ordered list">
          <Button
            active={editor?.isActive("orderedList")}
            color="tertiary"
            onClick={() => editor?.chain().focus().toggleOrderedList().run()}
            size="sm"
            variant="naked"
          >
            <OrderedListIcon className="h-5 w-auto" />
          </Button>
        </Tooltip>
        <Tooltip title="Unordered list">
          <Button
            active={editor?.isActive("bulletList")}
            color="tertiary"
            onClick={() => editor?.chain().focus().toggleBulletList().run()}
            size="sm"
            variant="naked"
          >
            <UnorderedListIcon className="h-5 w-auto" />
          </Button>
        </Tooltip>
        <Tooltip title="Check list">
          <Button
            active={editor?.isActive("taskList")}
            color="tertiary"
            onClick={() => editor?.chain().focus().toggleTaskList().run()}
            size="sm"
            variant="naked"
          >
            <CheckListIcon className="h-5 w-auto" />
          </Button>
        </Tooltip>
        <Tooltip title="Code">
          <Button
            active={editor?.isActive("code")}
            color="tertiary"
            onClick={() => editor?.chain().focus().toggleCode().run()}
            size="sm"
            variant="naked"
          >
            <CodeIcon className="h-5 w-auto" />
          </Button>
        </Tooltip>
        <span className="opacity-30">|</span>
        <Tooltip title="Quote">
          <Button
            active={editor?.isActive("blockquote")}
            color="tertiary"
            onClick={() => editor?.chain().focus().toggleBlockquote().run()}
            size="sm"
            variant="naked"
          >
            <QuoteIcon className="h-5 w-auto" />
          </Button>
        </Tooltip>
        <Tooltip title="Code block">
          <Button
            active={editor?.isActive("codeBlock")}
            color="tertiary"
            onClick={() => editor?.chain().focus().toggleCodeBlock().run()}
            size="sm"
            variant="naked"
          >
            <CodeBlockIcon className="h-5 w-auto" />
          </Button>
        </Tooltip>
      </Flex>

      <Flex align="center" gap={2}>
        <Tooltip title="Print">
          <Button color="tertiary" size="sm" variant="outline">
            <PrintIcon className="h-5 w-auto" />
          </Button>
        </Tooltip>
        <Menu>
          <Menu.Button>
            <Button
              className="aspect-square"
              color="tertiary"
              leftIcon={<MoreVerticalIcon className="h-5 w-auto" />}
              size="sm"
              variant="naked"
            >
              <span className="sr-only">More actions</span>
            </Button>
          </Menu.Button>
          <Menu.Items align="end" className="w-48">
            <Menu.Group>
              <Menu.Item>
                <EditIcon className="h-4 w-auto" />
                Edit
              </Menu.Item>
              <Menu.Item>
                <LinkIcon className="h-4 w-auto" />
                Copy link
              </Menu.Item>
              <Menu.Item>
                <DuplicateIcon className="h-4 w-auto" />
                Duplicate
              </Menu.Item>
              <Menu.Item>
                <FileLockedIcon className="h-4 w-auto" />
                Lock
              </Menu.Item>
              <Menu.Item>
                <DeleteIcon className="h-4 w-auto text-danger" />
                Delete
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
    </Container>
  );
};
