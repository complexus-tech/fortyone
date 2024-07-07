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
import type { StoryPriority, StoryStatus } from "@/modules/stories/types";
import { StatusesMenu } from "./story/statuses-menu";
import { StoryStatusIcon } from "./story-status-icon";
import { PrioritiesMenu } from "./story/priorities-menu";
import { PriorityIcon } from "./priority-icon";

export const NewStoryDialog = ({
  isOpen,
  setIsOpen,
  status = "Backlog",
  priority = "No Priority",
}: {
  isOpen: boolean;
  setIsOpen: Dispatch<SetStateAction<boolean>>;
  status?: StoryStatus;
  priority?: StoryPriority;
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
      Placeholder.configure({ placeholder: "Story description" }),
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
            <ArrowRightIcon className="h-4 w-auto opacity-40" strokeWidth={3} />
            <Text color="muted">New story</Text>
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
              <span className="sr-only">Expand story to full screen</span>
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
                  leftIcon={<StoryStatusIcon className="h-4" status={status} />}
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
                  leftIcon={
                    <PriorityIcon className="h-4" priority={priority} />
                  }
                  size="xs"
                  type="button"
                  variant="outline"
                >
                  {priority}
                </Button>
              </PrioritiesMenu.Trigger>
              <PrioritiesMenu.Items priority={priority} />
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
              <span className="sr-only">Add labels to the story</span>
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
            Create story
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
