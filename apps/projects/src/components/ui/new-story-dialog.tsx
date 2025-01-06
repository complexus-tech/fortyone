"use client";
import { useEffect, useState, type Dispatch, type SetStateAction } from "react";
import {
  Button,
  Dialog,
  Flex,
  Switch,
  Text,
  TextEditor,
  DatePicker,
  Menu,
  Tooltip,
  Avatar,
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
  MinimizeIcon,
  PlusIcon,
  TagsIcon,
} from "icons";
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
import { useLocalStorage } from "@/hooks";
import { Team } from "@/modules/teams/types";
import { useSession } from "next-auth/react";
import { useCreateStoryMutation } from "@/modules/story/hooks/create-mutation";
import { useStatuses } from "@/lib/hooks/statuses";
import { AssigneesMenu } from "@/components/ui/story/assignees-menu";
import { useMembers } from "@/lib/hooks/members";
import { useTeams } from "@/modules/teams/hooks/teams";
import { TeamColor } from "./team-color";
export const NewStoryDialog = ({
  isOpen,
  setIsOpen,
  statusId,
  teamId,
  priority = "No Priority",
  assigneeId,
}: {
  isOpen: boolean;
  setIsOpen: Dispatch<SetStateAction<boolean>>;
  statusId?: string;
  teamId?: string;
  priority?: StoryPriority;
  assigneeId?: string | null;
}) => {
  const session = useSession();
  const { data: teams = [] } = useTeams();
  const { data: statuses = [] } = useStatuses();
  const { data: members = [] } = useMembers();
  const [isExpanded, setIsExpanded] = useState(false);
  const [activeTeam, setActiveTeam] = useLocalStorage<Team>(
    "activeTeam",
    teams.at(0)!!,
  );
  const { id: defaultStateId } = (statuses.find(
    (state) => state.id === statusId,
  ) || statuses.at(0))!!;

  const initialForm: NewStory = {
    title: "",
    description: "",
    descriptionHTML: "",
    teamId: teamId || activeTeam.id,
    statusId: defaultStateId,
    endDate: null,
    startDate: null,
    assigneeId,
    priority,
  };
  const [storyForm, setStoryForm] = useState<NewStory>(initialForm);
  const [loading, setLoading] = useState(false);
  const [createMore, setCreateMore] = useState(false);
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
      teamId: activeTeam.id,
      priority: storyForm.priority,
      statusId: storyForm.statusId,
      endDate: storyForm.endDate,
      startDate: storyForm.startDate,
      reporterId: session?.data?.user?.id,
      assigneeId: storyForm.assigneeId,
    };

    try {
      await mutation.mutateAsync(newStory);
      if (!createMore) {
        setIsOpen(false);
        setIsExpanded(false);
      }
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
    if (teamId) {
      const team = teams.find((team) => team.id === teamId);
      if (team) {
        setActiveTeam(team);
      }
    }
  }, [isOpen]);

  return (
    <Dialog onOpenChange={setIsOpen} open={isOpen}>
      <Dialog.Content hideClose size={isExpanded ? "xl" : "lg"}>
        <Dialog.Header className="flex items-center justify-between px-6 pt-6">
          <Dialog.Title className="flex items-center gap-1 text-lg">
            <Menu>
              <Menu.Button>
                <Button
                  size="xs"
                  className="gap-1.5 font-semibold tracking-wide"
                  leftIcon={<TeamColor color={activeTeam.color} />}
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
                        <TeamColor color={team?.color} className="shrink-0" />
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
            <Tooltip title={isExpanded ? "Minimize dialog" : "Expand dialog"}>
              <Button
                className="px-[0.35rem] dark:hover:bg-dark-100"
                color="tertiary"
                size="xs"
                variant="naked"
                onClick={() => {
                  setIsExpanded((prev) => !prev);
                }}
              >
                {isExpanded ? (
                  <MinimizeIcon className="h-[1.2rem] w-auto" />
                ) : (
                  <MaximizeIcon className="h-[1.2rem] w-auto" />
                )}
                <span className="sr-only">
                  {isExpanded ? "Minimize" : "Expand"} dialog
                </span>
              </Button>
            </Tooltip>
            <Dialog.Close />
          </Flex>
        </Dialog.Header>
        <Dialog.Body className="max-h-[60vh] pt-0">
          <TextEditor
            asTitle
            className="text-2xl font-medium"
            editor={titleEditor}
          />
          <TextEditor
            editor={editor}
            className={cn({
              "min-h-96": isExpanded,
            })}
          />
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
                    statuses.find((state) => state.id === storyForm.statusId)
                      ?.name
                  }
                </Button>
              </StatusesMenu.Trigger>
              <StatusesMenu.Items
                statusId={storyForm.statusId}
                setStatusId={(statusId) => {
                  setStoryForm((prev) => ({ ...prev, statusId }));
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
                  {members.find((member) => member.id === storyForm.assigneeId)
                    ?.username || "Assignee"}
                </Button>
              </AssigneesMenu.Trigger>
              <AssigneesMenu.Items
                assigneeId={storyForm.assigneeId}
                onAssigneeSelected={(assigneeId) => {
                  setStoryForm({ ...storyForm, assigneeId });
                }}
              />
            </AssigneesMenu>
            {/* <Button
              className="px-2 text-sm"
              color="tertiary"
              leftIcon={<TagsIcon className="h-4 w-auto" />}
              size="xs"
              variant="outline"
            >
              <span className="sr-only">Add labels to the story</span>
            </Button> */}
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
            leftIcon={<PlusIcon className="text-white dark:text-gray-200" />}
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
