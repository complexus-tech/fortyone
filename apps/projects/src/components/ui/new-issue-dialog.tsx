"use client";
import { type Dispatch, type SetStateAction } from "react";
import { Button, Badge, Dialog, Flex, Switch, Text, TextEditor } from "ui";
import { ChevronRight, Maximize2, Plus } from "lucide-react";
import { useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Underline from "@tiptap/extension-underline";
import TaskItem from "@tiptap/extension-task-item";
import TaskList from "@tiptap/extension-task-list";
import Link from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import TextExt from "@tiptap/extension-text";
import { PrioritiesMenu } from "./issue/priorities-menu";

export const NewIssueDialog = ({
  isOpen,
  setIsOpen,
}: {
  isOpen: boolean;
  setIsOpen: Dispatch<SetStateAction<boolean>>;
}) => {
  const titleEditor = useEditor({
    extensions: [
      Document,
      Paragraph,
      TextExt,
      Placeholder.configure({ placeholder: "Enter title..." }),
    ],
    content: "This is a sample title.",
    editable: true,
  });

  const editor = useEditor({
    extensions: [
      StarterKit,
      Underline,
      TaskList,
      TaskItem.configure({
        nested: true,
      }),
      Link.configure({
        autolink: true,
      }),
      Placeholder.configure({ placeholder: "Issue description" }),
    ],
    content: "",
    editable: true,
  });

  return (
    <Dialog onOpenChange={setIsOpen} open={isOpen}>
      <Dialog.Content hideClose size="lg">
        <Dialog.Header className="flex items-center justify-between px-6 pt-6">
          <Dialog.Title className="flex items-center gap-1 text-lg">
            <Badge color="tertiary">COMP-1</Badge>
            <ChevronRight className="h-4 w-auto opacity-40" strokeWidth={3} />
            <Text color="muted">New issue</Text>
          </Dialog.Title>
          <Flex gap={2}>
            <Button
              className="px-[0.35rem] dark:hover:bg-dark-100"
              color="tertiary"
              href="/"
              size="xs"
              variant="naked"
            >
              <Maximize2 className="h-[1.2rem] w-auto" />
              <span className="sr-only">Expand issue to full screen</span>
            </Button>
            <Dialog.Close />
          </Flex>
        </Dialog.Header>
        <Dialog.Body className="pt-0">
          <TextEditor
            asTitle
            className="text-2xl font-medium"
            editor={titleEditor}
          />
          <TextEditor editor={editor} />
          <Flex className="mt-4" gap={1}>
            <PrioritiesMenu priority="Urgent" />
            <Badge color="tertiary">COMP-1</Badge>
            <Badge color="tertiary">COMP-1</Badge>
            <Badge color="tertiary">COMP-1</Badge>
            <Badge color="tertiary">COMP-1</Badge>
            <Badge color="tertiary">COMP-1</Badge>
          </Flex>
        </Dialog.Body>
        <Dialog.Footer className="flex items-center justify-between gap-2">
          <Text color="muted">
            <label className="flex items-center gap-2" htmlFor="more">
              Create more <Switch id="more" />
            </label>
          </Text>
          <Button leftIcon={<Plus className="h-5 w-auto" />} size="md">
            Create issue
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
