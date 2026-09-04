import { useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import type { Session } from "@/auth";
import type { Team } from "@/modules/teams/public/types";
import { storyKeys } from "@/modules/stories/public/keys";
import {
  labelKeys,
  memberKeys,
  sprintKeys,
  statusKeys,
  teamKeys,
} from "@/constants/keys";
import { objectiveKeys } from "@/shared/objectives/keys";
import { strategyKeys } from "@/shared/strategy-map/hooks";
import type { ImportDraft } from "../schema";
import type { ImportRunResult, ImportStructureMode } from "../import-run-model";
import { createImportTeam } from "../api";
import { runImport } from "../run-import";
import type { DestinationChoice } from "./import-wizard-model";
import { getNewTeamImportSignature } from "./import-wizard-model";

const EMPTY_IMPORT_SOURCE_IDS = new Set<string>();
type StartImportOptions = {
  draft: ImportDraft | null;
  destination: DestinationChoice;
  fileHash: string;
  selectedEntityCount: number;
  lockMemberMappings: () => Map<string, string | null>;
  onStart: () => void;
  hasSelectedTeamScopedImport: boolean;
  knownWorkspaceTeams: Team[];
  joinedTeamIds: ReadonlySet<string>;
  selectedObjectives: ImportDraft["objectives"];
  selectedStrategicPillars: ImportDraft["strategicPillars"];
  excludedRows: ReadonlySet<number>;
  structureMode: ImportStructureMode;
};
export const useImportExecution = (
  session: Session | null | undefined,
  workspaceSlug: string,
) => {
  const queryClient = useQueryClient();
  const createdSourceTeamIds = useRef(new Map<string, string>());
  const sourceObjectiveCache = useRef(
    new Map<string, { id: string; teamId: string }>(),
  );
  const createdFallbackTeam = useRef<{ id: string; signature: string } | null>(
    null,
  );
  const [createdFallbackTeamForReview, setCreatedFallbackTeamForReview] =
    useState<{ id: string; signature: string } | null>(null);
  const [sourceTeamMappingsForReview, setSourceTeamMappingsForReview] =
    useState<Map<string, string>>(new Map());
  const [
    sourceObjectiveMappingsForReview,
    setSourceObjectiveMappingsForReview,
  ] = useState<Map<string, { id: string; teamId: string }>>(new Map());
  const [importPending, setImportPending] = useState(false);
  const [importProgress, setImportProgress] = useState(0);
  const [hasAttemptedImport, setHasAttemptedImport] = useState(false);
  const [outcome, setOutcome] = useState<ImportRunResult | null>(null);
  const [runError, setRunError] = useState("");
  const reset = () => {
    createdSourceTeamIds.current.clear();
    sourceObjectiveCache.current.clear();
    createdFallbackTeam.current = null;
    setCreatedFallbackTeamForReview(null);
    setSourceTeamMappingsForReview(new Map());
    setSourceObjectiveMappingsForReview(new Map());
    setImportPending(false);
    setImportProgress(0);
    setHasAttemptedImport(false);
    setOutcome(null);
    setRunError("");
  };
  const getDestinationTeam = async (
    destination: DestinationChoice,
    fileHash: string,
  ) => {
    if (destination.kind === "existing") {
      return { created: false, id: destination.teamId };
    }
    if (!session)
      throw new Error("Your session expired. Refresh and try again.");

    const signature = getNewTeamImportSignature(fileHash, destination);
    if (createdFallbackTeam.current?.signature === signature) {
      return { created: false, id: createdFallbackTeam.current.id };
    }

    const response = await createImportTeam(
      {
        code: destination.code.trim().toUpperCase(),
        color: destination.color,
        isPrivate: destination.isPrivate,
        name: destination.name.trim(),
      },
      { session, workspaceSlug },
    );
    if (response.error?.message || !response.data?.id) {
      throw new Error(response.error?.message || "Unable to create the team");
    }
    createdFallbackTeam.current = { id: response.data.id, signature };
    setCreatedFallbackTeamForReview({ id: response.data.id, signature });
    return { created: true, id: response.data.id };
  };

  const startImport = ({
    draft,
    destination,
    fileHash,
    selectedEntityCount,
    lockMemberMappings,
    onStart,
    hasSelectedTeamScopedImport,
    knownWorkspaceTeams,
    joinedTeamIds,
    selectedObjectives,
    selectedStrategicPillars,
    excludedRows,
    structureMode,
  }: StartImportOptions) => {
    if (!draft || !session || selectedEntityCount === 0) return;
    const memberIdsByIdentityKey = lockMemberMappings();
    setHasAttemptedImport(true);
    onStart();
    setImportPending(true);
    setImportProgress(0);
    setRunError("");
    setOutcome(null);

    const finishImport = () => {
      setSourceTeamMappingsForReview(new Map(createdSourceTeamIds.current));
      setSourceObjectiveMappingsForReview(
        new Map(sourceObjectiveCache.current),
      );
      void Promise.allSettled([
        queryClient.invalidateQueries({
          queryKey: storyKeys.all(workspaceSlug),
        }),
        queryClient.invalidateQueries({
          queryKey: teamKeys.all(workspaceSlug),
        }),
        queryClient.invalidateQueries({
          queryKey: statusKeys.all(workspaceSlug),
        }),
        queryClient.invalidateQueries({
          queryKey: memberKeys.all(workspaceSlug),
        }),
        queryClient.invalidateQueries({
          queryKey: objectiveKeys.all(workspaceSlug),
        }),
        queryClient.invalidateQueries({
          queryKey: sprintKeys.all(workspaceSlug),
        }),
        queryClient.invalidateQueries({
          queryKey: labelKeys.all(workspaceSlug),
        }),
        queryClient.invalidateQueries({
          queryKey: strategyKeys.map(workspaceSlug),
        }),
      ]).then(() => {
        setImportPending(false);
      });
    };
    const fallbackTeamPromise = hasSelectedTeamScopedImport
      ? getDestinationTeam(destination, fileHash)
      : Promise.resolve(null);
    void fallbackTeamPromise
      .then((fallbackTeam) => {
        const fallbackTeamId = fallbackTeam?.id ?? null;
        const existingFallbackTeam = fallbackTeamId
          ? knownWorkspaceTeams.find((team) => team.id === fallbackTeamId)
          : undefined;
        return runImport({
          actorUserId: session.user.id,
          confirmedMemberIdsByIdentityKey: memberIdsByIdentityKey,
          ctx: { session, workspaceSlug },
          draft,
          existingTeams: knownWorkspaceTeams,
          fallbackTeamCode:
            destination.kind === "new"
              ? destination.code
              : existingFallbackTeam?.code ?? "",
          fallbackTeamCreated: fallbackTeam?.created ?? false,
          fallbackTeamId,
          fallbackTeamIsPrivate:
            destination.kind === "new"
              ? destination.isPrivate
              : existingFallbackTeam?.isPrivate ?? false,
          fallbackTeamIsNew: Boolean(
            fallbackTeam && destination.kind === "new",
          ),
          fallbackTeamName:
            destination.kind === "new"
              ? destination.name
              : existingFallbackTeam?.name ?? "",
          forceCreateObjectiveSourceIds: EMPTY_IMPORT_SOURCE_IDS,
          joinedTeamIds,
          onProgress: setImportProgress,
          selectedObjectiveSourceIds: new Set(
            selectedObjectives.map((objective) => objective.sourceId),
          ),
          selectedStrategicPillarSourceIds: new Set(
            selectedStrategicPillars.map((pillar) => pillar.sourceId),
          ),
          selectedTaskIndexes: new Set(
            draft.tasks.flatMap((_, index) =>
              excludedRows.has(index) ? [] : [index],
            ),
          ),
          sourceTeamCache: createdSourceTeamIds.current,
          sourceObjectiveCache: sourceObjectiveCache.current,
          structureMode,
        });
      })
      .then(
        (result) => {
          setOutcome(result);
          finishImport();
        },
        (error: unknown) => {
          setRunError(
            error instanceof Error
              ? error.message
              : "The import could not finish",
          );
          finishImport();
        },
      );
  };
  return {
    importPending,
    importProgress,
    hasAttemptedImport,
    outcome,
    runError,
    setOutcome,
    createdFallbackTeamForReview,
    sourceTeamMappingsForReview,
    sourceObjectiveMappingsForReview,
    startImport,
    reset,
  };
};
