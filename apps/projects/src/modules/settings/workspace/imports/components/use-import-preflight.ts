import { useEffect, useState } from "react";
import type { Session } from "@/auth";
import type { Objective } from "@/modules/objectives/public/types";
import type { Member } from "@/types/member";
import type { StrategyMap } from "@/shared/strategy-map/types";
import {
  getImportTeamObjectives,
  getImportWorkspaceMembers,
  getImportStrategyMap,
} from "../api";

const EMPTY_IMPORT_STRATEGY_MAP: StrategyMap = {
  description: null,
  pillars: [],
  ultimateGoal: "",
};
export const useImportPreflight = ({
  session,
  workspaceSlug,
  fileHash,
  objectiveTargetTeamIds,
  hasImportIdentities,
  strategyPreflightRequired,
  objectiveTermPlural,
}: {
  session: Session | null | undefined;
  workspaceSlug: string;
  fileHash: string;
  objectiveTargetTeamIds: string[];
  hasImportIdentities: boolean;
  strategyPreflightRequired: boolean;
  objectiveTermPlural: string;
}) => {
  const [objectivesByTeamId, setObjectivesByTeamId] = useState<
    Map<string, Objective[]>
  >(new Map());
  const [objectivePreflightPending, setObjectivePreflightPending] =
    useState(false);
  const [objectivePreflightError, setObjectivePreflightError] = useState("");
  const [objectivePreflightRevision, setObjectivePreflightRevision] =
    useState(0);
  const [workspaceMembersForReview, setWorkspaceMembersForReview] = useState<
    Member[]
  >([]);
  const [peoplePreflightPending, setPeoplePreflightPending] = useState(false);
  const [peoplePreflightError, setPeoplePreflightError] = useState("");
  const [peoplePreflightRevision, setPeoplePreflightRevision] = useState(0);
  const [strategyMapForReview, setStrategyMapForReview] = useState<StrategyMap>(
    EMPTY_IMPORT_STRATEGY_MAP,
  );
  const [strategyPreflightPending, setStrategyPreflightPending] =
    useState(false);
  const [strategyPreflightError, setStrategyPreflightError] = useState("");
  const [strategyPreflightRevision, setStrategyPreflightRevision] = useState(0);
  useEffect(() => {
    let cancelled = false;
    if (!session || objectiveTargetTeamIds.length === 0) {
      setObjectivesByTeamId(new Map());
      setObjectivePreflightPending(false);
      setObjectivePreflightError("");
      return;
    }

    const loadObjectives = () => {
      setObjectivePreflightPending(true);
      setObjectivePreflightError("");
      void Promise.all(
        objectiveTargetTeamIds.map(async (teamId) => {
          const objectives = await getImportTeamObjectives(teamId, {
            session,
            workspaceSlug,
          });
          return [teamId, objectives] as const;
        }),
      ).then(
        (entries) => {
          if (cancelled) return;
          setObjectivesByTeamId(new Map(entries));
          setObjectivePreflightPending(false);
        },
        () => {
          if (cancelled) return;
          setObjectivePreflightError(
            `Existing ${objectiveTermPlural} could not be checked safely.`,
          );
          setObjectivePreflightPending(false);
        },
      );
    };

    loadObjectives();
    return () => {
      cancelled = true;
    };
  }, [
    objectivePreflightRevision,
    objectiveTermPlural,
    objectiveTargetTeamIds,
    session,
    workspaceSlug,
  ]);

  useEffect(() => {
    let cancelled = false;
    if (!session || !hasImportIdentities) {
      setWorkspaceMembersForReview([]);
      setPeoplePreflightPending(false);
      setPeoplePreflightError("");
      return;
    }

    const loadWorkspaceMembers = () => {
      setPeoplePreflightPending(true);
      setPeoplePreflightError("");
      void getImportWorkspaceMembers({
        session,
        workspaceSlug,
      }).then(
        (members) => {
          if (cancelled) return;
          setWorkspaceMembersForReview(members);
          setPeoplePreflightPending(false);
        },
        () => {
          if (cancelled) return;
          setPeoplePreflightError(
            "Workspace members could not be checked safely.",
          );
          setPeoplePreflightPending(false);
        },
      );
    };

    loadWorkspaceMembers();
    return () => {
      cancelled = true;
    };
  }, [
    fileHash,
    hasImportIdentities,
    peoplePreflightRevision,
    session,
    workspaceSlug,
  ]);

  useEffect(() => {
    let cancelled = false;
    if (!session || !strategyPreflightRequired) {
      setStrategyMapForReview(EMPTY_IMPORT_STRATEGY_MAP);
      setStrategyPreflightPending(false);
      setStrategyPreflightError("");
      return;
    }

    const loadStrategyMap = () => {
      setStrategyPreflightPending(true);
      setStrategyPreflightError("");
      void getImportStrategyMap({
        session,
        workspaceSlug,
      }).then(
        (strategyMap) => {
          if (cancelled) return;
          setStrategyMapForReview(strategyMap);
          setStrategyPreflightPending(false);
        },
        () => {
          if (cancelled) return;
          setStrategyPreflightError(
            "Existing strategic pillars could not be checked safely.",
          );
          setStrategyPreflightPending(false);
        },
      );
    };

    loadStrategyMap();
    return () => {
      cancelled = true;
    };
  }, [
    fileHash,
    session,
    strategyPreflightRequired,
    strategyPreflightRevision,
    workspaceSlug,
  ]);

  const retryObjectives = () => {
    setObjectivePreflightRevision((current) => current + 1);
  };
  const retryPeople = () => {
    setPeoplePreflightRevision((current) => current + 1);
  };
  const retryStrategy = () => {
    setStrategyPreflightRevision((current) => current + 1);
  };
  const retryPreflight = () => {
    retryObjectives();
    retryPeople();
    retryStrategy();
  };
  const reset = () => {
    setObjectivesByTeamId(new Map());
    setObjectivePreflightPending(false);
    setObjectivePreflightError("");
    setObjectivePreflightRevision(0);
    setWorkspaceMembersForReview([]);
    setPeoplePreflightPending(false);
    setPeoplePreflightError("");
    setPeoplePreflightRevision(0);
    setStrategyMapForReview(EMPTY_IMPORT_STRATEGY_MAP);
    setStrategyPreflightPending(false);
    setStrategyPreflightError("");
    setStrategyPreflightRevision(0);
  };
  return {
    objectivesByTeamId,
    objectivePreflightPending,
    objectivePreflightError,
    workspaceMembersForReview,
    peoplePreflightPending,
    peoplePreflightError,
    strategyMapForReview,
    strategyPreflightPending,
    strategyPreflightError,
    retryObjectives,
    retryPeople,
    retryStrategy,
    retryPreflight,
    reset,
  };
};
