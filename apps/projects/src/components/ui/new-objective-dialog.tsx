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
} from "icons";
import { toast } from "sonner";
import nProgress from "nprogress";
import { format } from "date-fns";
import { cn } from "lib";
import { useSession } from "next-auth/react";
import { useLocalStorage } from "@/hooks";
import type { Team } from "@/modules/teams/types";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useMembers } from "@/lib/hooks/members";
import { TeamColor } from "./team-color";

type ObjectiveHealth = "On Track" | "At Risk" | "Off Track" | "Not Started";
type ObjectiveStatus =
  | "Not Started"
  | "In Progress"
  | "Completed"
  | "Cancelled";

type MeasureType = "Number" | "Percent (%)" | "Boolean (Complete/Incomplete)";

type KeyResult = {
  id: string;
  name: string;
  measureType: MeasureType;
  startValue: number;
  targetValue: number;
};

type NewObjective = {
  name: string;
  description: string;
  descriptionHTML: string;
  teamId: string;
  status: ObjectiveStatus;
  health: ObjectiveHealth;
  startDate: string | null;
  endDate: string | null;
  leadUserId: string | null;
  keyResults: KeyResult[];
};

const KeyResultEditor = ({
  keyResult,
  onUpdate,
  onRemove,
}: {
  keyResult: KeyResult;
  onUpdate: (id: string, updates: Partial<KeyResult>) => void;
  onRemove: (id: string) => void;
}) => {
  const editor = useEditor({
    extensions: [
      Document,
      Paragraph,
      TextExt,
      Placeholder.configure({
        placeholder: "Example: Increase user adoption from 100 to 150",
      }),
    ],
    content: keyResult.name,
    editable: true,
    onUpdate: ({ editor }) => {
      onUpdate(keyResult.id, { name: editor.getText() });
    },
  });

  return (
    <Box className="mb-6 rounded-lg border border-gray-200 px-6 pb-6 dark:border-dark-100">
      <Box className="space-y-3">
        <TextEditor className="flex-1" editor={editor} />
        <Flex gap={4}>
          <Box className="flex-1">
            <Flex align="center" className="mb-2" gap={2}>
              <Text className="font-medium">Measure as</Text>
              <Text color="muted">Required</Text>
            </Flex>
            <Menu>
              <Menu.Button className="w-full">
                <Button
                  className="w-full justify-between"
                  color="tertiary"
                  rightIcon={<ArrowRightIcon className="h-4 w-4 rotate-90" />}
                  size="md"
                  variant="outline"
                >
                  {keyResult.measureType}
                </Button>
              </Menu.Button>
              <Menu.Items align="start" className="w-full">
                <Menu.Group>
                  {[
                    "Number",
                    "Percent (%)",
                    "Boolean (Complete/Incomplete)",
                  ].map((type) => (
                    <Menu.Item
                      active={type === keyResult.measureType}
                      key={type}
                      onClick={() => {
                        onUpdate(keyResult.id, {
                          measureType: type as MeasureType,
                        });
                      }}
                    >
                      {type}
                    </Menu.Item>
                  ))}
                </Menu.Group>
              </Menu.Items>
            </Menu>
          </Box>
          <Box className="flex-1">
            <Text className="mb-2 font-medium">Starting Value</Text>
            <input
              className="w-full rounded-md border border-gray-200 px-3 py-2 dark:border-dark-100 dark:bg-dark-200"
              onChange={(e) => {
                onUpdate(keyResult.id, { startValue: Number(e.target.value) });
              }}
              placeholder="0"
              type="number"
              value={keyResult.startValue || ""}
            />
          </Box>
          <Box className="flex-1">
            <Text className="mb-2 font-medium">Target Value</Text>
            <input
              className="w-full rounded-md border border-gray-200 px-3 py-2 dark:border-dark-100 dark:bg-dark-200"
              onChange={(e) => {
                onUpdate(keyResult.id, { targetValue: Number(e.target.value) });
              }}
              placeholder="0"
              type="number"
              value={keyResult.targetValue || ""}
            />
          </Box>
        </Flex>
        <Flex className="mb-4" gap={2}>
          <Button
            color="tertiary"
            onClick={() => {
              onRemove(keyResult.id);
            }}
            size="sm"
          >
            Remove
          </Button>
          <Button size="sm">Create Key Result</Button>
        </Flex>
      </Box>
    </Box>
  );
};

export const NewObjectiveDialog = ({
  isOpen,
  setIsOpen,
  teamId: initialTeamId,
}: {
  isOpen: boolean;
  setIsOpen: Dispatch<SetStateAction<boolean>>;
  teamId?: string;
}) => {
  const session = useSession();
  const { data: teams = [] } = useTeams();
  const { data: members = [] } = useMembers();
  const [isExpanded, setIsExpanded] = useState(false);
  const [activeTeam, setActiveTeam] = useLocalStorage<Team>(
    "activeTeam",
    teams.at(0)!,
  );

  const initialForm: NewObjective = {
    name: "",
    description: "",
    descriptionHTML: "",
    teamId: initialTeamId || activeTeam.id,
    status: "Not Started",
    health: "Not Started",
    startDate: null,
    endDate: null,
    leadUserId: session.data?.user?.id || null,
    keyResults: [],
  };

  const [objectiveForm, setObjectiveForm] = useState<NewObjective>(initialForm);
  const [loading, setLoading] = useState(false);
  const [keyResultEditors, setKeyResultEditors] = useState<
    Record<string, ReturnType<typeof useEditor>>
  >({});

  const titleEditor = useEditor({
    extensions: [
      Document,
      Paragraph,
      TextExt,
      Placeholder.configure({ placeholder: "Enter objective name..." }),
    ],
    content: "",
    editable: true,
  });

  const editor = useEditor({
    extensions: [
      StarterKit,
      Underline,
      Link.configure({
        autolink: true,
      }),
      Placeholder.configure({ placeholder: "Objective description" }),
    ],
    content: "",
    editable: true,
  });

  const handleCreateObjective = async () => {
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

    try {
      // TODO: Implement objective creation mutation
      await Promise.resolve(); // Placeholder for actual API call
      setIsOpen(false);
      setIsExpanded(false);
      titleEditor.commands.setContent("");
      editor.commands.setContent("");
      Object.values(keyResultEditors).forEach((editor) => {
        editor?.commands.setContent("");
      });
      setObjectiveForm(initialForm);
      setKeyResultEditors({});
    } finally {
      setLoading(false);
      nProgress.done();
    }
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

  const lead = members.find((member) => member.id === objectiveForm.leadUserId);

  return (
    <Dialog onOpenChange={setIsOpen} open={isOpen}>
      <Dialog.Content hideClose size={isExpanded ? "xl" : "lg"}>
        <Dialog.Header className="flex items-center justify-between px-6 pt-6">
          <Dialog.Title className="flex items-center gap-1 text-lg">
            <Menu>
              <Menu.Button>
                <Button
                  className="gap-1.5 font-semibold tracking-wide"
                  color="tertiary"
                  leftIcon={<TeamColor color={activeTeam.color} />}
                  size="xs"
                >
                  {activeTeam.code}
                </Button>
              </Menu.Button>
              <Menu.Items align="start" className="w-52">
                <Menu.Group>
                  {teams.map((team) => (
                    <Menu.Item
                      active={team.id === activeTeam.id}
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
                      {team.id === activeTeam.id && (
                        <CheckIcon className="h-[1.1rem] w-auto" />
                      )}
                    </Menu.Item>
                  ))}
                </Menu.Group>
              </Menu.Items>
            </Menu>
            <ArrowRightIcon className="h-4 w-auto opacity-40" strokeWidth={3} />
            <Text color="muted">New Objective</Text>
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
        <Dialog.Body className="max-h-[60vh] pt-0">
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
            <Menu>
              <Menu.Button>
                <Button
                  color="tertiary"
                  size="sm"
                  type="button"
                  variant="outline"
                >
                  {objectiveForm.status}
                </Button>
              </Menu.Button>
              <Menu.Items align="start" className="w-52">
                {["Not Started", "In Progress", "Completed", "Cancelled"].map(
                  (status) => (
                    <Menu.Item
                      active={status === objectiveForm.status}
                      key={status}
                      onClick={() => {
                        setObjectiveForm((prev) => ({
                          ...prev,
                          status: status as ObjectiveStatus,
                        }));
                      }}
                    >
                      {status}
                    </Menu.Item>
                  ),
                )}
              </Menu.Items>
            </Menu>
            <Menu>
              <Menu.Button>
                <Button
                  color="tertiary"
                  size="sm"
                  type="button"
                  variant="outline"
                >
                  {objectiveForm.health}
                </Button>
              </Menu.Button>
              <Menu.Items align="start" className="w-52">
                {["Not Started", "On Track", "At Risk", "Off Track"].map(
                  (health) => (
                    <Menu.Item
                      active={health === objectiveForm.health}
                      key={health}
                      onClick={() => {
                        setObjectiveForm((prev) => ({
                          ...prev,
                          health: health as ObjectiveHealth,
                        }));
                      }}
                    >
                      {health}
                    </Menu.Item>
                  ),
                )}
              </Menu.Items>
            </Menu>
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
            <Menu>
              <Menu.Button>
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
              </Menu.Button>
              <Menu.Items align="start" className="w-52">
                <Menu.Group>
                  {members.map((member) => (
                    <Menu.Item
                      active={member.id === objectiveForm.leadUserId}
                      className="justify-between gap-3"
                      key={member.id}
                      onClick={() => {
                        setObjectiveForm((prev) => ({
                          ...prev,
                          leadUserId: member.id,
                        }));
                      }}
                    >
                      <span className="flex items-center gap-1.5">
                        <Avatar
                          className="shrink-0"
                          name={member.fullName}
                          size="xs"
                          src={member.avatarUrl}
                        />
                        <span className="block truncate">
                          {member.username}
                        </span>
                      </span>
                      {member.id === objectiveForm.leadUserId && (
                        <CheckIcon className="h-[1.1rem] w-auto" />
                      )}
                    </Menu.Item>
                  ))}
                </Menu.Group>
              </Menu.Items>
            </Menu>
          </Flex>
          <Divider className="my-4" />
          <Box>
            <Text className="mb-4 font-medium">Key Results (OKRs)</Text>
            {objectiveForm.keyResults.map((kr) => (
              <KeyResultEditor
                key={kr.id}
                keyResult={kr}
                onRemove={(id) => {
                  setObjectiveForm((prev) => ({
                    ...prev,
                    keyResults: prev.keyResults.filter((k) => k.id !== id),
                  }));
                }}
                onUpdate={(id, updates) => {
                  setObjectiveForm((prev) => ({
                    ...prev,
                    keyResults: prev.keyResults.map((k) =>
                      k.id === id ? { ...k, ...updates } : k,
                    ),
                  }));
                }}
              />
            ))}
            <Button
              className="mt-2"
              color="tertiary"
              leftIcon={<PlusIcon />}
              onClick={() => {
                setObjectiveForm((prev) => ({
                  ...prev,
                  keyResults: [
                    ...prev.keyResults,
                    {
                      id: crypto.randomUUID(),
                      name: "",
                      measureType: "Number",
                      startValue: 0,
                      targetValue: 0,
                    },
                  ],
                }));
              }}
              size="sm"
              variant="outline"
            >
              Add Key Result
            </Button>
          </Box>
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
            leftIcon={<PlusIcon className="text-white dark:text-gray-200" />}
            loading={loading}
            loadingText="Creating objective..."
            onClick={handleCreateObjective}
            size="md"
          >
            Create objective
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
