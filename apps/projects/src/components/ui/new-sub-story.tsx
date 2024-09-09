"use client";
import { Dispatch, SetStateAction, useEffect, useState } from "react";
import { Button, Flex, TextEditor, DatePicker, Box } from "ui";
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
import { CalendarIcon, CloseIcon, PlusIcon, TagsIcon } from "icons";
import type { StoryPriority } from "@/modules/stories/types";
import { StatusesMenu } from "./story/statuses-menu";
import { StoryStatusIcon } from "./story-status-icon";
import { PrioritiesMenu } from "./story/priorities-menu";
import { PriorityIcon } from "./priority-icon";
import { NewStory } from "@/modules/story/types";
import { toast } from "sonner";
import nProgress from "nprogress";
import { addDays, format } from "date-fns";
import { cn } from "lib";
import { useSession } from "next-auth/react";
import { useCreateStoryMutation } from "@/modules/story/hooks/create-mutation";
import { useStatuses } from "@/lib/hooks/statuses";

export const NewSubStory = ({
  statusId,
  teamId,
  parentId,
  priority = "No Priority",
  isOpen,
  setIsOpen,
}: {
  statusId?: string;
  teamId: string;
  parentId: string;
  priority?: StoryPriority;
  isOpen?: boolean;
  setIsOpen?: Dispatch<SetStateAction<boolean>>;
}) => {
  const session = useSession();
  const { data: statuses = [] } = useStatuses();
  const { id: defaultStateId } = (statuses.find(
    (state) => state.id === statusId,
  ) || statuses.at(0))!!;

  const initialForm: NewStory = {
    title: "",
    description: "",
    descriptionHTML: "",
    teamId: teamId,
    statusId: defaultStateId,
    endDate: null,
    startDate: null,
    priority,
  };
  const [storyForm, setStoryForm] = useState<NewStory>(initialForm);
  const [loading, setLoading] = useState(false);
  const mutation = useCreateStoryMutation();

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

  const handleCreateStory = async () => {
    if (!titleEditor || !editor) return;
    if (!titleEditor.getText()) {
      titleEditor.commands.focus();
      toast.warning("Validation Error", {
        description: "Title is required",
      });
      return;
    }
    setLoading(true);
    nProgress.start();

    const newStory: NewStory = {
      title: titleEditor.getText(),
      description: editor.getText(),
      descriptionHTML: editor.getHTML(),
      teamId,
      priority: storyForm.priority,
      statusId: storyForm.statusId,
      endDate: storyForm.endDate,
      startDate: storyForm.startDate,
      reporterId: session?.data?.user?.id,
      parentId,
      // assigneeId: "",
    };

    try {
      await mutation.mutateAsync(newStory);
      titleEditor.commands.setContent("");
      editor.commands.setContent("");
      setStoryForm(initialForm);
    } catch (error) {
      console.log(error);
    } finally {
      setLoading(false);
      nProgress.done();
    }
  };

  useEffect(() => {
    if (isOpen) {
      titleEditor?.commands.focus();
    }
  }, [isOpen]);

  return (
    <Box>
      {isOpen && (
        <Box className="mt-2 rounded-lg border border-gray-100/60 bg-gray-50/40 p-3 dark:border-dark-100 dark:bg-dark-300">
          <TextEditor
            asTitle
            className="text-xl font-medium"
            editor={titleEditor}
          />
          <TextEditor editor={editor} />

          <Flex align="center" justify="between">
            <Flex gap={1}>
              <StatusesMenu>
                <StatusesMenu.Trigger>
                  <Button
                    color="tertiary"
                    leftIcon={
                      <StoryStatusIcon
                        className="h-4"
                        statusId={storyForm.statusId}
                      />
                    }
                    size="xs"
                    type="button"
                    variant="outline"
                  >
                    {
                      statuses.find((state) => state.id === storyForm.statusId)
                        ?.name
                    }
                  </Button>
                </StatusesMenu.Trigger>
                <StatusesMenu.Items
                  statusId={storyForm.statusId}
                  setStatusId={(id) => {
                    setStoryForm((prev) => ({ ...prev, statusId: id }));
                  }}
                />
              </StatusesMenu>
              <PrioritiesMenu>
                <PrioritiesMenu.Trigger>
                  <Button
                    color="tertiary"
                    leftIcon={
                      <PriorityIcon
                        className="h-4"
                        priority={storyForm.priority}
                      />
                    }
                    size="xs"
                    type="button"
                    variant="outline"
                  >
                    {storyForm.priority}
                  </Button>
                </PrioritiesMenu.Trigger>
                <PrioritiesMenu.Items
                  priority={storyForm.priority}
                  setPriority={(priority) => {
                    setStoryForm((prev) => ({ ...prev, priority }));
                  }}
                />
              </PrioritiesMenu>
              <DatePicker>
                <DatePicker.Trigger>
                  <Button
                    className="px-2 text-sm"
                    color="tertiary"
                    leftIcon={<CalendarIcon className="h-4 w-auto" />}
                    rightIcon={
                      storyForm.startDate && (
                        <CloseIcon
                          role="button"
                          onClick={() => {
                            setStoryForm((prev) => ({
                              ...prev,
                              startDate: null,
                            }));
                          }}
                          aria-label="Remove date"
                          className="h-4 w-auto"
                        />
                      )
                    }
                    size="xs"
                    variant="outline"
                  >
                    {storyForm.startDate
                      ? format(new Date(storyForm.startDate), "MMM d, yyyy")
                      : "Start date"}
                  </Button>
                </DatePicker.Trigger>
                <DatePicker.Calendar
                  onDayClick={(date) => {
                    setStoryForm({
                      ...storyForm,
                      startDate: date.toISOString(),
                    });
                  }}
                />
              </DatePicker>
              <DatePicker>
                <DatePicker.Trigger>
                  <Button
                    className={cn("px-2 text-sm", {
                      "text-primary dark:text-primary": storyForm.endDate
                        ? new Date(storyForm.endDate) < new Date()
                        : false,
                      "text-warning dark:text-warning": storyForm.endDate
                        ? new Date(storyForm.endDate) <=
                            addDays(new Date(), 7) &&
                          new Date(storyForm.endDate) >= new Date()
                        : false,
                    })}
                    color="tertiary"
                    leftIcon={<CalendarIcon className="h-4 w-auto" />}
                    rightIcon={
                      storyForm.endDate && (
                        <CloseIcon
                          role="button"
                          onClick={() => {
                            setStoryForm((prev) => ({
                              ...prev,
                              endDate: null,
                            }));
                          }}
                          aria-label="Remove date"
                          className="h-4 w-auto"
                        />
                      )
                    }
                    size="xs"
                    variant="outline"
                  >
                    {storyForm.endDate
                      ? format(new Date(storyForm.endDate), "MMM d, yyyy")
                      : "Due date"}
                  </Button>
                </DatePicker.Trigger>
                <DatePicker.Calendar
                  fromDate={
                    storyForm.startDate
                      ? new Date(storyForm.startDate)
                      : undefined
                  }
                  onDayClick={(date) => {
                    setStoryForm({ ...storyForm, endDate: date.toISOString() });
                  }}
                />
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
            <Flex gap={1}>
              <Button
                size="xs"
                className="px-2"
                variant="naked"
                color="tertiary"
                onClick={() => {
                  setIsOpen(false);
                  titleEditor?.commands.setContent("");
                  editor?.commands.setContent("");
                }}
              >
                Cancel
              </Button>
              <Button
                leftIcon={<PlusIcon className="h-4 w-auto" />}
                size="xs"
                color="tertiary"
                onClick={handleCreateStory}
                loading={loading}
                loadingText="Creating story..."
              >
                Create
              </Button>
            </Flex>
          </Flex>
        </Box>
      )}
    </Box>
  );
};
