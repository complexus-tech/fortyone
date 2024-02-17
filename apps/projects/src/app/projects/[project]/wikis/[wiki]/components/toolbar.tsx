import { Button, Container, Flex, Tooltip, Menu } from "ui";
import {
  Bold,
  Code,
  Code2,
  Italic,
  List,
  ListChecks,
  ListOrdered,
  Quote,
  Strikethrough,
  Underline,
  Link,
  ChevronDown,
  MoreVertical,
  Pencil,
  Trash2,
  Link2,
  Copy,
  LockKeyhole,
  Globe2,
  Undo,
  Redo,
} from "lucide-react";
import type { Editor } from "@tiptap/react";
import { cn } from "lib";

export const Toolbar = ({ editor }: { editor: Editor | null }) => {
  const headings = [
    {
      name: "Paragraph",
      level: 0,
      classes: "",
    },
    {
      name: "Heading 1",
      level: 1,
      classes: "text-4xl",
    },
    {
      name: "Heading 2",
      level: 2,
      classes: "text-3xl",
    },
    {
      name: "Heading 3",
      level: 3,
      classes: "text-2xl",
    },
    {
      name: "Heading 4",
      level: 4,
      classes: "text-xl",
    },
    {
      name: "Heading 5",
      level: 5,
      classes: "text-lg",
    },
    {
      name: "Heading 6",
      level: 6,
      classes: "text-base font-medium text-gray-250 dark:text-gray-200",
    },
  ];
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
            <Undo className="h-5 w-auto" />
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
            <Redo className="h-5 w-auto" />
          </Button>
        </Tooltip>

        <Menu>
          <Menu.Button>
            <Button
              color="tertiary"
              rightIcon={<ChevronDown className="h-4 w-auto" />}
              size="md"
              variant="naked"
            >
              Paragraph
            </Button>
          </Menu.Button>
          <Menu.Items align="center" className="w-64">
            <Menu.Group>
              {headings.map(({ name, level, classes }) => (
                <Menu.Item
                  active={editor?.isActive("heading", { level })}
                  className={cn("justify-between px-3", classes)}
                  key={name}
                >
                  {name}
                  {/* {editor?.isActive("heading", { level }) && (
                    <Check className="h-5 w-auto" strokeWidth={2.1} />
                  )} */}
                </Menu.Item>
              ))}
            </Menu.Group>
          </Menu.Items>
        </Menu>
        <Tooltip title="Bold">
          <Button
            active={editor?.isActive("bold")}
            color="tertiary"
            onClick={() => editor?.chain().focus().toggleBold().run()}
            size="sm"
            variant="naked"
          >
            <Bold className="h-5 w-auto" />
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
            <Italic className="h-5 w-auto" />
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
            <Underline className="h-5 w-auto" />
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
            <Strikethrough className="h-5 w-auto" />
          </Button>
        </Tooltip>
        <Tooltip title="Link">
          <Button
            active={editor?.isActive("link")}
            color="tertiary"
            size="sm"
            variant="naked"
          >
            <Link className="h-5 w-auto" />
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
            <ListOrdered className="h-5 w-auto" />
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
            <List className="h-5 w-auto" />
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
            <ListChecks className="h-5 w-auto" />
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
            <Code className="h-5 w-auto" />
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
            <Quote className="h-5 w-auto" />
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
            <Code2 className="h-5 w-auto" />
          </Button>
        </Tooltip>
      </Flex>

      <Flex align="center" gap={2}>
        <Button
          color="tertiary"
          leftIcon={<Globe2 className="h-[1.1rem] w-auto" />}
          size="sm"
          variant="outline"
        >
          Make private
        </Button>
        <Menu>
          <Menu.Button>
            <Button
              className="aspect-square"
              color="tertiary"
              leftIcon={<MoreVertical className="h-5 w-auto" />}
              size="sm"
              variant="naked"
            >
              <span className="sr-only">More actions</span>
            </Button>
          </Menu.Button>
          <Menu.Items align="end" className="w-48">
            <Menu.Group>
              <Menu.Item>
                <Pencil className="h-4 w-auto" />
                Edit
              </Menu.Item>
              <Menu.Item>
                <Link2 className="h-4 w-auto" />
                Copy link
              </Menu.Item>
              <Menu.Item>
                <Copy className="h-4 w-auto" />
                Duplicate
              </Menu.Item>
              <Menu.Item>
                <LockKeyhole className="h-4 w-auto" />
                Lock wiki
              </Menu.Item>
              <Menu.Item>
                <Trash2 className="h-4 w-auto text-danger" />
                Delete
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
    </Container>
  );
};
