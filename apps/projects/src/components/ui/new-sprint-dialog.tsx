"use client";
import { useEffect, useState, type Dispatch, type SetStateAction } from "react";
import { Button, Dialog, Flex, Text, TextEditor, Menu, DatePicker } from "ui";
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
  CheckIcon,
  MaximizeIcon,
  MinimizeIcon,
  PlusIcon,
  ObjectiveIcon,
  CalendarIcon,
  CloseIcon,
} from "icons";
import { toast } from "sonner";
import { cn } from "lib";
import { format } from "date-fns";
import { useFeatures, useLocalStorage, useTerminology } from "@/hooks";
import type { Team } from "@/modules/teams/types";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useTeamObjectives } from "@/modules/objectives/hooks/use-objectives";
import type { NewSprint } from "@/modules/sprints/types";
import { useCreateSprintMutation } from "@/modules/sprints/hooks/create-sprint-mutation";
import { useTeamSprints } from "@/modules/sprints/hooks/team-sprints";
import { validateSprintDates } from "@/modules/sprints/utils/validate-sprint-dates";
import { TeamColor } from "./team-color";
import { ObjectivesMenu } from "./story/objectives-menu";

export const NewSprintDialog = ({
  isOpen,
  setIsOpen,
  teamId: initialTeamId,
}: {
  isOpen: boolean;
  setIsOpen: Dispatch<SetStateAction<boolean>>;
  teamId?: string;
}) => {
  const { data: teams = [] } = useTeams();
  const { getTermDisplay } = useTerminology();

  const [isExpanded, setIsExpanded] = useState(false);
  const firstTeam = teams.length > 0 ? teams[0] : null;
  const [activeTeam, setActiveTeam] = useLocalStorage<Team | null>(
    "activeTeam",
    firstTeam,
  );

  const initialForm: NewSprint = {
    name: "",
    goal: "",
    teamId: initialTeamId || activeTeam?.id || "",
    objectiveId: null,
    startDate: "",
    endDate: "",
  };
  const features = useFeatures();
  const [sprintForm, setSprintForm] = useState<NewSprint>(initialForm);
  const { data: teamSprints = [] } = useTeamSprints(sprintForm.teamId);
  const { data: objectives = [] } = useTeamObjectives(sprintForm.teamId);

  const titleEditor = useEditor({
    extensions: [
      Document,
      Paragraph,
      TextExt,
      Placeholder.configure({
        placeholder: `${getTermDisplay("sprintTerm", { capitalize: true })} name eg. ${getTermDisplay("sprintTerm", { capitalize: true })} 1`,
      }),
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
      Link.configure({
        autolink: true,
      }),
      Placeholder.configure({
        placeholder: `What does the team want to accomplish in this ${getTermDisplay("sprintTerm")}?`,
      }),
    ],
    content: "",
    editable: true,
    immediatelyRender: false,
  });

  const { mutateAsync } = useCreateSprintMutation();

  const handleCreateSprint = () => {
    if (!titleEditor || !editor) return;
    if (!titleEditor.getText()) {
      titleEditor.commands.focus();
      toast.warning("Validation Error", {
        description: `${getTermDisplay("sprintTerm", { capitalize: true })} name is required`,
      });
      return;
    }
    if (!sprintForm.startDate || !sprintForm.endDate) {
      toast.warning("Validation Error", {
        description: "Start date and deadline are required",
      });
      return;
    }

    const dateValidation = validateSprintDates(
      sprintForm.startDate,
      sprintForm.endDate,
      teamSprints,
    );

    if (!dateValidation.isValid) {
      toast.warning("Validation Error", {
        description: dateValidation.error,
      });
      return;
    }

    mutateAsync({
      ...sprintForm,
      name: titleEditor.getText(),
      goal: editor.getHTML(),
    }).then(() => {
      setIsOpen(false);
      setIsExpanded(false);
      titleEditor.commands.setContent("");
      editor.commands.setContent("");
      setSprintForm(initialForm);
    });
  };

  useEffect(() => {
    if (initialTeamId) {
      const team = teams.find((team) => team.id === initialTeamId);
      if (team) {
        setActiveTeam(team);
      }
    }
  }, [isOpen, initialTeamId, teams, setActiveTeam, titleEditor]);

  useEffect(() => {
    if (isOpen && titleEditor) {
      titleEditor.commands.focus();
    }
  }, [isOpen, titleEditor]);

  const objective = objectives.find((o) => o.id === sprintForm.objectiveId);

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
                  leftIcon={<TeamColor color={activeTeam?.color} />}
                  size="xs"
                >
                  {activeTeam?.code}
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
                        setSprintForm((prev) => ({
                          ...prev,
                          teamId: team.id,
                        }));
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
            <ArrowRightIcon className="h-4 w-auto opacity-30" strokeWidth={3} />
            <Text className="opacity-80" color="muted">
              New {getTermDisplay("sprintTerm")}
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
                    sprintForm.startDate ? (
                      <CloseIcon
                        aria-label="Remove date"
                        className="h-4 w-auto"
                        onClick={(e) => {
                          e.stopPropagation();
                          setSprintForm((prev) => ({
                            ...prev,
                            startDate: "",
                          }));
                        }}
                        role="button"
                      />
                    ) : null
                  }
                  size="sm"
                  variant="outline"
                >
                  {sprintForm.startDate
                    ? format(new Date(sprintForm.startDate), "MMM d, yyyy")
                    : "Start date"}
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar
                onDayClick={(date) => {
                  setSprintForm((prev) => ({
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
                    sprintForm.endDate ? (
                      <CloseIcon
                        aria-label="Remove date"
                        className="h-4 w-auto"
                        onClick={(e) => {
                          e.stopPropagation();
                          setSprintForm((prev) => ({
                            ...prev,
                            endDate: "",
                          }));
                        }}
                        role="button"
                      />
                    ) : null
                  }
                  size="sm"
                  variant="outline"
                >
                  {sprintForm.endDate
                    ? format(new Date(sprintForm.endDate), "MMM d, yyyy")
                    : "Deadline"}
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar
                fromDate={
                  sprintForm.startDate
                    ? new Date(sprintForm.startDate)
                    : undefined
                }
                onDayClick={(date) => {
                  setSprintForm((prev) => ({
                    ...prev,
                    endDate: date.toISOString(),
                  }));
                }}
              />
            </DatePicker>
            {features.objectiveEnabled && objectives.length > 0 ? (
              <ObjectivesMenu>
                <ObjectivesMenu.Trigger>
                  <Button
                    className="gap-1 px-2 text-sm"
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
                  objectiveId={sprintForm.objectiveId ?? undefined}
                  setObjectiveId={(objectiveId) => {
                    setSprintForm({ ...sprintForm, objectiveId });
                  }}
                  teamId={sprintForm.teamId}
                />
              </ObjectivesMenu>
            ) : null}
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
          <Button
            leftIcon={<PlusIcon className="text-white dark:text-gray-200" />}
            onClick={handleCreateSprint}
            size="md"
          >
            Create {getTermDisplay("sprintTerm")}
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
