"use client";
import {
  useReducer,
  useRef,
  useState,
  type Dispatch,
  type SetStateAction,
} from "react";
import { useRouter } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import {
  useFeatures,
  useLocalStorage,
  useTerminology,
  useUserRole,
  useSprintsEnabled,
  useWorkspacePath,
} from "@/hooks";
import { useDebouncedCallback } from "@/hooks/debounce";
import type { Team } from "@/modules/teams/types";
import type { DetailedStory } from "@/modules/story/types";
import type { StoryPriority } from "@/modules/stories/types";
import { useCreateStoryMutation } from "@/modules/story/hooks/create-mutation";
import { useStoryDescriptionMedia } from "@/modules/story/hooks/use-story-description-media";
import { useStatuses } from "@/lib/hooks/statuses";
import { useLabels } from "@/lib/hooks/labels";
import { DEFAULT_ESTIMATE_SCHEME } from "@/lib/estimate";
import { useMayaAssignee, useMembers } from "@/lib/hooks/members";
import { useJoinedTeams } from "@/modules/teams/hooks/teams";
import { useTeamSettings } from "@/modules/teams/hooks/use-team-settings";
import { useTeamObjectives } from "@/modules/objectives/hooks/use-objectives";
import { useKeyResults } from "@/modules/objectives/hooks";
import { useTeamSprints } from "@/modules/sprints/hooks/team-sprints";
import { useAutomationPreferences } from "@/lib/hooks/users/preferences";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { useTotalStories } from "@/modules/stories/hooks/total-stories";
import { useLinkFigmaStory } from "@/lib/hooks/figma";
import type { FigmaArtifact } from "@/modules/settings/workspace/integrations/figma/types";
import { storyKeys } from "@/modules/stories/constants";
import { useSimilarStories } from "@/modules/search/hooks/use-similar-stories";
import { getStoryPath } from "@/shared/routing/story";
import { MINIMUM_SIMILARITY_TITLE_CHARACTERS } from "@/constants/similarity";
import { NewStoryDialogContent } from "./new-story-dialog-content";
import {
  createInitialNewStoryDialogForm,
  getInitialDeadlineSource,
  storyFormReducer,
  toDateOnly,
  type DeadlineSource,
} from "./new-story-dialog-form";
import { useNewStoryDialogCreation } from "./use-new-story-dialog-creation";
import { useNewStoryDialogEditors } from "./use-new-story-dialog-editors";
import { useNewStoryDialogInitialization } from "./use-new-story-dialog-initialization";
import {
  useMayaAutoScheduling,
  useNewStoryDialogLifecycle,
} from "./use-new-story-dialog-lifecycle";
import { NewStoryDialogLimitGuard } from "./new-story-dialog-limit-guard";
import {
  getNewStoryDialogFieldSelections,
  getSelectedNewStoryLabels,
} from "./new-story-dialog-selections";

type NewStoryDialogProps = {
  isOpen: boolean;
  setIsOpen: Dispatch<SetStateAction<boolean>>;
  statusId?: string;
  teamId?: string;
  objectiveId?: string;
  sprintId?: string;
  priority?: StoryPriority;
  assigneeId?: string | null;
  description?: string;
  onCreated?: (story: DetailedStory) => Promise<void> | void;
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
  onCreated,
}: NewStoryDialogProps) => {
  const router = useRouter();
  const session = useSession();
  const { userRole } = useUserRole();
  const queryClient = useQueryClient();
  const features = useFeatures();
  const { workspaceSlug, withWorkspace } = useWorkspacePath();
  const { data: teams = [] } = useJoinedTeams();
  const { data: statuses = [] } = useStatuses();
  const { data: members = [] } = useMembers();
  const { data: allLabels = [] } = useLabels();
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
  const sprintsEnabled = useSprintsEnabled(currentTeamId ?? "");
  const currentTeam =
    teams.find((team) => team.id === currentTeamId) || firstTeam;
  const { data: objectives = [] } = useTeamObjectives(currentTeamId ?? "");
  const { data: sprints = [] } = useTeamSprints(currentTeamId ?? "");
  const initialSprintEndDate = toDateOnly(
    sprints.find((candidate) => candidate.id === sprintId)?.endDate,
  );
  const initialDeadlineSource = getInitialDeadlineSource({
    sprintId,
    sprintEndDate: initialSprintEndDate,
  });
  const { data: teamSettings } = useTeamSettings(currentTeamId);
  const { data: automationPreferences } = useAutomationPreferences();
  const { tier, getLimit, hasFeature } = useSubscriptionFeatures();
  const { data: totalStories = 0 } = useTotalStories();

  const teamStatuses = statuses.filter(
    (status) => status.teamId === currentTeamId,
  );

  const defaultStatus =
    teamStatuses.find((status) => status.id === statusId) ||
    (teamStatuses.length > 0
      ? teamStatuses.find((status) => status.isDefault) || teamStatuses[0]
      : null);
  const estimateScheme =
    teamSettings?.estimationSettings.scheme ?? DEFAULT_ESTIMATE_SCHEME;
  const autoAssignSelf = automationPreferences?.autoAssignSelf ?? false;
  const canUseBackgroundMaya = hasFeature("backgroundMaya");
  const autoSchedulingDefaultEnabled =
    canUseBackgroundMaya && (automationPreferences?.autoScheduling ?? true);
  const { data: mayaAssignee, isLoading: isMayaAssigneeLoading } =
    useMayaAssignee(isOpen && canUseBackgroundMaya);
  const currentUserId = session.data?.user.id ?? null;
  const autoAssignedUserId = autoAssignSelf ? currentUserId : null;
  const getInitialForm = () =>
    createInitialNewStoryDialogForm({
      assigneeId,
      autoAssignedUserId,
      autoSchedulingDefaultEnabled,
      currentTeamId,
      endDate: initialSprintEndDate,
      objectiveId,
      priority,
      sprintId,
      statusId: defaultStatus?.id,
    });
  const initialForm = getInitialForm();

  const [storyForm, dispatch] = useReducer(storyFormReducer, initialForm);
  const deadlineSourceRef = useRef<DeadlineSource>(initialDeadlineSource);
  const [createMore, setCreateMore] = useState(false);
  const [storyTitle, setStoryTitle] = useState("");
  const [storySearchQuery, setStorySearchQuery] = useState("");
  const [figmaArtifacts, setFigmaArtifacts] = useState<FigmaArtifact[]>([]);
  const mutation = useCreateStoryMutation();
  const linkFigmaStory = useLinkFigmaStory();
  const {
    cancelStagedUploads,
    finalizeStagedMedia,
    handleMediaFiles,
    inputRef: mediaInputRef,
    openMediaPicker,
    resetForNextStory,
  } = useStoryDescriptionMedia();
  const { data: keyResults = [] } = useKeyResults(
    storyForm.objectiveId ?? "",
    Boolean(storyForm.objectiveId && storyForm.keyResultId),
  );
  const { isMayaAssigned, member, sprint, strategyLinkLabel } =
    getNewStoryDialogFieldSelections({
      currentTeamCode: currentTeam?.code,
      keyResults,
      mayaAssignee,
      members,
      objectives,
      sprints,
      storyForm,
    });
  const selectedLabels = getSelectedNewStoryLabels(
    allLabels,
    storyForm.labelIds,
  );
  const { callback: searchSimilarStories, cancel: cancelStorySearch } =
    useDebouncedCallback(setStorySearchQuery, 300);
  const similarStories = useSimilarStories({
    limit: 3,
    title:
      isOpen && storyTitle.trim().length >= MINIMUM_SIMILARITY_TITLE_CHARACTERS
        ? storySearchQuery
        : "",
    teamId: currentTeamId,
  });
  const similarStoryItems =
    storyTitle.trim() === storySearchQuery ? similarStories.data ?? [] : [];

  const storyTerm = getTermDisplay("storyTerm");
  const storyTermCapitalized = getTermDisplay("storyTerm", {
    capitalize: true,
  });
  const storyTermPlural = getTermDisplay("storyTerm", { variant: "plural" });
  const { descriptionEditor: editor, titleEditor } = useNewStoryDialogEditors({
    description,
    onMediaFiles: handleMediaFiles,
    onMediaRequest: openMediaPicker,
    onStoryTitleChange: setStoryTitle,
    storyTerm: storyTermCapitalized,
  });
  const { handleCreateStory, isCreating } = useNewStoryDialogCreation({
    createMore,
    currentTeamId,
    editor,
    figmaArtifacts,
    finalizeStagedMedia,
    isMayaAssigneeLoading,
    linkFigmaStory: linkFigmaStory.mutateAsync,
    mayaAssigneeId: mayaAssignee?.id,
    mutateStory: mutation.mutateAsync,
    onCreated,
    onDialogClose: () => {
      setIsOpen(false);
      setIsExpanded(false);
    },
    onFigmaArtifactsReset: () => {
      setFigmaArtifacts([]);
    },
    onFormReset: () => {
      dispatch({ type: "RESET_FORM", payload: getInitialForm() });
    },
    onFreeStoryCreated:
      tier === "free"
        ? () => {
            void queryClient.invalidateQueries({
              queryKey: storyKeys.total(workspaceSlug),
            });
          }
        : undefined,
    onResetDeadlineSource: () => {
      deadlineSourceRef.current = initialDeadlineSource;
    },
    resetStagedMedia: resetForNextStory,
    setStoryTitle,
    storyForm,
    storyTerm: storyTermCapitalized,
    titleEditor,
  });

  useNewStoryDialogLifecycle({
    activeTeamId: activeTeam?.id,
    cancelStagedUploads,
    cancelStorySearch,
    currentTeamId,
    dispatch,
    firstTeam,
    isOpen,
    searchSimilarStories,
    setActiveTeam,
    statusId,
    storyForm,
    storyTitle,
    teamStatuses,
    teams,
    titleEditor,
  });

  useNewStoryDialogInitialization({
    assigneeId,
    autoAssignedUserId,
    autoSchedulingDefaultEnabled,
    currentTeamId,
    deadlineSourceRef,
    defaultStatusId: defaultStatus?.id,
    dispatch,
    initialDeadlineSource,
    initialSprintEndDate,
    objectiveId,
    priority,
    sprintId,
  });

  useMayaAutoScheduling({
    dispatch,
    isMayaAssigned,
    isAutoSchedulingEnabled: storyForm.autoSchedulingEnabled,
  });

  return (
    <NewStoryDialogLimitGuard
      billingHref={withWorkspace("/settings/workspace/billing")}
      canUpgrade={userRole === "admin"}
      isOpen={isOpen}
      maxStories={getLimit("maxStories")}
      onClose={() => {
        setIsOpen(false);
      }}
      planName={tier.replace("free", "hobby")}
      totalStories={totalStories}
    >
      <NewStoryDialogContent
        activeTeamId={activeTeam?.id}
        canUseBackgroundMaya={canUseBackgroundMaya}
        createMore={createMore}
        currentTeam={currentTeam}
        currentTeamId={currentTeamId}
        deadlineSourceRef={deadlineSourceRef}
        descriptionEditor={editor}
        dispatch={dispatch}
        estimateScheme={estimateScheme}
        figmaArtifacts={figmaArtifacts}
        isCreating={isCreating}
        isExpanded={isExpanded}
        isMayaAssigned={isMayaAssigned}
        isMayaAssigneeLoading={isMayaAssigneeLoading}
        isOpen={isOpen}
        mayaAssigneeId={mayaAssignee?.id}
        mediaInputRef={mediaInputRef}
        member={member}
        members={members}
        objectiveTerm={getTermDisplay("objectiveTerm", { capitalize: true })}
        onActiveTeamChange={setActiveTeam}
        onCreate={handleCreateStory}
        onCreateMoreChange={setCreateMore}
        onFigmaArtifactsChange={setFigmaArtifacts}
        onMediaFiles={(files) => {
          if (editor) handleMediaFiles(editor, files);
        }}
        onOpenChange={(open) => {
          if (!open) setFigmaArtifacts([]);
          setIsOpen(open);
        }}
        onSimilarStorySelect={({ id, sequenceId, teamCode }) => {
          setIsOpen(false);
          router.push(
            withWorkspace(
              getStoryPath({
                id,
                sequenceId,
                teamCode,
              }),
            ),
          );
        }}
        onToggleExpanded={() => {
          setIsExpanded((previous) => !previous);
        }}
        selectedLabels={selectedLabels}
        showObjectives={
          Boolean(features.objectiveEnabled) && objectives.length > 0
        }
        showSprints={Boolean(sprintsEnabled) && sprints.length > 0}
        similarStories={similarStoryItems}
        sprintName={sprint?.name}
        sprintTerm={getTermDisplay("sprintTerm", { capitalize: true })}
        statuses={statuses}
        storyForm={storyForm}
        storyTerm={storyTerm}
        storyTermPlural={storyTermPlural}
        strategyLinkLabel={strategyLinkLabel}
        teamStatuses={teamStatuses}
        teams={teams}
        titleEditor={titleEditor}
      />
    </NewStoryDialogLimitGuard>
  );
};
