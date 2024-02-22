"use client";
import { type Dispatch, type SetStateAction } from "react";
import { Button, Badge, Dialog, Flex, TextEditor, DatePicker } from "ui";
import { useEditor } from "@tiptap/react";
import Placeholder from "@tiptap/extension-placeholder";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import TextExt from "@tiptap/extension-text";
import { CalendarPlusIcon, CalendarIcon, PlusIcon, ProjectsIcon } from "icons";

export const NewProjectDialog = ({
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
      Placeholder.configure({ placeholder: "Project name" }),
    ],
    content: "",
    editable: true,
  });
  const descriptionEditor = useEditor({
    extensions: [
      Document,
      Paragraph,
      TextExt,
      Placeholder.configure({ placeholder: "Enter description..." }),
    ],
    content: "",
    editable: true,
  });

  return (
    <Dialog onOpenChange={setIsOpen} open={isOpen}>
      <Dialog.Content hideClose size="lg">
        <Dialog.Header className="flex items-center justify-between px-6 pt-6">
          <Dialog.Title className="flex items-center gap-1 text-lg">
            <Badge color="tertiary">New project</Badge>
          </Dialog.Title>
          <Dialog.Close />
        </Dialog.Header>
        <Dialog.Body className="pt-0">
          <Flex gap={2}>
            <Button
              className="relative top-[0.1rem] aspect-square"
              color="tertiary"
              leftIcon={
                <ProjectsIcon className="relative left-[0.15rem] h-5 w-auto text-gray-250" />
              }
              size="xs"
              variant="outline"
            >
              <span className="sr-only">Change project icon</span>
            </Button>
            <TextEditor
              asTitle
              className="text-2xl font-medium"
              editor={titleEditor}
            />
          </Flex>
          <TextEditor editor={descriptionEditor} />
          <Flex className="mt-8" gap={1}>
            <DatePicker>
              <DatePicker.Trigger>
                <Button
                  className="px-2 text-sm"
                  color="tertiary"
                  leftIcon={<CalendarIcon className="h-4 w-auto" />}
                  size="xs"
                  variant="outline"
                >
                  Sep 27
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar />
            </DatePicker>
            <DatePicker>
              <DatePicker.Trigger>
                <Button
                  className="px-2 text-sm"
                  color="tertiary"
                  leftIcon={<CalendarPlusIcon className="h-4 w-auto" />}
                  size="xs"
                  variant="outline"
                >
                  Due date
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar />
            </DatePicker>
            <DatePicker>
              <DatePicker.Trigger>
                <Button
                  className="px-2 text-sm"
                  color="tertiary"
                  leftIcon={<CalendarIcon className="h-4 w-auto" />}
                  size="xs"
                  variant="outline"
                >
                  Sep 27
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar />
            </DatePicker>
            <DatePicker>
              <DatePicker.Trigger>
                <Button
                  className="px-2 text-sm"
                  color="tertiary"
                  leftIcon={<CalendarIcon className="h-4 w-auto" />}
                  size="xs"
                  variant="outline"
                >
                  Sep 27
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar />
            </DatePicker>
          </Flex>
        </Dialog.Body>
        <Dialog.Footer className="flex items-center justify-end gap-2">
          <Button
            className="px-5"
            color="tertiary"
            onClick={() => {
              setIsOpen(false);
            }}
            variant="outline"
          >
            Discard
          </Button>
          <Button leftIcon={<PlusIcon className="h-5 w-auto" />}>
            Create project
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
