"use client";
import { useEffect, useState, type Dispatch, type SetStateAction } from "react";
import {
  Button,
  Badge,
  Dialog,
  Flex,
  Switch,
  Text,
  TextEditor,
  DatePicker,
  Menu,
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
  CheckIcon,
  CloseIcon,
  MaximizeIcon,
  PlusIcon,
  TagsIcon,
} from "icons";
import type { StoryPriority } from "@/modules/stories/types";
import { StatusesMenu } from "./story/statuses-menu";
import { StoryStatusIcon } from "./story-status-icon";
import { PrioritiesMenu } from "./story/priorities-menu";
import { PriorityIcon } from "./priority-icon";
import { NewStory } from "@/modules/story/types";
import { createStoryAction } from "@/modules/story/actions/create-story";
import { toast } from "sonner";
import nProgress, { set } from "nprogress";
import { useRouter } from "next/navigation";
import { slugify } from "@/utils";
import { addDays, format } from "date-fns";
import { cn } from "lib";
import { useStore } from "@/hooks/store";
import { useLocalStorage } from "@/hooks";
import { Team } from "@/modules/teams/types";

export const NewStoryDialog = ({
  isOpen,
  setIsOpen,
  statusId,
  priority = "No Priority",
}: {
  isOpen: boolean;
  setIsOpen: Dispatch<SetStateAction<boolean>>;
  statusId?: string;
  priority?: StoryPriority;
}) => {
  const { states, teams } = useStore();
  const [activeTeam, setActiveTeam] = useLocalStorage<Team>(
    "activeTeam",
    teams.at(0)!!,
  );
  const { id: defaultStateId } = (states.find(
    (state) => state.id === statusId,
  ) || states.at(0))!!;

  const initialForm: NewStory = {
    title: "",
    description: "",
    descriptionHTML: "",
    teamId: activeTeam.id,
    statusId: defaultStateId,
    endDate: null,
    startDate: null,
    priority,
  };

  const [storyForm, setStoryForm] = useState<NewStory>(initialForm);
  const [loading, setLoading] = useState(false);
  const [createMore, setCreateMore] = useState(false);
  const router = useRouter();

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
      teamId: activeTeam.id,
      priority: storyForm.priority,
      statusId: storyForm.statusId,
      endDate: storyForm.endDate,
      startDate: storyForm.startDate,
      // assigneeId: "",
    };

    try {
      const createdStory = await createStoryAction(newStory);
      toast.success("Success", {
        description: "Story created successfully",
        action: {
          label: "View story",
          onClick: () => {
            nProgress.start();
            titleEditor.commands.setContent("");
            editor.commands.setContent("");
            router.push(
              `/story/${createdStory.id}/${slugify(createdStory.title)}`,
            );
          },
        },
      });
      if (!createMore) {
        setIsOpen(false);
      }
      titleEditor.commands.setContent("");
      editor.commands.setContent("");
      setStoryForm(initialForm);
    } catch (error) {
      console.log(error);
      toast.error("Error", {
        description: "Failed to create story",
      });
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
    <Dialog onOpenChange={setIsOpen} open={isOpen}>
      <Dialog.Content hideClose size="lg">
        <Dialog.Header className="flex items-center justify-between px-6 pt-6">
          <Dialog.Title className="flex items-center gap-1 text-lg">
            <Menu>
              <Menu.Button>
                <Button
                  size="xs"
                  className="gap-1 font-semibold tracking-wide"
                  leftIcon={<span>{activeTeam.icon}</span>}
                  color="tertiary"
                >
                  {activeTeam.code}
                </Button>
              </Menu.Button>
              <Menu.Items className="w-52" align="start">
                <Menu.Group>
                  {teams.map((team) => (
                    <Menu.Item
                      active={team.id === activeTeam.id}
                      onClick={() => {
                        setActiveTeam(team);
                      }}
                      className="justify-between gap-3"
                      key={team.id}
                    >
                      <span className="flex items-center gap-1.5">
                        <span className="shrink-0">{team.icon}</span>
                        <span className="block truncate">{team.name}</span>
                      </span>
                      {team.id === activeTeam.id && (
                        <CheckIcon className="h-[1.1rem] w-auto" />
                      )}
                    </Menu.Item>
                  ))}
                </Menu.Group>
              </Menu.Items>
            </Menu>
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
                    states.find((state) => state.id === storyForm.statusId)
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
                  setStoryForm({ ...storyForm, startDate: date.toISOString() });
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
                      ? new Date(storyForm.endDate) <= addDays(new Date(), 7) &&
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
                          setStoryForm((prev) => ({ ...prev, endDate: null }));
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
        </Dialog.Body>
        <Dialog.Footer className="flex items-center justify-between gap-2">
          <Text color="muted">
            <label className="flex items-center gap-2" htmlFor="more">
              Create more{" "}
              <Switch
                id="more"
                checked={createMore}
                onCheckedChange={setCreateMore}
              />
            </label>
          </Text>
          <Button
            leftIcon={<PlusIcon className="h-5 w-auto" />}
            size="md"
            onClick={handleCreateStory}
            loading={loading}
            loadingText="Creating story..."
          >
            Create story
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
