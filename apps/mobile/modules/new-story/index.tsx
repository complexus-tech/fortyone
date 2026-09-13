import React from "react";
import {
  KeyboardAvoidingView,
  ScrollView,
  TextInput,
  View,
} from "react-native";
import { SafeContainer, Text, Wrapper, Button } from "@/components/ui";
import { DiscardDraftButton } from "@/components/rich-text/draft-recovery";
import { colors } from "@/constants";
import { useTheme } from "@/hooks";
import { Header } from "./components/header";
import { DescriptionEditor } from "./components/description-editor";
import { MetadataRow } from "./components/metadata-row";
import {
  MetadataSheet,
  type MetadataOption,
} from "./components/metadata-sheet";
import { PRIORITIES, type SheetName } from "./form-state";
import { useNewStoryForm } from "./hooks/use-new-story-form";

type SheetConfig = {
  title: string;
  options: MetadataOption[];
  selectedIds: string[];
  emptyText: string;
  multiple?: boolean;
};

export const NewStory = () => {
  const { resolvedTheme } = useTheme();
  const {
    state: {
      title,
      description,
      statusId,
      assigneeId,
      priority,
      labelIds,
      activeSheet,
    },
    draft,
    teams,
    statuses,
    members,
    labels,
    selectedTeamId,
    selectedTeam,
    selectedStatus,
    selectedAssignee,
    selectedLabels,
    canSubmit,
    isSubmitting,
    submitError,
    submit,
    setTitle,
    setDescription,
    openSheet,
    closeSheet,
    selectMetadata,
  } = useNewStoryForm();
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
      <Header disabled={!canSubmit} loading={isSubmitting} onSubmit={submit} />
      <KeyboardAvoidingView behavior="padding" style={{ flex: 1 }}>
        <ScrollView
          keyboardShouldPersistTaps="handled"
          showsVerticalScrollIndicator={false}
          contentContainerStyle={{ gap: 18, paddingBottom: 28 }}
        >
          <View>
            {draft.error && (
              <Text color="danger" accessibilityRole="alert">
                {draft.error}
              </Text>
            )}
            {draft.error && !draft.ready && (
              <DiscardDraftButton onDiscard={draft.reset} />
            )}
            {submitError && (
              <Text color="danger" accessibilityRole="alert">
                {submitError}
              </Text>
            )}
            {!draft.ready && <Text color="muted">Restoring your draft…</Text>}
            <TextInput
              accessibilityLabel="Task title"
              editable={draft.ready && !isSubmitting}
              value={title}
              onChangeText={setTitle}
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
              disabled={!draft.ready || isSubmitting}
              onChange={setDescription}
            />
          </View>

          <Wrapper className="border-0 bg-gray-100/60 py-1 dark:bg-dark-100/45">
            <MetadataRow
              required
              label="Team"
              value={selectedTeam?.name ?? "Choose team"}
              onPress={() => openSheet("team")}
            />
            <MetadataRow
              label="Status"
              value={selectedStatus?.name ?? "No status"}
              onPress={() => openSheet("status")}
            />
            <MetadataRow
              label="Priority"
              value={priority}
              onPress={() => openSheet("priority")}
            />
            <MetadataRow
              label="Assignee"
              value={
                selectedAssignee?.fullName ||
                selectedAssignee?.username ||
                "Unassigned"
              }
              onPress={() => openSheet("assignee")}
            />
            <MetadataRow
              label="Labels"
              value={
                selectedLabels.length > 0
                  ? selectedLabels.map((label) => label.name).join(", ")
                  : "None"
              }
              onPress={() => openSheet("labels")}
            />
          </Wrapper>

          <Button
            size="lg"
            rounded="lg"
            color="invert"
            loading={isSubmitting}
            disabled={!canSubmit}
            onPress={submit}
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
          onSelect={(option) => selectMetadata(option.id)}
        />
      ) : null}
    </SafeContainer>
  );
};
