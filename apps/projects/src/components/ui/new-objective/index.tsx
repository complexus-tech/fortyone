"use client";
import { useEffect, useState, type Dispatch, type SetStateAction } from "react";
import {
  Button,
  Dialog,
  Flex,
  Text,
  TextEditor,
  DatePicker,
  Menu,
  Avatar,
  Box,
  Divider,
  Wrapper,
} from "ui";
import { useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Underline from "@tiptap/extension-underline";
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
  CrownIcon,
} from "icons";
import { toast } from "sonner";
import { format } from "date-fns";
import { cn } from "lib";
import { useRouter } from "next/navigation";
import {
  useFeatures,
  useLocalStorage,
  useTerminology,
  useUserRole,
} from "@/hooks";
import type { Team } from "@/modules/teams/types";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useMembers } from "@/lib/hooks/members";
import { AssigneesMenu } from "@/components/ui/story/assignees-menu";
import type { NewKeyResult, NewObjective } from "@/modules/objectives/types";
import { useCreateObjectiveMutation } from "@/modules/objectives/hooks";
import { useObjectiveStatuses } from "@/lib/hooks/objective-statuses";
import {
  useObjectives,
  useTeamObjectives,
} from "@/modules/objectives/hooks/use-objectives";
import { FeatureGuard } from "@/components/ui/feature-guard";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { TeamColor } from "../team-color";
import { PriorityIcon } from "../priority-icon";
import { StoryStatusIcon } from "../story-status-icon";
import { PrioritiesMenu } from "../story/priorities-menu";
import { ObjectiveStatusesMenu } from "../objective-statuses-menu";
import { KeyResultEditor } from "./key-result-editor";
import { KeyResultsList } from "./key-results-list";

type KeyResultFormMode = "add" | "edit" | null;

type KeyResultUpdate = Partial<NewKeyResult>;

export const NewObjectiveDialog = ({
  isOpen,
  setIsOpen,
  teamId: initialTeamId,
}: {
  isOpen: boolean;
  setIsOpen: Dispatch<SetStateAction<boolean>>;
  teamId?: string;
}) => {
  const router = useRouter();
  const { userRole } = useUserRole();
  const { data: teams = [] } = useTeams();
  const { data: members = [] } = useMembers();
  const { data: statuses = [] } = useObjectiveStatuses();
  const { data: objectives = [] } = useObjectives();
  const { getTermDisplay } = useTerminology();
  const { tier, getLimit } = useSubscriptionFeatures();
  const [isExpanded, setIsExpanded] = useState(false);
  const firstTeam = teams.length > 0 ? teams[0] : null;
  const [activeTeam, setActiveTeam] = useLocalStorage<Team | null>(
    "activeTeam",
    firstTeam,
  );

  // Add validation to ensure activeTeam exists in teams list
  const validActiveTeam =
    teams.find((team) => team.id === activeTeam?.id) || firstTeam;

  // Update the current team logic
  const currentTeamId = initialTeamId || validActiveTeam?.id;
  const currentTeam =
    teams.find((team) => team.id === currentTeamId) || firstTeam;
  const defaultStatus =
    statuses.find((status) => status.isDefault) || statuses[0];

  // Add effect to update activeTeam if it's not valid
  useEffect(() => {
    if (!teams.find((team) => team.id === activeTeam?.id)) {
      setActiveTeam(firstTeam);
    }
  }, [teams, activeTeam, setActiveTeam, firstTeam]);

  const initialForm: NewObjective = {
    name: "",
    description: "",
    leadUser: null,
    teamId: currentTeamId || "",
    startDate: null,
    endDate: null,
    statusId: statuses.length > 0 ? defaultStatus.id : "",
    priority: "No Priority",
    keyResults: [],
  };
  const features = useFeatures();
  const [objectiveForm, setObjectiveForm] = useState<NewObjective>(initialForm);
  const [keyResultMode, setKeyResultMode] = useState<KeyResultFormMode>(null);
  const [editingKeyResult, setEditingKeyResult] = useState<NewKeyResult | null>(
    null,
  );
  const [editingIndex, setEditingIndex] = useState<number | null>(null);
  const { data: teamObjectives = [] } = useTeamObjectives(
    currentTeam?.id ?? "",
  );
  const createMutation = useCreateObjectiveMutation();

  const titleEditor = useEditor({
    extensions: [
      Document,
      Paragraph,
      TextExt,
      Placeholder.configure({ placeholder: "eg. Increase revenue by 20%" }),
    ],
    content: "",
    editable: true,
    immediatelyRender: false,
    autofocus: true,
  });

  const editor = useEditor({
    extensions: [
      StarterKit,
      Underline,
      Link.configure({
        autolink: true,
      }),
      Placeholder.configure({ placeholder: "Add description..." }),
    ],
    content: "",
    editable: true,
    immediatelyRender: false,
  });

  const handleCreateObjective = () => {
    if (!titleEditor || !editor) return;
    if (!titleEditor.getText()) {
      titleEditor.commands.focus();
      toast.warning("Validation Error", {
        description: "Title is required",
      });
      return;
    }
    if (
      teamObjectives.some(
        (objective) =>
          objective.name.toLowerCase() === titleEditor.getText().toLowerCase(),
      )
    ) {
      toast.warning("Validation Error", {
        description: `${getTermDisplay("objectiveTerm", { capitalize: true })} with this name already exists`,
      });
      return;
    }

    createMutation.mutate({
      ...objectiveForm,
      name: titleEditor.getText(),
      description: editor.getText(),
    });
    setIsOpen(false);
    setIsExpanded(false);
    titleEditor.commands.setContent("");
    editor.commands.setContent("");
    setObjectiveForm(initialForm);
  };

  useEffect(() => {
    if (isOpen) {
      titleEditor?.commands.focus();
    }
    if (initialTeamId) {
      const team = teams.find((team) => team.id === initialTeamId);
      if (team) {
        setActiveTeam(team);
      }
    }
  }, [isOpen, initialTeamId, teams, setActiveTeam, titleEditor]);

  const lead = members.find((member) => member.id === objectiveForm.leadUser);

  useEffect(() => {
    if (isOpen && teams.length === 0) {
      toast.warning("Join or create a team", {
        description: "You need to be part of a team to create an objective",
        action: {
          label: "Join a team",
          onClick: () => {
            router.push("/settings/workspace/teams");
          },
        },
      });
      setIsOpen(false);
    }
  }, [isOpen, teams, setIsOpen, router]);

  return (
    <FeatureGuard
      count={objectives.length}
      fallback={
        <Dialog open={isOpen}>
          <Dialog.Content hideClose>
            <Dialog.Header className="flex items-center gap-2 px-6 pt-6 text-xl">
              <CrownIcon className="relative -top-px h-6 text-warning" />
              <Dialog.Title>
                {getLimit("maxObjectives") === 0 ? (
                  <>
                    Your plan does not support creating{" "}
                    {getTermDisplay("objectiveTerm", {
                      variant: "plural",
                    })}
                  </>
                ) : (
                  <>
                    {getTermDisplay("objectiveTerm", {
                      variant: "plural",
                      capitalize: true,
                    })}{" "}
                    Limit Reached
                  </>
                )}
              </Dialog.Title>
            </Dialog.Header>
            <Dialog.Body>
              <Text
                className="mb-4 dark:font-normal"
                color="muted"
                fontSize="lg"
              >
                You&apos;ve reached the limit of {getLimit("maxObjectives")}{" "}
                {getTermDisplay("objectiveTerm", {
                  variant:
                    getLimit("maxObjectives") !== 1 ? "plural" : "singular",
                })}{" "}
                on your {tier.replace("free", "hobby")} plan.{" "}
                {userRole === "admin"
                  ? "Upgrade "
                  : "Ask your admin to upgrade "}
                to create unlimited{" "}
                {getTermDisplay("objectiveTerm", {
                  variant: "plural",
                })}{" "}
                and unlock premium features.
              </Text>
              <Wrapper className="dark:bg-dark-300/60">
                <Flex align="center" gap={3} justify="between">
                  <Text color="muted" fontSize="lg">
                    Current plan:
                  </Text>
                  <Text fontSize="lg" transform="capitalize">
                    {tier.replace("free", "hobby")}
                  </Text>
                </Flex>
                <Divider className="my-3" />
                <Flex align="center" gap={3} justify="between">
                  <Text color="muted" fontSize="lg">
                    {getTermDisplay("objectiveTerm", {
                      variant: "plural",
                      capitalize: true,
                    })}
                    :
                  </Text>
                  <Text color="primary" fontSize="lg">
                    {objectives.length}/{getLimit("maxObjectives")}
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
      feature="maxObjectives"
    >
      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content
          className={cn("max-w-4xl", {
            "max-w-5xl": isExpanded,
          })}
          hideClose
        >
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
                        active={team.id === currentTeam?.id}
                        className="justify-between gap-3"
                        key={team.id}
                        onClick={() => {
                          setActiveTeam(team);
                          setObjectiveForm((prev) => ({
                            ...prev,
                            teamId: team.id,
                          }));
                        }}
                      >
                        <span className="flex items-center gap-1.5">
                          <TeamColor className="shrink-0" color={team.color} />
                          <span className="block truncate">{team.name}</span>
                        </span>
                        {team.id === currentTeam?.id && (
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
                New {getTermDisplay("objectiveTerm")}
              </Text>
            </Dialog.Title>
            <Flex gap={2}>
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
              className={cn({
                "min-h-96": isExpanded,
              })}
              editor={editor}
            />
            <Flex align="center" className="mt-4 gap-1.5" wrap>
              <ObjectiveStatusesMenu>
                <ObjectiveStatusesMenu.Trigger>
                  <Button
                    color="tertiary"
                    leftIcon={
                      <StoryStatusIcon statusId={objectiveForm.statusId} />
                    }
                    size="sm"
                    type="button"
                    variant="outline"
                  >
                    {
                      statuses.find((s) => s.id === objectiveForm.statusId)
                        ?.name
                    }
                  </Button>
                </ObjectiveStatusesMenu.Trigger>
                <ObjectiveStatusesMenu.Items
                  setStatusId={(statusId) => {
                    setObjectiveForm((prev) => ({
                      ...prev,
                      statusId,
                    }));
                  }}
                  statusId={objectiveForm.statusId}
                />
              </ObjectiveStatusesMenu>
              <PrioritiesMenu>
                <PrioritiesMenu.Trigger>
                  <Button
                    color="tertiary"
                    leftIcon={
                      <PriorityIcon priority={objectiveForm.priority} />
                    }
                    size="sm"
                    type="button"
                    variant="outline"
                  >
                    {objectiveForm.priority ?? "No Priority"}
                  </Button>
                </PrioritiesMenu.Trigger>
                <PrioritiesMenu.Items
                  priority={objectiveForm.priority}
                  setPriority={(priority) => {
                    setObjectiveForm((prev) => ({
                      ...prev,
                      priority,
                    }));
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
                      objectiveForm.startDate ? (
                        <CloseIcon
                          aria-label="Remove date"
                          className="h-4 w-auto"
                          onClick={(e) => {
                            e.stopPropagation();
                            setObjectiveForm((prev) => ({
                              ...prev,
                              startDate: null,
                            }));
                          }}
                          role="button"
                        />
                      ) : null
                    }
                    size="sm"
                    variant="outline"
                  >
                    {objectiveForm.startDate
                      ? format(new Date(objectiveForm.startDate), "MMM d, yyyy")
                      : "Start date"}
                  </Button>
                </DatePicker.Trigger>
                <DatePicker.Calendar
                  onDayClick={(date) => {
                    setObjectiveForm((prev) => ({
                      ...prev,
                      startDate: date.toISOString(),
                    }));
                  }}
                />
              </DatePicker>
              <DatePicker>
                <DatePicker.Trigger>
                  <Button
                    className="px-2 text-sm"
                    color="tertiary"
                    leftIcon={<CalendarIcon className="h-4 w-auto" />}
                    rightIcon={
                      objectiveForm.endDate ? (
                        <CloseIcon
                          aria-label="Remove date"
                          className="h-4 w-auto"
                          onClick={(e) => {
                            e.stopPropagation();
                            setObjectiveForm((prev) => ({
                              ...prev,
                              endDate: null,
                            }));
                          }}
                          role="button"
                        />
                      ) : null
                    }
                    size="sm"
                    variant="outline"
                  >
                    {objectiveForm.endDate
                      ? format(new Date(objectiveForm.endDate), "MMM d, yyyy")
                      : "Target date"}
                  </Button>
                </DatePicker.Trigger>
                <DatePicker.Calendar
                  fromDate={
                    objectiveForm.startDate
                      ? new Date(objectiveForm.startDate)
                      : undefined
                  }
                  onDayClick={(date) => {
                    setObjectiveForm((prev) => ({
                      ...prev,
                      endDate: date.toISOString(),
                    }));
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
                        name={lead?.fullName}
                        size="xs"
                        src={lead?.avatarUrl}
                      />
                    }
                    size="sm"
                    variant="outline"
                  >
                    <span className="relative -top-px inline-block max-w-[12ch] truncate">
                      {lead?.username || "Lead"}
                    </span>
                  </Button>
                </AssigneesMenu.Trigger>
                <AssigneesMenu.Items
                  assigneeId={objectiveForm.leadUser || null}
                  onAssigneeSelected={(leadUserId: string | null) => {
                    setObjectiveForm((prev) => ({
                      ...prev,
                      leadUser: leadUserId || undefined,
                    }));
                  }}
                />
              </AssigneesMenu>
            </Flex>
            {features.keyResultEnabled ? (
              <>
                <Divider className="my-4" />
                <Box>
                  <Text className="mb-3 font-medium">
                    {getTermDisplay("keyResultTerm", {
                      variant: "plural",
                      capitalize: true,
                    })}
                  </Text>
                  {keyResultMode === null ? (
                    <>
                      <KeyResultsList
                        keyResults={objectiveForm.keyResults || []}
                        onEdit={(index) => {
                          const kr = objectiveForm.keyResults?.[index];
                          if (kr) {
                            setEditingKeyResult(kr);
                            setEditingIndex(index);
                            setKeyResultMode("edit");
                          }
                        }}
                        onRemove={(index) => {
                          setObjectiveForm((prev) => ({
                            ...prev,
                            keyResults:
                              prev.keyResults?.filter((_, i) => i !== index) ||
                              [],
                          }));
                        }}
                      />
                      <Button
                        color="tertiary"
                        leftIcon={<PlusIcon />}
                        onClick={() => {
                          const newKr: NewKeyResult = {
                            name: "",
                            measurementType: "number",
                            currentValue: 0,
                            startValue: 0,
                            targetValue: 0,
                          };
                          setEditingKeyResult(newKr);
                          setEditingIndex(null);
                          setKeyResultMode("add");
                        }}
                        variant="outline"
                      >
                        Add{" "}
                        {getTermDisplay("keyResultTerm", { capitalize: true })}
                      </Button>
                    </>
                  ) : (
                    <KeyResultEditor
                      keyResult={editingKeyResult}
                      onCancel={() => {
                        setKeyResultMode(null);
                        setEditingKeyResult(null);
                        setEditingIndex(null);
                      }}
                      onSave={() => {
                        if (keyResultMode === "add") {
                          setObjectiveForm((prev) => ({
                            ...prev,
                            keyResults: [
                              ...(prev.keyResults || []),
                              {
                                ...editingKeyResult!,
                                currentValue: editingKeyResult!.startValue,
                              },
                            ],
                          }));
                        } else if (editingIndex !== null) {
                          setObjectiveForm((prev) => ({
                            ...prev,
                            keyResults:
                              prev.keyResults?.map((kr, i) =>
                                i === editingIndex ? editingKeyResult! : kr,
                              ) || [],
                          }));
                        }
                        setKeyResultMode(null);
                        setEditingKeyResult(null);
                        setEditingIndex(null);
                      }}
                      onUpdate={(_index: number, updates: KeyResultUpdate) => {
                        setEditingKeyResult((prev) => {
                          if (!prev) return null;
                          return {
                            ...prev,
                            ...updates,
                          };
                        });
                      }}
                    />
                  )}
                </Box>
              </>
            ) : null}
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
            <Button
              disabled={editingKeyResult !== null}
              leftIcon={<PlusIcon className="text-white dark:text-gray-200" />}
              loading={createMutation.isPending}
              loadingText="Creating..."
              onClick={handleCreateObjective}
              size="md"
            >
              Create {getTermDisplay("objectiveTerm", { capitalize: true })}
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </FeatureGuard>
  );
};
