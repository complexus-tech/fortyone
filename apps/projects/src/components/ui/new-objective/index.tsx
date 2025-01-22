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
import { AssigneesMenu } from "@/components/ui/story/assignees-menu";
import { TeamColor } from "../team-color";
import { KeyResultsList } from "./components/key-results-list";
import { KeyResultEditor } from "./components/key-result-editor";
import type { KeyResult, NewObjective, ObjectiveStatus } from "./types";

type KeyResultFormMode = "add" | "edit" | null;

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
    startDate: null,
    endDate: null,
    leadUserId: session.data?.user?.id || null,
    keyResults: [],
  };

  const [objectiveForm, setObjectiveForm] = useState<NewObjective>(initialForm);
  const [loading, setLoading] = useState(false);
  const [keyResultMode, setKeyResultMode] = useState<KeyResultFormMode>(null);
  const [editingKeyResult, setEditingKeyResult] = useState<KeyResult | null>(
    null,
  );

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
      setObjectiveForm(initialForm);
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
                assigneeId={objectiveForm.leadUserId}
                onAssigneeSelected={(leadUserId) => {
                  setObjectiveForm((prev) => ({
                    ...prev,
                    leadUserId,
                  }));
                }}
              />
            </AssigneesMenu>
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
          </Flex>
          <Divider className="my-4" />
          <Box>
            <Text className="mb-4 font-medium">Objective Key Results</Text>
            {keyResultMode === null ? (
              <>
                <KeyResultsList
                  keyResults={objectiveForm.keyResults}
                  onEdit={(id) => {
                    const kr = objectiveForm.keyResults.find(
                      (k) => k.id === id,
                    );
                    if (kr) {
                      setEditingKeyResult(kr);
                      setKeyResultMode("edit");
                    }
                  }}
                  onRemove={(id) => {
                    setObjectiveForm((prev) => ({
                      ...prev,
                      keyResults: prev.keyResults.filter((k) => k.id !== id),
                    }));
                  }}
                />
                <Button
                  className="mt-4"
                  color="tertiary"
                  leftIcon={<PlusIcon />}
                  onClick={() => {
                    const newKr = {
                      id: crypto.randomUUID(),
                      name: "",
                      measureType: "Number" as const,
                      startValue: 0,
                      targetValue: 0,
                    };
                    setEditingKeyResult(newKr);
                    setKeyResultMode("add");
                  }}
                  size="sm"
                  variant="outline"
                >
                  Add Key Result
                </Button>
              </>
            ) : (
              <KeyResultEditor
                keyResult={editingKeyResult!}
                onCancel={() => {
                  setKeyResultMode(null);
                  setEditingKeyResult(null);
                }}
                onSave={() => {
                  if (keyResultMode === "add") {
                    setObjectiveForm((prev) => ({
                      ...prev,
                      keyResults: [...prev.keyResults, editingKeyResult!],
                    }));
                  } else {
                    setObjectiveForm((prev) => ({
                      ...prev,
                      keyResults: prev.keyResults.map((kr) =>
                        kr.id === editingKeyResult!.id ? editingKeyResult! : kr,
                      ),
                    }));
                  }
                  setKeyResultMode(null);
                  setEditingKeyResult(null);
                }}
                onUpdate={(id, updates) => {
                  setEditingKeyResult((prev) => ({
                    ...prev!,
                    ...updates,
                  }));
                }}
              />
            )}
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
            disabled={editingKeyResult !== null}
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
