"use client";
import {
  useEffect,
  useRef,
  useState,
  type Dispatch,
  type SetStateAction,
} from "react";
import { useEditor } from "@tiptap/react";
import Underline from "@tiptap/extension-underline";
import Link from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import TextExt from "@tiptap/extension-text";
import { marked } from "marked";
import { toast } from "sonner";
import { createRichTextStarterKit } from "@/lib/tiptap/starter-kit";
import { useFeatures, useLocalStorage, useTerminology } from "@/hooks";
import type { Team } from "@/modules/teams/types";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useMembers } from "@/lib/hooks/members";
import type { NewObjective } from "@/modules/objectives/types";
import { useCreateObjectiveMutation } from "@/modules/objectives/hooks";
import { useObjectiveStatuses } from "@/lib/hooks/objective-statuses";
import {
  useObjectives,
  useTeamObjectives,
} from "@/modules/objectives/hooks/use-objectives";
import { FeatureGuard } from "@/components/ui/feature-guard";
import { OkrQualityBanner } from "@/modules/objectives/components/okr-quality-banner";
import { useOkrQualityAssessment } from "@/modules/objectives/hooks/use-okr-quality-assessment";
import { getAvailableObjectiveColor } from "./color-utils";
import { NewObjectiveDialogContent } from "./new-objective-dialog-content";
import { ObjectivePlanLimitDialog } from "./objective-plan-limit-dialog";
import { useNewObjectiveKeyResults } from "./use-new-objective-key-results";

export const NewObjectiveDialog = ({
  description,
  isOpen,
  setIsOpen,
  teamId: initialTeamId,
}: {
  description?: string;
  isOpen: boolean;
  setIsOpen: Dispatch<SetStateAction<boolean>>;
  teamId?: string;
}) => {
  const { data: teams = [] } = useTeams();
  const { data: members = [] } = useMembers();
  const { data: statuses = [] } = useObjectiveStatuses();
  const { data: objectives = [] } = useObjectives();
  const defaultObjectiveColor = getAvailableObjectiveColor(
    objectives.map(({ color }) => color),
  );
  const { getTermDisplay } = useTerminology();
  const objectiveTerm = getTermDisplay("objectiveTerm");
  const objectiveTermCapitalized = getTermDisplay("objectiveTerm", {
    capitalize: true,
  });
  const keyResultTerm = getTermDisplay("keyResultTerm", {
    capitalize: true,
  });
  const keyResultsTerm = getTermDisplay("keyResultTerm", {
    capitalize: true,
    variant: "plural",
  });
  const [isExpanded, setIsExpanded] = useState(false);
  const firstTeam = teams.length > 0 ? teams[0] : null;
  const [activeTeam, setActiveTeam] = useLocalStorage<Team | null>(
    "activeTeam",
    firstTeam,
  );

  const validActiveTeam =
    teams.find((team) => team.id === activeTeam?.id) || firstTeam;

  const currentTeamId = initialTeamId || validActiveTeam?.id;
  const currentTeam =
    teams.find((team) => team.id === currentTeamId) || firstTeam;
  const defaultStatus =
    statuses.find((status) => status.isDefault) || statuses[0];

  useEffect(() => {
    if (!teams.find((team) => team.id === activeTeam?.id)) {
      setActiveTeam(firstTeam);
    }
  }, [teams, activeTeam, setActiveTeam, firstTeam]);

  const initialForm: NewObjective = {
    name: "",
    description: "",
    shortSummary: "",
    leadUser: null,
    teamId: currentTeamId || "",
    startDate: null,
    endDate: null,
    statusId: statuses.length > 0 ? defaultStatus.id : "",
    priority: "No Priority",
    color: defaultObjectiveColor,
    keyResults: [],
  };
  const features = useFeatures();
  const [objectiveForm, setObjectiveForm] = useState<NewObjective>(initialForm);
  const {
    editingIndex,
    editingKeyResult,
    handleAddKeyResult,
    handleEditKeyResult,
    handleKeyResultUpdate,
    handleRemoveKeyResult,
    handleSaveKeyResult,
    isEditingKeyResult,
    resetKeyResultEditor,
  } = useNewObjectiveKeyResults({
    keyResults: objectiveForm.keyResults ?? [],
    onKeyResultsChange: (updateKeyResults) => {
      setObjectiveForm((current) => ({
        ...current,
        keyResults: updateKeyResults(current.keyResults ?? []),
      }));
    },
  });
  const [objectiveName, setObjectiveName] = useState("");
  const { data: teamObjectives = [] } = useTeamObjectives(
    currentTeam?.id ?? "",
  );
  const createMutation = useCreateObjectiveMutation();
  const objectiveQualityRequest = objectiveName.trim()
    ? {
        kind: "objective" as const,
        draft: {
          name: objectiveName,
          summary: objectiveForm.shortSummary ?? "",
          startDate: objectiveForm.startDate ?? null,
          endDate: objectiveForm.endDate ?? null,
        },
        existingObjectives: teamObjectives.map((objective) => ({
          id: objective.id,
          name: objective.name,
        })),
      }
    : null;
  const { assessment: objectiveQuality, isAssessing: isAssessingObjective } =
    useOkrQualityAssessment(objectiveQualityRequest);

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
    onUpdate: ({ editor: currentEditor }) => {
      setObjectiveName(currentEditor.getText());
    },
  });

  const editor = useEditor({
    extensions: [
      createRichTextStarterKit(),
      Underline,
      Link.configure({
        autolink: true,
      }),
      Placeholder.configure({ placeholder: "Add description..." }),
    ],
    content: marked.parse(description || "", { gfm: true }),
    editable: true,
    immediatelyRender: false,
  });
  const wasOpenRef = useRef(false);

  useEffect(() => {
    if (isOpen && !wasOpenRef.current) {
      setObjectiveForm((current) => ({
        ...current,
        color: defaultObjectiveColor,
      }));
      if (editor) {
        editor.commands.setContent(
          marked.parse(description || "", { gfm: true }),
        );
      }
    }
    wasOpenRef.current = isOpen;
  }, [defaultObjectiveColor, description, editor, isOpen]);

  const handleCreateObjective = () => {
    if (!titleEditor || !editor) return;
    if (!titleEditor.getText()) {
      titleEditor.commands.focus();
      toast.warning("Validation Error", {
        description: "Title is required",
      });
      return;
    }

    if (!objectiveForm.startDate || !objectiveForm.endDate) {
      toast.warning("Validation Error", {
        description: "Start date and deadline are required",
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
        description: `${objectiveTermCapitalized} with this name already exists`,
      });
      return;
    }

    createMutation.mutate({
      ...objectiveForm,
      name: titleEditor.getText(),
      description: editor.getHTML(),
    });
    setIsOpen(false);
    setIsExpanded(false);
    titleEditor.commands.setContent("");
    setObjectiveName("");
    editor.commands.setContent("");
    setObjectiveForm(initialForm);
  };

  useEffect(() => {
    if (isOpen && titleEditor) {
      titleEditor.commands.focus();
    }
  }, [isOpen, titleEditor]);

  const lead = members.find((member) => member.id === objectiveForm.leadUser);

  const updateObjectiveForm = (updates: Partial<NewObjective>) => {
    setObjectiveForm((current) => ({ ...current, ...updates }));
  };

  const handleTeamSelect = (teamId: string) => {
    const team = teams.find((candidate) => candidate.id === teamId);
    if (!team) return;

    setActiveTeam(team);
    updateObjectiveForm({ teamId: team.id });
  };

  useEffect(() => {
    if (isOpen && teams.length === 0) {
      toast.warning("Join or create a team", {
        description:
          "You need to be part of a team to create an objective. Open Team Settings to join or create one.",
      });
      setIsOpen(false);
    }
  }, [isOpen, teams, setIsOpen]);

  return (
    <FeatureGuard
      count={objectives.length}
      fallback={
        <ObjectivePlanLimitDialog
          isOpen={isOpen}
          objectiveCount={objectives.length}
          setIsOpen={setIsOpen}
        />
      }
      feature="maxObjectives"
    >
      <NewObjectiveDialogContent
        currentTeam={currentTeam}
        descriptionEditor={editor}
        form={objectiveForm}
        isCreating={createMutation.isPending}
        isExpanded={isExpanded}
        isOpen={isOpen}
        keyResults={
          features.keyResultEnabled
            ? {
                editingKeyResult,
                existingKeyResults: (objectiveForm.keyResults ?? []).filter(
                  (_keyResult, index) => index !== editingIndex,
                ),
                isEditing: isEditingKeyResult,
                keyResultTerm,
                keyResults: objectiveForm.keyResults ?? [],
                keyResultsTerm,
                objectiveEndDate: objectiveForm.endDate ?? null,
                objectiveName,
                objectiveStartDate: objectiveForm.startDate ?? null,
                onAdd: handleAddKeyResult,
                onCancel: resetKeyResultEditor,
                onEdit: handleEditKeyResult,
                onRemove: handleRemoveKeyResult,
                onSave: handleSaveKeyResult,
                onUpdate: handleKeyResultUpdate,
              }
            : undefined
        }
        lead={lead}
        newObjectiveTerm={objectiveTerm}
        objectiveTerm={objectiveTermCapitalized}
        onCreate={handleCreateObjective}
        onDiscard={() => {
          setIsOpen(false);
        }}
        onObjectiveFormChange={updateObjectiveForm}
        onOpenChange={setIsOpen}
        onTeamSelect={handleTeamSelect}
        onToggleExpanded={() => {
          setIsExpanded((current) => !current);
        }}
        qualityBanner={
          <OkrQualityBanner
            assessment={objectiveQuality}
            isAssessing={isAssessingObjective}
            onUseSuggestion={(suggestion) => {
              titleEditor?.commands.setContent(suggestion);
              setObjectiveName(suggestion);
            }}
          />
        }
        statuses={statuses}
        teams={teams}
        titleEditor={titleEditor}
      />
    </FeatureGuard>
  );
};
