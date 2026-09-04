import { useMemo } from "react";
import type { Team } from "@/modules/teams/public/types";
import type { ImportDraft } from "../schema";
import type { ImportStructureMode } from "../import-run-model";
import {
  getImportSourceTeamDestination,
  resolveImportSourceTeam,
} from "../import-team-model";
import type {
  DestinationChoice,
  ObjectiveTargetPlan,
} from "./import-wizard-model";
import {
  getNewTeamImportSignature,
  normalizeImportReviewName,
} from "./import-wizard-model";

export const useImportTargetPlans = ({
  draft,
  destination,
  fileHash,
  structureMode,
  knownWorkspaceTeams,
  createdFallbackTeamForReview,
  sourceTeamMappingsForReview,
}: {
  draft: ImportDraft | null;
  destination: DestinationChoice;
  fileHash: string;
  structureMode: ImportStructureMode;
  knownWorkspaceTeams: Team[];
  createdFallbackTeamForReview: { id: string; signature: string } | null;
  sourceTeamMappingsForReview: ReadonlyMap<string, string>;
}) => {
  const draftTeams = draft?.teams;
  const fallbackTargetPlan = useMemo<ObjectiveTargetPlan>(() => {
    const newTeamSignature =
      destination.kind === "new"
        ? getNewTeamImportSignature(fileHash, destination)
        : null;
    const createdFallbackTeamId =
      createdFallbackTeamForReview?.signature === newTeamSignature
        ? createdFallbackTeamForReview.id
        : null;
    const fallbackTeam =
      destination.kind === "existing"
        ? knownWorkspaceTeams.find((team) => team.id === destination.teamId)
        : undefined;
    const teamId =
      destination.kind === "existing"
        ? destination.teamId
        : createdFallbackTeamId;
    return {
      teamConflict: false,
      teamId,
      teamKey: teamId ? `existing:${teamId}` : "new:fallback",
      teamLabel:
        destination.kind === "existing"
          ? fallbackTeam?.name ?? "destination team"
          : destination.name.trim() || "new destination team",
    };
  }, [
    createdFallbackTeamForReview,
    destination,
    fileHash,
    knownWorkspaceTeams,
  ]);
  const sourceTeamTargetPlans = useMemo(() => {
    const plans = new Map<string, ObjectiveTargetPlan>();
    if (!draftTeams || structureMode !== "preserve") return plans;

    for (const sourceTeam of draftTeams) {
      const cachedTeamId = sourceTeamMappingsForReview.get(sourceTeam.sourceId);
      if (cachedTeamId) {
        plans.set(sourceTeam.sourceId, {
          teamConflict: false,
          teamId: cachedTeamId,
          teamKey: `existing:${cachedTeamId}`,
          teamLabel: sourceTeam.name,
        });
        continue;
      }
      const resolution = resolveImportSourceTeam(
        sourceTeam,
        knownWorkspaceTeams,
      );
      if (resolution.kind === "unique") {
        plans.set(sourceTeam.sourceId, {
          teamConflict: false,
          teamId: resolution.team.id,
          teamKey: `existing:${resolution.team.id}`,
          teamLabel: resolution.team.name,
        });
        continue;
      }
      const matchesNewFallback =
        resolution.kind === "none" &&
        destination.kind === "new" &&
        sourceTeam.isPrivate === destination.isPrivate &&
        normalizeImportReviewName(
          getImportSourceTeamDestination(sourceTeam).name,
        ) === normalizeImportReviewName(destination.name);
      let teamKey = `new:source:${sourceTeam.sourceId}`;
      if (resolution.kind === "ambiguous") {
        teamKey = `conflict:source:${sourceTeam.sourceId}`;
      } else if (matchesNewFallback) {
        teamKey = "new:fallback";
      }
      plans.set(sourceTeam.sourceId, {
        teamConflict: resolution.kind === "ambiguous",
        teamId: null,
        teamKey,
        teamLabel: matchesNewFallback
          ? destination.name.trim() || sourceTeam.name
          : sourceTeam.name,
      });
    }

    const destinationGroups = new Map<string, string[]>();
    for (const sourceTeam of draftTeams) {
      const plan = plans.get(sourceTeam.sourceId);
      if (!plan || plan.teamConflict) continue;
      let collisionKey = `new-name:${sourceTeam.isPrivate}:${normalizeImportReviewName(
        getImportSourceTeamDestination(sourceTeam).name,
      )}`;
      if (plan.teamId) collisionKey = `existing:${plan.teamId}`;
      else if (plan.teamKey === "new:fallback") collisionKey = plan.teamKey;
      const sourceIds = destinationGroups.get(collisionKey) ?? [];
      sourceIds.push(sourceTeam.sourceId);
      destinationGroups.set(collisionKey, sourceIds);
    }
    for (const [collisionKey, sourceIds] of destinationGroups) {
      if (sourceIds.length < 2) continue;
      for (const sourceId of sourceIds) {
        const plan = plans.get(sourceId);
        if (!plan) continue;
        plans.set(sourceId, {
          ...plan,
          teamConflict: true,
          teamKey: `conflict:collision:${collisionKey}`,
        });
      }
    }
    return plans;
  }, [
    destination,
    draftTeams,
    knownWorkspaceTeams,
    sourceTeamMappingsForReview,
    structureMode,
  ]);
  const objectiveTargetPlans = useMemo(() => {
    const plans = new Map<string, ObjectiveTargetPlan>();
    if (!draft) return plans;

    for (const objective of draft.objectives) {
      if (structureMode === "single" || !objective.teamSourceId) {
        plans.set(objective.sourceId, fallbackTargetPlan);
        continue;
      }
      const sourceTeamPlan = sourceTeamTargetPlans.get(objective.teamSourceId);
      plans.set(objective.sourceId, sourceTeamPlan ?? fallbackTargetPlan);
    }
    return plans;
  }, [draft, fallbackTargetPlan, sourceTeamTargetPlans, structureMode]);
  const hasSourceTeamConflict = [...sourceTeamTargetPlans.values()].some(
    (plan) => plan.teamConflict,
  );
  const conflictedSourceTeamIds = useMemo(
    () =>
      new Set(
        [...sourceTeamTargetPlans].flatMap(([sourceId, plan]) =>
          plan.teamConflict ? [sourceId] : [],
        ),
      ),
    [sourceTeamTargetPlans],
  );
  const objectiveTargetTeamIds = useMemo(() => {
    const teamIds = [...objectiveTargetPlans.values()].flatMap((plan) =>
      plan.teamId ? [plan.teamId] : [],
    );
    return [...new Set(teamIds)].sort();
  }, [objectiveTargetPlans]);

  return {
    fallbackTargetPlan,
    sourceTeamTargetPlans,
    objectiveTargetPlans,
    hasSourceTeamConflict,
    conflictedSourceTeamIds,
    objectiveTargetTeamIds,
  };
};
