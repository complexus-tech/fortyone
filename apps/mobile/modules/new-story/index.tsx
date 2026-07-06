import React, { useReducer } from "react";
import { KeyboardAvoidingView, ScrollView, TextInput, View } from "react-native";
import { useRouter } from "expo-router";
import { SafeContainer, Text, Wrapper, Button } from "@/components/ui";
import { Header } from "./components/header";
import { DescriptionEditor } from "./components/description-editor";
import { MetadataRow } from "./components/metadata-row";
import {
  MetadataOption,
  MetadataSheet,
} from "./components/metadata-sheet";
import { descriptionToHtml } from "./utils/description";
import { useCreateStoryMutation } from "./hooks/use-create-story-mutation";
import { useTeams } from "@/modules/teams/hooks/use-teams";
import { useTeamStatuses } from "@/modules/statuses/hooks/use-statuses";
import { useTeamMembers } from "@/modules/members/hooks/use-team-members";
import { useTeamLabels } from "@/modules/labels/hooks/use-labels";
import { StoryPriority } from "@/modules/stories/types";
import { colors } from "@/constants";
import { useTheme } from "@/hooks";

type SheetName = "team" | "status" | "priority" | "assignee" | "labels";
type SheetConfig = {
  title: string;
  options: MetadataOption[];
  selectedIds: string[];
  emptyText: string;
  multiple?: boolean;
};

const PRIORITIES: StoryPriority[] = [
  "No Priority",
  "Low",
  "Medium",
  "High",
  "Urgent",
];

type FormState = {
  title: string;
  description: string;
  teamId: string;
  statusId: string | null;
  assigneeId: string | null;
  priority: StoryPriority;
  labelIds: string[];
  activeSheet: SheetName | null;
};

type FormAction =
  | { type: "setTitle"; title: string }
  | { type: "setDescription"; description: string }
  | { type: "setTeam"; teamId: string }
  | { type: "setStatus"; statusId: string }
  | { type: "setAssignee"; assigneeId: string }
  | { type: "setPriority"; priority: StoryPriority }
  | { type: "toggleLabel"; labelId: string }
  | { type: "setSheet"; sheet: SheetName | null };

const initialState: FormState = {
  title: "",
  description: "",
  teamId: "",
  statusId: null,
  assigneeId: null,
  priority: "No Priority",
  labelIds: [],
  activeSheet: null,
};

const formReducer = (state: FormState, action: FormAction): FormState => {
  switch (action.type) {
    case "setTitle":
      return { ...state, title: action.title };
    case "setDescription":
      return { ...state, description: action.description };
    case "setTeam":
      return {
        ...state,
        teamId: action.teamId,
        statusId: null,
        assigneeId: null,
        labelIds: [],
        activeSheet: null,
      };
    case "setStatus":
      return { ...state, statusId: action.statusId, activeSheet: null };
    case "setAssignee":
      return { ...state, assigneeId: action.assigneeId, activeSheet: null };
    case "setPriority":
      return { ...state, priority: action.priority, activeSheet: null };
    case "toggleLabel":
      return {
        ...state,
        labelIds: state.labelIds.includes(action.labelId)
          ? state.labelIds.filter((id) => id !== action.labelId)
          : [...state.labelIds, action.labelId],
      };
    case "setSheet":
      return { ...state, activeSheet: action.sheet };
  }
};

export const NewStory = () => {
  const router = useRouter();
  const { resolvedTheme } = useTheme();
  const createStoryMutation = useCreateStoryMutation();
  const { data: teams = [] } = useTeams();
  const [state, dispatch] = useReducer(formReducer, initialState);
  const {
    title,
    description,
    teamId,
    statusId,
    assigneeId,
    priority,
    labelIds,
    activeSheet,
  } = state;

  const selectedTeamId = teamId || teams[0]?.id || "";
  const selectedTeam = teams.find((team) => team.id === selectedTeamId);
  const { data: statuses = [] } = useTeamStatuses(selectedTeamId);
  const { data: members = [] } = useTeamMembers(selectedTeamId);
  const { data: labels = [] } = useTeamLabels(selectedTeamId);

  const selectedStatus = statuses.find((status) => status.id === statusId);
  const selectedAssignee = members.find((member) => member.id === assigneeId);
  const selectedLabels = labels.filter((label) => labelIds.includes(label.id));

  const titleColor = resolvedTheme === "light" ? colors.black : colors.white;
  const placeholderColor =
    resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[300];

  const teamOptions: MetadataOption[] = teams.map((team) => ({
    id: team.id,
    label: team.name,
    description: team.code,
    color: team.color,
  }));
  const statusOptions: MetadataOption[] = statuses.map((status) => ({
    id: status.id,
    label: status.name,
    description: status.category,
    color: status.color,
  }));
  const assigneeOptions: MetadataOption[] = members.map((member) => ({
    id: member.id,
    label: member.fullName || member.username || member.email,
    description: member.email,
  }));
  const priorityOptions: MetadataOption[] = PRIORITIES.map((item) => ({
    id: item,
    label: item,
  }));
  const labelOptions: MetadataOption[] = labels.map((label) => ({
    id: label.id,
    label: label.name,
    color: label.color,
  }));

  const canSubmit = title.trim().length > 0 && Boolean(selectedTeamId);

  const closeSheet = () => dispatch({ type: "setSheet", sheet: null });

  const handleSubmit = () => {
    if (!canSubmit || createStoryMutation.isPending) {
      return;
    }

    const trimmedDescription = description.trim();
    createStoryMutation.mutate(
      {
        title: title.trim(),
        description: trimmedDescription || undefined,
        descriptionHTML: trimmedDescription
          ? descriptionToHtml(trimmedDescription)
          : undefined,
        teamId: selectedTeamId,
        statusId,
        assigneeId,
        priority,
        labelIds,
      },
      {
        onSuccess: (response) => {
          if (response.data?.id) {
            router.replace(`/story/${response.data.id}`);
          }
        },
      }
    );
  };

  const handleSheetSelect = (option: MetadataOption) => {
    switch (activeSheet) {
      case "team":
        dispatch({ type: "setTeam", teamId: option.id });
        break;
      case "status":
        dispatch({ type: "setStatus", statusId: option.id });
        break;
      case "priority":
        dispatch({ type: "setPriority", priority: option.id as StoryPriority });
        break;
      case "assignee":
        dispatch({ type: "setAssignee", assigneeId: option.id });
        break;
      case "labels":
        dispatch({ type: "toggleLabel", labelId: option.id });
        break;
    }
  };

  const sheetConfig: Record<SheetName, SheetConfig> = {
    team: {
      title: "Team",
      options: teamOptions,
      selectedIds: selectedTeamId ? [selectedTeamId] : [],
      emptyText: "No teams available",
    },
    status: {
      title: "Status",
      options: statusOptions,
      selectedIds: statusId ? [statusId] : [],
      emptyText: "No statuses available",
    },
    priority: {
      title: "Priority",
      options: priorityOptions,
      selectedIds: [priority],
      emptyText: "No priorities available",
    },
    assignee: {
      title: "Assignee",
      options: assigneeOptions,
      selectedIds: assigneeId ? [assigneeId] : [],
      emptyText: "No members available",
    },
    labels: {
      title: "Labels",
      options: labelOptions,
      selectedIds: labelIds,
      emptyText: "No labels available",
      multiple: true,
    },
  };

  const currentSheet = activeSheet ? sheetConfig[activeSheet] : null;

  return (
    <SafeContainer edges={["top", "bottom"]}>
      <Header
        disabled={!canSubmit}
        loading={createStoryMutation.isPending}
        onSubmit={handleSubmit}
      />
      <KeyboardAvoidingView behavior="padding" style={{ flex: 1 }}>
        <ScrollView
          keyboardShouldPersistTaps="handled"
          showsVerticalScrollIndicator={false}
          contentContainerStyle={{ gap: 18, paddingBottom: 28 }}
        >
          <View>
            <TextInput
              value={title}
              onChangeText={(nextTitle) =>
                dispatch({ type: "setTitle", title: nextTitle })
              }
              placeholder="Task title"
              placeholderTextColor={placeholderColor}
              autoFocus
              multiline
              style={{
                color: titleColor,
                fontSize: 34,
                lineHeight: 39,
                fontWeight: "700",
                padding: 0,
                minHeight: 84,
              }}
            />
          </View>

          <View style={{ gap: 8 }}>
            <Text color="muted" fontSize="sm">
              Description
            </Text>
            <DescriptionEditor
              value={description}
              onChangeText={(nextDescription) =>
                dispatch({
                  type: "setDescription",
                  description: nextDescription,
                })
              }
            />
          </View>

          <Wrapper className="border-0 bg-gray-100/60 py-1 dark:bg-dark-100/45">
            <MetadataRow
              required
              label="Team"
              value={selectedTeam?.name ?? "Choose team"}
              onPress={() => dispatch({ type: "setSheet", sheet: "team" })}
            />
            <MetadataRow
              label="Status"
              value={selectedStatus?.name ?? "No status"}
              onPress={() => dispatch({ type: "setSheet", sheet: "status" })}
            />
            <MetadataRow
              label="Priority"
              value={priority}
              onPress={() => dispatch({ type: "setSheet", sheet: "priority" })}
            />
            <MetadataRow
              label="Assignee"
              value={
                selectedAssignee?.fullName ||
                selectedAssignee?.username ||
                "Unassigned"
              }
              onPress={() => dispatch({ type: "setSheet", sheet: "assignee" })}
            />
            <MetadataRow
              label="Labels"
              value={
                selectedLabels.length > 0
                  ? selectedLabels.map((label) => label.name).join(", ")
                  : "None"
              }
              onPress={() => dispatch({ type: "setSheet", sheet: "labels" })}
            />
          </Wrapper>

          <Button
            size="lg"
            rounded="lg"
            color="invert"
            loading={createStoryMutation.isPending}
            disabled={!canSubmit}
            onPress={handleSubmit}
          >
            <Text>Create task</Text>
          </Button>
        </ScrollView>
      </KeyboardAvoidingView>

      {currentSheet ? (
        <MetadataSheet
          isOpen={Boolean(activeSheet)}
          title={currentSheet.title}
          options={currentSheet.options}
          selectedIds={currentSheet.selectedIds}
          multiple={currentSheet.multiple}
          emptyText={currentSheet.emptyText}
          onClose={closeSheet}
          onSelect={handleSheetSelect}
        />
      ) : null}
    </SafeContainer>
  );
};
