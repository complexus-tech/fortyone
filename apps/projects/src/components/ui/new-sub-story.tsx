"use client";
import type { Dispatch, SetStateAction } from "react";
import { useEffect, useState } from "react";
import { Button, Flex, TextEditor, DatePicker, Box, Avatar } from "ui";
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
import { CalendarIcon, CloseIcon, PlusIcon } from "icons";
import { toast } from "sonner";
import { addDays, format } from "date-fns";
import { cn } from "lib";
import { useSession } from "next-auth/react";
import type { NewStory } from "@/modules/story/types";
import type { StoryPriority } from "@/modules/stories/types";
import { useCreateStoryMutation } from "@/modules/story/hooks/create-mutation";
import { useTeamStatuses } from "@/lib/hooks/statuses";
import { AssigneesMenu } from "@/components/ui/story/assignees-menu";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { PriorityIcon } from "./priority-icon";
import { PrioritiesMenu } from "./story/priorities-menu";
import { StoryStatusIcon } from "./story-status-icon";
import { StatusesMenu } from "./story/statuses-menu";

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
  isOpen: boolean;
  setIsOpen: Dispatch<SetStateAction<boolean>>;
}) => {
  const session = useSession();
  const { data: statuses = [] } = useTeamStatuses(teamId);
  const { data: members = [] } = useTeamMembers(teamId);
  const defaultStatus =
    statuses.find((status) => status.id === statusId) || statuses.at(0);

  const initialForm: NewStory = {
    title: "",
    description: "",
    descriptionHTML: "",
    teamId,
    statusId: defaultStatus?.id,
    endDate: null,
    startDate: null,
    assigneeId: null,
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

    const newStory: NewStory = {
      title: titleEditor.getText(),
      description: editor.getText(),
      descriptionHTML: editor.getHTML(),
      teamId,
      priority: storyForm.priority,
      statusId: storyForm.statusId,
      endDate: storyForm.endDate,
      startDate: storyForm.startDate,
      reporterId: session.data?.user?.id,
      parentId,
      assigneeId: storyForm.assigneeId,
    };

    try {
      await mutation.mutateAsync(newStory);
      titleEditor.commands.setContent("");
      editor.commands.setContent("");
      setStoryForm(initialForm);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (isOpen) {
      titleEditor?.commands.focus();
    }
  }, [isOpen, titleEditor]);

  return (
    <Box>
      {isOpen ? (
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
                  setStatusId={(id) => {
                    setStoryForm((prev) => ({ ...prev, statusId: id }));
                  }}
                  statusId={storyForm.statusId}
                  teamId={teamId}
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
                      storyForm.startDate ? (
                        <CloseIcon
                          aria-label="Remove date"
                          className="h-4 w-auto"
                          onClick={() => {
                            setStoryForm((prev) => ({
                              ...prev,
                              startDate: null,
                            }));
                          }}
                          role="button"
                        />
                      ) : null
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
                      storyForm.endDate ? (
                        <CloseIcon
                          aria-label="Remove date"
                          className="h-4 w-auto"
                          onClick={() => {
                            setStoryForm((prev) => ({
                              ...prev,
                              endDate: null,
                            }));
                          }}
                          role="button"
                        />
                      ) : null
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
              <AssigneesMenu>
                <AssigneesMenu.Trigger>
                  <Button
                    className="gap-1.5 px-2 text-sm"
                    color="tertiary"
                    leftIcon={
                      <Avatar
                        color="tertiary"
                        name={
                          members.find(
                            (member) => member.id === storyForm.assigneeId,
                          )?.fullName
                        }
                        size="xs"
                        src={
                          members.find(
                            (member) => member.id === storyForm.assigneeId,
                          )?.avatarUrl
                        }
                      />
                    }
                    size="xs"
                    variant="outline"
                  >
                    {members.find(
                      (member) => member.id === storyForm.assigneeId,
                    )?.username || "Assignee"}
                  </Button>
                </AssigneesMenu.Trigger>
                <AssigneesMenu.Items
                  assigneeId={storyForm.assigneeId}
                  onAssigneeSelected={(assigneeId) => {
                    setStoryForm({ ...storyForm, assigneeId });
                  }}
                />
              </AssigneesMenu>
            </Flex>
            <Flex gap={1}>
              <Button
                className="px-2"
                color="tertiary"
                onClick={() => {
                  setIsOpen(false);
                  titleEditor?.commands.setContent("");
                  editor?.commands.setContent("");
                }}
                size="xs"
                variant="naked"
              >
                Cancel
              </Button>
              <Button
                color="tertiary"
                leftIcon={<PlusIcon className="h-4 w-auto" />}
                loading={loading}
                loadingText="Creating story..."
                onClick={handleCreateStory}
                size="xs"
              >
                Create
              </Button>
            </Flex>
          </Flex>
        </Box>
      ) : null}
    </Box>
  );
};
