"use client";
import { type Dispatch, type SetStateAction } from "react";
import {
  Button,
  Badge,
  Dialog,
  Flex,
  Switch,
  Text,
  TextEditor,
  DatePicker,
} from "ui";
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
import {
  ArrowRightIcon,
  CalendarIcon,
  MaximizeIcon,
  PlusIcon,
  TagsIcon,
} from "icons";
import type { IssueStatus } from "@/types/issue";
import { StatusesMenu } from "./issue/statuses-menu";
import { IssueStatusIcon } from "./issue-status-icon";
import { PrioritiesMenu } from "./issue/priorities-menu";
import { PriorityIcon } from "./priority-icon";

export const NewIssueDialog = ({
  isOpen,
  setIsOpen,
  status = "Backlog",
}: {
  isOpen: boolean;
  setIsOpen: Dispatch<SetStateAction<boolean>>;
  status?: IssueStatus;
}) => {
  const titleEditor = useEditor({
    extensions: [
      Document,
      Paragraph,
      TextExt,
      Placeholder.configure({ placeholder: "Enter title..." }),
    ],
    content: "",
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
            <Badge color="tertiary" rounded="sm">
              COMP-1
            </Badge>
            <ArrowRightIcon className="h-4 w-auto opacity-40" strokeWidth={3} />
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
              <MaximizeIcon className="h-[1.2rem] w-auto" />
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
            <StatusesMenu>
              <StatusesMenu.Trigger>
                <Button
                  color="tertiary"
                  leftIcon={<IssueStatusIcon className="h-4" status={status} />}
                  size="xs"
                  type="button"
                  variant="outline"
                >
                  {status}
                </Button>
              </StatusesMenu.Trigger>
              <StatusesMenu.Items status={status} />
            </StatusesMenu>
            <PrioritiesMenu>
              <PrioritiesMenu.Trigger>
                <Button
                  color="tertiary"
                  leftIcon={<PriorityIcon className="h-4" />}
                  size="xs"
                  type="button"
                  variant="outline"
                >
                  No Priority
                </Button>
              </PrioritiesMenu.Trigger>
              <PrioritiesMenu.Items />
            </PrioritiesMenu>
            <DatePicker>
              <DatePicker.Trigger>
                <Button
                  className="px-2 text-sm"
                  color="tertiary"
                  leftIcon={<CalendarIcon className="h-4 w-auto" />}
                  size="xs"
                  variant="outline"
                >
                  Due date
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar />
            </DatePicker>
            <Button
              className="px-2 text-sm"
              color="tertiary"
              leftIcon={<TagsIcon className="h-4 w-auto" />}
              size="xs"
              variant="outline"
            >
              <span className="sr-only">Add labels to the issue</span>
            </Button>
          </Flex>
        </Dialog.Body>
        <Dialog.Footer className="flex items-center justify-between gap-2">
          <Text color="muted">
            <label className="flex items-center gap-2" htmlFor="more">
              Create more <Switch id="more" />
            </label>
          </Text>
          <Button leftIcon={<PlusIcon className="h-5 w-auto" />} size="md">
            Create issue
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
