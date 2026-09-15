import { useEffect, useRef, useState } from "react";
import { useNavigation, useRouter } from "expo-router";
import {
  usePreventRemove,
  type NavigationAction,
} from "expo-router/react-navigation";
import type { RichTextValue } from "@/components/rich-text/content";
import type { EditorMetadata } from "@/components/rich-text/metadata-icon";
import {
  RichTextEditorSurface,
  type RichTextEditorHandle,
} from "@/components/rich-text/editor";
import {
  KeyboardAvoidingView,
  Keyboard,
  Alert,
  TextInput,
  View,
} from "react-native";
import { SafeContainer, Text } from "@/components/ui";
import { DiscardDraftButton } from "@/components/rich-text/draft-recovery";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks";
import { useAuthStore } from "@/store/auth";
import { Header } from "./components/header";
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

type ComposerAction = {
  operation: { current: boolean };
  setProcessing: (processing: boolean) => void;
  setError: (message: string | null) => void;
  run: () => Promise<void>;
  recover: () => void;
  fallbackError: string;
};

async function runComposerAction(action: ComposerAction) {
  if (action.operation.current) return;
  action.operation.current = true;
  action.setProcessing(true);
  action.setError(null);
  try {
    await action.run();
  } catch (cause) {
    action.setError(
      cause instanceof Error ? cause.message : action.fallbackError,
    );
    action.recover();
  } finally {
    action.operation.current = false;
    action.setProcessing(false);
  }
}

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
    isFinalized,
    canEditDraft,
    submitError,
    submit,
    setTitle,
    setDescription,
    openSheet,
    closeSheet,
    selectMetadata,
  } = useNewStoryForm();
  const titleColor = themeColors[resolvedTheme].foreground;
  const placeholderColor = themeColors[resolvedTheme].textMuted;

  const teamOptions: MetadataOption[] = teams.map((team) => ({
    id: team.id,
    label: team.name,
    description: team.code,
    color: team.color,
  }));
  const statusOptions: MetadataOption[] = statuses.map((status) => ({
    id: status.id,
    label: status.name,
    color: status.color,
    icon: { kind: "status", category: status.category, color: status.color },
  }));
  const assigneeOptions: MetadataOption[] = members.map((member) => ({
    id: member.id,
    label: member.fullName || member.username || member.email,
    description: member.email,
    icon: {
      kind: "assignee",
      name: member.fullName || member.username || member.email,
      src: member.avatarUrl,
    },
  }));
  const priorityOptions: MetadataOption[] = PRIORITIES.map((item) => ({
    id: item,
    label: item,
    icon: { kind: "priority", priority: item },
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
  const navigation = useNavigation();
  const router = useRouter();
  const sessionIdentity = useAuthStore((state) =>
    state.isAuthenticated && !state.isLoading
      ? `${state.userId}:${state.workspace}:${state.sessionEpoch}`
      : null,
  );
  const [openingSession] = useState(sessionIdentity);
  const editorRef = useRef<RichTextEditorHandle>(null);
  const lastSnapshot = useRef<RichTextValue | null>(null);
  const operation = useRef(false);
  const [processing, setProcessing] = useState(false);
  const [editorReady, setEditorReady] = useState(false);
  const [editorReadOnly, setEditorReadOnly] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);
  const [destination, setDestination] = useState<
    { action: NavigationAction } | { storyId: string } | null
  >(null);
  const busy = processing || isSubmitting;
  const metadataDisabled = !draft.ready || busy || isFinalized;
  const handleOpenSheet = (sheet: SheetName) => {
    if (operation.current || !canEditDraft()) return;
    Keyboard.dismiss();
    openSheet(sheet);
  };

  const handleSubmit = () =>
    runComposerAction({
      operation,
      setProcessing,
      setError: setActionError,
      run: async () => {
        const snapshot = canEditDraft()
          ? await editorRef.current?.flush()
          : lastSnapshot.current;
        if (!snapshot)
          throw new Error("The editor is still loading. Please try again.");
        lastSnapshot.current = snapshot;
        const storyId = await submit(snapshot);
        setDestination({ storyId });
      },
      recover: () => {
        if (canEditDraft()) editorRef.current?.resume();
      },
      fallbackError: "Could not save your task. Your draft is still here.",
    });

  const handleClose = async (action: NavigationAction) => {
    if (operation.current) return;
    if (isFinalized) {
      await handleSubmit();
      return;
    }
    if (!draft.ready) {
      setDestination({ action });
      return;
    }
    if (!editorReady) {
      Alert.alert(
        "Close with your saved draft?",
        "The editor is unavailable. Your last saved draft will be kept, but recent changes may be missing.",
        [
          { text: "Keep editing", style: "cancel" },
          {
            text: "Close with saved draft",
            onPress: () => setDestination({ action }),
          },
        ],
      );
      return;
    }
    await runComposerAction({
      operation,
      setProcessing,
      setError: setActionError,
      run: async () => {
        if (editorReadOnly) {
          // Unsupported HTML was never editable; preserve it while saving native fields.
          await draft.persist(draft.valueRef.current);
        } else {
          const snapshot = await editorRef.current?.flush();
          if (!snapshot)
            throw new Error("The editor is still loading. Please try again.");
        }
        setDestination({ action });
      },
      recover: () => editorRef.current?.resume(),
      fallbackError:
        "Could not save your draft. Keep this screen open and try again.",
    });
  };

  // Covers header Close, Android Back, and native sheet dismissal.
  usePreventRemove(
    destination === null &&
      sessionIdentity !== null &&
      sessionIdentity === openingSession,
    ({ data }) => {
      void handleClose(data.action);
    },
  );
  useEffect(() => {
    if (!destination) return;
    if ("storyId" in destination)
      router.replace(`/story/${destination.storyId}`);
    else navigation.dispatch(destination.action);
  }, [destination, navigation, router]);

  const metadata: EditorMetadata[] = [
    {
      key: "status",
      kind: "status",
      label: "Status",
      value: selectedStatus?.name ?? "No status",
      muted: !selectedStatus,
      color: selectedStatus?.color,
      category: selectedStatus?.category,
    },
    {
      key: "priority",
      kind: "priority",
      label: "Priority",
      value: priority === "No Priority" ? "Priority" : priority,
      accessibilityValue: priority,
      priority,
      muted: priority === "No Priority",
    },
    {
      key: "assignee",
      kind: "assignee",
      label: "Assignee",
      value:
        selectedAssignee?.fullName || selectedAssignee?.username || "Assignee",
      accessibilityValue:
        selectedAssignee?.fullName ||
        selectedAssignee?.username ||
        "Unassigned",
      muted: !selectedAssignee,
      avatar: selectedAssignee
        ? {
            name:
              selectedAssignee.fullName ||
              selectedAssignee.username ||
              selectedAssignee.email,
            src: selectedAssignee.avatarUrl,
          }
        : undefined,
    },
    {
      key: "labels",
      kind: "labels",
      label: "Labels",
      value:
        selectedLabels.length === 1
          ? selectedLabels[0].name
          : selectedLabels.length > 1
            ? `${selectedLabels.length} labels`
            : "Labels",
      accessibilityValue:
        selectedLabels.map((label) => label.name).join(", ") || "None",
      muted: selectedLabels.length === 0,
    },
  ];

  return (
    <SafeContainer
      isFull
      edges={["top", "bottom"]}
      style={{
        backgroundColor: themeColors[resolvedTheme].surface,
      }}
    >
      <Header
        disabled={
          !canSubmit || editorReadOnly || (!editorReady && !isFinalized)
        }
        loading={busy}
        onSubmit={() => {
          void handleSubmit();
        }}
        teamName={selectedTeam?.name}
        teamColor={selectedTeam?.color}
        metadataDisabled={metadataDisabled}
        onTeamPress={() => handleOpenSheet("team")}
      />
      <KeyboardAvoidingView behavior="padding" style={{ flex: 1, minWidth: 0 }}>
        <View
          style={{ paddingHorizontal: 20, paddingTop: 8, paddingBottom: 2 }}
        >
          {draft.error && (
            <Text color="danger" accessibilityRole="alert">
              {draft.error}
            </Text>
          )}
          {draft.error && !draft.ready && (
            <DiscardDraftButton onDiscard={draft.reset} />
          )}
          {(actionError || submitError) && (
            <Text color="danger" accessibilityRole="alert">
              {actionError || submitError}
            </Text>
          )}
          {!draft.ready && <Text color="muted">Restoring your draft…</Text>}
          <TextInput
            accessibilityLabel="Task title"
            editable={draft.ready && !busy && !isFinalized}
            value={title}
            onChangeText={setTitle}
            placeholder="Task title"
            placeholderTextColor={placeholderColor}
            autoFocus
            multiline
            style={{
              color: titleColor,
              fontSize: 24,
              lineHeight: 30,
              fontWeight: "600",
              padding: 0,
              minHeight: 44,
              maxHeight: 160,
              textAlignVertical: "top",
            }}
          />
        </View>
        {draft.ready && (
          <RichTextEditorSurface
            ref={editorRef}
            inline
            initialHtml={description.html}
            placeholder="Description…"
            disabled={busy || isFinalized}
            metadata={metadata}
            onMetadataPress={(key) => {
              if (
                key === "status" ||
                key === "priority" ||
                key === "assignee" ||
                key === "labels"
              )
                handleOpenSheet(key);
            }}
            onDraft={setDescription}
            onReadyChange={setEditorReady}
            onReadOnlyChange={setEditorReadOnly}
          />
        )}
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
