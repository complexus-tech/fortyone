"use client";
import {
  useEffect,
  useState,
  useReducer,
  useMemo,
  type Dispatch,
  type SetStateAction,
} from "react";
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
  Wrapper,
  Divider,
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
import Table from "@tiptap/extension-table";
import TableCell from "@tiptap/extension-table-cell";
import TableHeader from "@tiptap/extension-table-header";
import TableRow from "@tiptap/extension-table-row";
import { marked } from "marked";
import {
  ArrowRightIcon,
  CalendarIcon,
  CheckIcon,
  CloseIcon,
  MaximizeIcon,
  MinimizeIcon,
  PlusIcon,
  ObjectiveIcon,
  SprintsIcon,
  CrownIcon,
} from "icons";
import { toast } from "sonner";
import { addDays, format, formatISO } from "date-fns";
import { cn } from "lib";
import { useSession } from "next-auth/react";
import { useRouter } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import {
  useFeatures,
  useLocalStorage,
  useTerminology,
  useUserRole,
} from "@/hooks";
import type { Team } from "@/modules/teams/types";
import type { NewStory } from "@/modules/story/types";
import type { StoryPriority } from "@/modules/stories/types";
import { useCreateStoryMutation } from "@/modules/story/hooks/create-mutation";
import { useStatuses } from "@/lib/hooks/statuses";
import { AssigneesMenu } from "@/components/ui/story/assignees-menu";
import { useMembers } from "@/lib/hooks/members";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useTeamObjectives } from "@/modules/objectives/hooks/use-objectives";
import { useTeamSprints } from "@/modules/sprints/hooks/team-sprints";
import { useAutomationPreferences } from "@/lib/hooks/users/preferences";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { useTotalStories } from "@/modules/stories/hooks/total-stories";
import { storyKeys } from "@/modules/stories/constants";
import { PriorityIcon } from "./priority-icon";
import { PrioritiesMenu } from "./story/priorities-menu";
import { StoryStatusIcon } from "./story-status-icon";
import { StatusesMenu } from "./story/statuses-menu";
import { TeamColor } from "./team-color";
import { ObjectivesMenu } from "./story/objectives-menu";
import { SprintsMenu } from "./story/sprints-menu";
import { FeatureGuard } from "./feature-guard";

type StoryFormAction =
  | { type: "INITIALIZE"; payload: NewStory }
  | { type: "SET_FIELD"; field: keyof NewStory; value: string | null }
  | { type: "RESET_FORM"; payload: NewStory }
  | { type: "SYNC_TEAM_STATUS"; teamId: string; statusId: string };

const storyFormReducer = (
  state: NewStory,
  action: StoryFormAction,
): NewStory => {
  switch (action.type) {
    case "INITIALIZE":
      return action.payload;
    case "SET_FIELD":
      return { ...state, [action.field]: action.value };
    case "SYNC_TEAM_STATUS":
      return {
        ...state,
        teamId: action.teamId,
        statusId: action.statusId,
      };
    case "RESET_FORM":
      return action.payload;
    default:
      return state;
  }
};

export const NewStoryDialog = ({
  isOpen,
  setIsOpen,
  statusId,
  teamId,
  priority = "No Priority",
  assigneeId,
  objectiveId,
  sprintId,
  description,
}: {
  isOpen: boolean;
  setIsOpen: Dispatch<SetStateAction<boolean>>;
  statusId?: string;
  teamId?: string;
  objectiveId?: string;
  sprintId?: string;
  priority?: StoryPriority;
  assigneeId?: string | null;
  description?: string;
}) => {
  const session = useSession();
  const { userRole } = useUserRole();
  const queryClient = useQueryClient();
  const router = useRouter();
  const features = useFeatures();
  const { data: teams = [] } = useTeams();
  const { data: statuses = [] } = useStatuses();
  const { data: members = [] } = useMembers();
  const { getTermDisplay } = useTerminology();
  const [isExpanded, setIsExpanded] = useState(false);
  const firstTeam = teams.length > 0 ? teams[0] : null;
  const [activeTeam, setActiveTeam] = useLocalStorage<Team | null>(
    "activeTeam",
    firstTeam,
  );

  const validActiveTeam =
    teams.find((team) => team.id === activeTeam?.id) || firstTeam;

  const currentTeamId = teamId || validActiveTeam?.id;
  const currentTeam =
    teams.find((team) => team.id === currentTeamId) || firstTeam;
  const { data: objectives = [] } = useTeamObjectives(currentTeamId ?? "");
  const { data: sprints = [] } = useTeamSprints(currentTeamId ?? "");
  const { data: automationPreferences } = useAutomationPreferences();
  const { tier, getLimit } = useSubscriptionFeatures();
  const { data: totalStories = 0 } = useTotalStories();

  const teamStatuses = statuses.filter(
    (status) => status.teamId === currentTeamId,
  );

  const defaultStatus =
    teamStatuses.find((status) => status.id === statusId) ||
    (teamStatuses.length > 0
      ? teamStatuses.find((status) => status.isDefault) || teamStatuses[0]
      : null);

  const initialForm: NewStory = useMemo(
    () => ({
      title: "",
      description: "",
      descriptionHTML: "",
      teamId: currentTeamId,
      statusId: defaultStatus?.id,
      endDate: null,
      startDate: null,
      assigneeId:
        assigneeId ||
        (automationPreferences?.autoAssignSelf ? session.data?.user?.id : null),
      priority,
      objectiveId: objectiveId || null,
      sprintId: sprintId || null,
    }),
    [
      currentTeamId,
      defaultStatus?.id,
      assigneeId,
      automationPreferences?.autoAssignSelf,
      session.data?.user?.id,
      priority,
      objectiveId,
      sprintId,
    ],
  );

  const [storyForm, dispatch] = useReducer(storyFormReducer, initialForm);
  const [loading, setLoading] = useState(false);
  const [createMore, setCreateMore] = useState(false);
  const mutation = useCreateStoryMutation();
  const objective = objectives.find((o) => o.id === storyForm.objectiveId);
  const sprint = sprints.find((s) => s.id === storyForm.sprintId);
  const member = members.find((m) => m.id === storyForm.assigneeId);

  const titleEditor = useEditor({
    extensions: [
      Document,
      Paragraph,
      TextExt,
      Placeholder.configure({ placeholder: "Enter title..." }),
    ],
    content: "",
    editable: true,
    autofocus: true,
    immediatelyRender: false,
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
      Placeholder.configure({
        placeholder: `${getTermDisplay("storyTerm", { capitalize: true })} description`,
      }),

      Table.configure({
        resizable: true,
      }),
      TableRow,
      TableHeader,
      TableCell,
    ],
    content: marked.parse(description || "", {
      gfm: true,
    }),
    editable: true,
    immediatelyRender: false,
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
      teamId: currentTeamId,
      priority: storyForm.priority,
      statusId: storyForm.statusId,
      endDate: storyForm.endDate,
      startDate: storyForm.startDate,
      assigneeId: storyForm.assigneeId,
      objectiveId: storyForm.objectiveId,
      sprintId: storyForm.sprintId,
    };

    try {
      await mutation.mutateAsync(newStory);
      if (!createMore) {
        setIsOpen(false);
        setIsExpanded(false);
      }
      titleEditor.commands.setContent("");
      editor.commands.setContent("");
      dispatch({ type: "RESET_FORM", payload: initialForm });
      if (tier === "free") {
        queryClient.invalidateQueries({ queryKey: storyKeys.total() });
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (teamId) {
      const team = teams.find((team) => team.id === teamId);
      if (team) {
        setActiveTeam(team);
      }
    }
  }, [isOpen, teamId, teams, setActiveTeam, titleEditor, router, setIsOpen]);

  useEffect(() => {
    const currentStatus = teamStatuses.find(
      (status) => status.id === storyForm.statusId,
    );
    if (!currentStatus && teamStatuses.length > 0 && currentTeamId) {
      dispatch({
        type: "SYNC_TEAM_STATUS",
        teamId: currentTeamId,
        statusId: teamStatuses[0].id,
      });
    }
  }, [currentTeamId, storyForm.statusId, teamStatuses, statusId]);

  useEffect(() => {
    if (!teams.find((team) => team.id === activeTeam?.id)) {
      setActiveTeam(firstTeam);
    }
  }, [teams, activeTeam, setActiveTeam, firstTeam]);

  useEffect(() => {
    if (isOpen && titleEditor) {
      titleEditor.commands.focus();
    }
  }, [isOpen, titleEditor]);

  // Initialize form when props change
  useEffect(() => {
    dispatch({ type: "INITIALIZE", payload: initialForm });
  }, [initialForm]);

  return (
    <FeatureGuard
      count={totalStories}
      fallback={
        <Dialog open={isOpen}>
          <Dialog.Content hideClose>
            <Dialog.Header className="flex items-center gap-2 px-6 pt-6 text-xl">
              <CrownIcon className="relative -top-px h-6 text-warning" />
              <Dialog.Title>
                {getTermDisplay("storyTerm", {
                  capitalize: true,
                })}{" "}
                Limit Reached
              </Dialog.Title>
            </Dialog.Header>
            <Dialog.Body>
              <Text className="mb-4 dark:font-normal" color="muted">
                You&apos;ve reached the limit of {getLimit("maxStories")}{" "}
                {getTermDisplay("storyTerm", {
                  variant: "plural",
                })}{" "}
                on your {tier.replace("free", "hobby")} plan.{" "}
                {userRole === "admin" ? "Upgrade" : "Ask your admin"} to create
                unlimited {getTermDisplay("storyTerm", { variant: "plural" })}{" "}
                and unlock premium features.
              </Text>
              <Wrapper className="dark:bg-dark-300/60">
                <Flex align="center" gap={3} justify="between">
                  <Text color="muted">Current plan:</Text>
                  <Text transform="capitalize">
                    {tier.replace("free", "hobby")}
                  </Text>
                </Flex>
                <Divider className="my-3" />
                <Flex align="center" gap={3} justify="between">
                  <Text color="muted">
                    {getTermDisplay("storyTerm", {
                      variant: "plural",
                      capitalize: true,
                    })}
                    :
                  </Text>
                  <Text color="primary">
                    {totalStories}/{getLimit("maxStories")}
                  </Text>
                </Flex>
              </Wrapper>
              {userRole === "admin" && (
                <Button
                  align="center"
                  className="mt-4 border-0"
                  fullWidth
                  href="/settings/workspace/billing"
                  rounded="lg"
                  size="lg"
                >
                  Upgrade now
                </Button>
              )}
              <Button
                align="center"
                className="mb-2 mt-3 border-[0.5px]"
                color="tertiary"
                fullWidth
                onClick={() => {
                  setIsOpen(false);
                }}
                rounded="lg"
                size="lg"
              >
                Maybe later
              </Button>
            </Dialog.Body>
          </Dialog.Content>
        </Dialog>
      }
      feature="maxStories"
    >
      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content hideClose size={isExpanded ? "xl" : "lg"}>
          <Dialog.Header className="flex items-center justify-between px-6 pt-6">
            <Dialog.Title className="flex items-center gap-1 text-lg">
              <Menu>
                <Menu.Button>
                  <Button
                    className="gap-1.5 font-semibold tracking-wide"
                    color="tertiary"
                    leftIcon={<TeamColor color={currentTeam?.color} />}
                    size="xs"
                  >
                    {currentTeam?.code}
                  </Button>
                </Menu.Button>
                <Menu.Items align="start" className="w-52">
                  <Menu.Group>
                    {teams.map((team) => (
                      <Menu.Item
                        active={team.id === activeTeam?.id}
                        className="justify-between gap-3"
                        key={team.id}
                        onClick={() => {
                          setActiveTeam(team);
                        }}
                      >
                        <span className="flex items-center gap-1.5">
                          <TeamColor className="shrink-0" color={team.color} />
                          <span className="block truncate">{team.name}</span>
                        </span>
                        {team.id === activeTeam?.id && (
                          <CheckIcon className="h-[1.1rem] w-auto" />
                        )}
                      </Menu.Item>
                    ))}
                  </Menu.Group>
                </Menu.Items>
              </Menu>
              <ArrowRightIcon
                className="h-4 w-auto opacity-30"
                strokeWidth={3}
              />
              <Text className="opacity-80" color="muted">
                New {getTermDisplay("storyTerm")}
              </Text>
            </Dialog.Title>
            <Flex gap={2}>
              <Tooltip title={isExpanded ? "Minimize dialog" : "Expand dialog"}>
                <Button
                  className="px-[0.35rem] dark:hover:bg-dark-100"
                  color="tertiary"
                  onClick={() => {
                    setIsExpanded((prev) => !prev);
                  }}
                  size="xs"
                  variant="naked"
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
          <Dialog.Body className="max-h-[60dvh] pt-0">
            <TextEditor
              asTitle
              className="text-2xl font-medium"
              editor={titleEditor}
            />
            <TextEditor
              className={cn("min-h-20", {
                "min-h-96": isExpanded,
              })}
              editor={editor}
            />
            <Flex align="center" className="mt-4 gap-1.5" wrap>
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
                    size="sm"
                    type="button"
                    variant="outline"
                  >
                    {
                      teamStatuses.find(
                        (state) => state.id === storyForm.statusId,
                      )?.name
                    }
                  </Button>
                </StatusesMenu.Trigger>
                <StatusesMenu.Items
                  setStatusId={(statusId) => {
                    dispatch({
                      type: "SET_FIELD",
                      field: "statusId",
                      value: statusId,
                    });
                  }}
                  statusId={storyForm.statusId}
                  teamId={currentTeamId ?? ""}
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
                    size="sm"
                    type="button"
                    variant="outline"
                  >
                    {storyForm.priority}
                  </Button>
                </PrioritiesMenu.Trigger>
                <PrioritiesMenu.Items
                  priority={storyForm.priority}
                  setPriority={(priority) => {
                    dispatch({
                      type: "SET_FIELD",
                      field: "priority",
                      value: priority,
                    });
                  }}
                />
              </PrioritiesMenu>
              <DatePicker>
                <DatePicker.Trigger>
                  <Button
                    className="px-2"
                    color="tertiary"
                    leftIcon={<CalendarIcon className="h-4 w-auto" />}
                    rightIcon={
                      storyForm.startDate ? (
                        <CloseIcon
                          aria-label="Remove date"
                          className="h-4 w-auto"
                          onClick={() => {
                            dispatch({
                              type: "SET_FIELD",
                              field: "startDate",
                              value: null,
                            });
                          }}
                          role="button"
                        />
                      ) : null
                    }
                    size="sm"
                    variant="outline"
                  >
                    {storyForm.startDate
                      ? format(new Date(storyForm.startDate), "MMM d, yyyy")
                      : "Start date"}
                  </Button>
                </DatePicker.Trigger>
                <DatePicker.Calendar
                  onDayClick={(date) => {
                    dispatch({
                      type: "SET_FIELD",
                      field: "startDate",
                      value: formatISO(date, { representation: "date" }),
                    });
                  }}
                />
              </DatePicker>
              <DatePicker>
                <DatePicker.Trigger>
                  <Button
                    className={cn("px-2", {
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
                            dispatch({
                              type: "SET_FIELD",
                              field: "endDate",
                              value: null,
                            });
                          }}
                          role="button"
                        />
                      ) : null
                    }
                    size="sm"
                    variant="outline"
                  >
                    {storyForm.endDate
                      ? format(new Date(storyForm.endDate), "MMM d, yyyy")
                      : "Deadline"}
                  </Button>
                </DatePicker.Trigger>
                <DatePicker.Calendar
                  fromDate={
                    storyForm.startDate
                      ? new Date(storyForm.startDate)
                      : undefined
                  }
                  onDayClick={(date) => {
                    dispatch({
                      type: "SET_FIELD",
                      field: "endDate",
                      value: formatISO(date, { representation: "date" }),
                    });
                  }}
                />
              </DatePicker>
              <AssigneesMenu>
                <AssigneesMenu.Trigger>
                  <Button
                    className="gap-1.5 px-2"
                    color="tertiary"
                    leftIcon={
                      <Avatar
                        name={member?.fullName}
                        size="xs"
                        src={member?.avatarUrl}
                      />
                    }
                    size="sm"
                    variant="outline"
                  >
                    <span className="relative -top-px inline-block max-w-[12ch] truncate">
                      {member?.username || "Assignee"}
                    </span>
                  </Button>
                </AssigneesMenu.Trigger>
                <AssigneesMenu.Items
                  assigneeId={storyForm.assigneeId}
                  onAssigneeSelected={(assigneeId) => {
                    dispatch({
                      type: "SET_FIELD",
                      field: "assigneeId",
                      value: assigneeId,
                    });
                  }}
                  teamId={currentTeamId}
                />
              </AssigneesMenu>
              {features.objectiveEnabled && objectives.length > 0 ? (
                <ObjectivesMenu>
                  <ObjectivesMenu.Trigger>
                    <Button
                      className="gap-1 px-2"
                      color="tertiary"
                      leftIcon={<ObjectiveIcon className="h-4 w-auto" />}
                      size="sm"
                      variant="outline"
                    >
                      <span className="inline-block max-w-[12ch] truncate">
                        {objective?.name ||
                          getTermDisplay("objectiveTerm", { capitalize: true })}
                      </span>
                    </Button>
                  </ObjectivesMenu.Trigger>
                  <ObjectivesMenu.Items
                    objectiveId={storyForm.objectiveId ?? undefined}
                    setObjectiveId={(objectiveId) => {
                      dispatch({
                        type: "SET_FIELD",
                        field: "objectiveId",
                        value: objectiveId,
                      });
                    }}
                    teamId={currentTeamId}
                  />
                </ObjectivesMenu>
              ) : null}
              {features.sprintEnabled && sprints.length > 0 ? (
                <SprintsMenu>
                  <SprintsMenu.Trigger>
                    <Button
                      className="gap-1 px-2"
                      color="tertiary"
                      leftIcon={<SprintsIcon className="h-4 w-auto" />}
                      size="sm"
                      variant="outline"
                    >
                      <span className="inline-block max-w-[12ch] truncate">
                        {sprint?.name ||
                          getTermDisplay("sprintTerm", { capitalize: true })}
                      </span>
                    </Button>
                  </SprintsMenu.Trigger>
                  <SprintsMenu.Items
                    setSprintId={(sprintId) => {
                      dispatch({
                        type: "SET_FIELD",
                        field: "sprintId",
                        value: sprintId,
                      });
                    }}
                    sprintId={storyForm.sprintId ?? undefined}
                    teamId={currentTeamId}
                  />
                </SprintsMenu>
              ) : null}
            </Flex>
          </Dialog.Body>
          <Dialog.Footer className="flex items-center justify-between gap-2">
            <Text color="muted">
              <label className="flex items-center gap-2" htmlFor="more">
                Create more
                <Switch
                  checked={createMore}
                  id="more"
                  onCheckedChange={setCreateMore}
                />
              </label>
            </Text>
            <Button
              leftIcon={<PlusIcon className="text-white dark:text-gray-200" />}
              loading={loading}
              loadingText={`Creating ${getTermDisplay("storyTerm")}...`}
              onClick={handleCreateStory}
              size="md"
            >
              Create {getTermDisplay("storyTerm")}
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </FeatureGuard>
  );
};
