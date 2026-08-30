import { useEffect, type Dispatch } from "react";
import { MINIMUM_SIMILARITY_TITLE_CHARACTERS } from "@/constants/similarity";
import type {
  NewStoryDialogForm,
  StoryFormAction,
} from "./new-story-dialog-form";

type TeamWithId = {
  id: string;
};

type StatusWithId = {
  id: string;
};

type TitleEditor = {
  commands: {
    focus: () => unknown;
  };
};

export const useNewStoryDialogLifecycle = <
  TTeam extends TeamWithId,
  TStatus extends StatusWithId,
>({
  activeTeamId,
  cancelStagedUploads,
  cancelStorySearch,
  currentTeamId,
  dispatch,
  firstTeam,
  isOpen,
  setActiveTeam,
  searchSimilarStories,
  statusId,
  storyForm,
  storyTitle,
  teamStatuses,
  teams,
  titleEditor,
}: {
  activeTeamId?: string;
  cancelStagedUploads: () => void;
  cancelStorySearch: () => void;
  currentTeamId?: string;
  dispatch: Dispatch<StoryFormAction>;
  firstTeam: TTeam | null;
  isOpen: boolean;
  setActiveTeam: (team: TTeam | null) => void;
  searchSimilarStories: (title: string) => void;
  statusId?: string;
  storyForm: NewStoryDialogForm;
  storyTitle: string;
  teamStatuses: TStatus[];
  teams: TTeam[];
  titleEditor: TitleEditor | null;
}) => {
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
  }, [currentTeamId, dispatch, statusId, storyForm.statusId, teamStatuses]);

  useEffect(() => {
    if (!teams.some((team) => team.id === activeTeamId)) {
      setActiveTeam(firstTeam);
    }
  }, [activeTeamId, firstTeam, setActiveTeam, teams]);

  useEffect(() => {
    if (isOpen && titleEditor) titleEditor.commands.focus();
  }, [isOpen, titleEditor]);

  useEffect(() => {
    if (!isOpen) cancelStagedUploads();
  }, [cancelStagedUploads, isOpen]);

  useEffect(() => {
    if (
      !isOpen ||
      storyTitle.trim().length < MINIMUM_SIMILARITY_TITLE_CHARACTERS
    ) {
      cancelStorySearch();
      return;
    }
    searchSimilarStories(storyTitle.trim());
  }, [cancelStorySearch, isOpen, searchSimilarStories, storyTitle]);
};

export const useMayaAutoScheduling = ({
  dispatch,
  isMayaAssigned,
  isAutoSchedulingEnabled,
}: {
  dispatch: Dispatch<StoryFormAction>;
  isMayaAssigned: boolean;
  isAutoSchedulingEnabled?: boolean;
}) => {
  useEffect(() => {
    if (isMayaAssigned && !isAutoSchedulingEnabled) {
      dispatch({
        type: "PATCH_FORM",
        payload: { autoSchedulingEnabled: true },
      });
    }
  }, [dispatch, isAutoSchedulingEnabled, isMayaAssigned]);
};
