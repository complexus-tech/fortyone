import type { Objective } from "@/modules/objectives/public/types";
import type { StrategyMap } from "@/shared/strategy-map/types";
import type { ImportDraft } from "../schema";
import { resolveImportEntityNameMatch } from "../execution";
import type { ObjectiveTargetPlan } from "./import-wizard-model";
import { normalizeImportReviewName } from "./import-wizard-model";
import type {
  ObjectiveImportDestinationPreview,
  StrategicPillarImportDestinationPreview,
} from "./import-review-types";

export const getObjectiveDestinationMatches = ({
  draft,
  excludedObjectives,
  excludedStrategicPillars,
  objectiveTerm,
  objectiveTargetPlans,
  objectivesByTeamId,
  sourceObjectiveMappingsForReview,
  strategyMapForReview,
}: {
  draft: ImportDraft | null;
  excludedObjectives: ReadonlySet<string>;
  excludedStrategicPillars: ReadonlySet<string>;
  objectiveTerm: string;
  objectiveTargetPlans: ReadonlyMap<string, ObjectiveTargetPlan>;
  objectivesByTeamId: ReadonlyMap<string, Objective[]>;
  sourceObjectiveMappingsForReview: ReadonlyMap<
    string,
    { id: string; teamId: string }
  >;
  strategyMapForReview: StrategyMap;
}) => {
  const matches = new Map<string, ObjectiveImportDestinationPreview>();
  if (!draft) return matches;

  const sourceNameCounts = new Map<string, number>();
  for (const objective of draft.objectives) {
    if (excludedObjectives.has(objective.sourceId)) continue;
    const target = objectiveTargetPlans.get(objective.sourceId);
    if (!target) continue;
    const key = `${target.teamKey}\0${normalizeImportReviewName(objective.name)}`;
    sourceNameCounts.set(key, (sourceNameCounts.get(key) ?? 0) + 1);
  }

  const withPillarReview = (
    objective: ImportDraft["objectives"][number],
    preview: ObjectiveImportDestinationPreview,
    destinationObjectiveId?: string,
  ): ObjectiveImportDestinationPreview => {
    if (!objective.pillarSourceId) return preview;
    const sourcePillar = draft.strategicPillars.find(
      (pillar) => pillar.sourceId === objective.pillarSourceId,
    );
    if (
      !sourcePillar ||
      excludedStrategicPillars.has(objective.pillarSourceId)
    ) {
      return {
        ...preview,
        pillarLabel: `Referenced pillar ${sourcePillar?.name ?? objective.pillarSourceId} is not selected; this ${objectiveTerm} will import without that alignment.`,
      };
    }
    if (!destinationObjectiveId) return preview;

    const currentPillars = strategyMapForReview.pillars.filter((pillar) =>
      pillar.objectiveIds.includes(destinationObjectiveId),
    );
    if (currentPillars.length === 0) return preview;
    const destinationPillar = resolveImportEntityNameMatch(
      sourcePillar.name,
      strategyMapForReview.pillars,
    );
    if (
      currentPillars.length === 1 &&
      destinationPillar.kind === "unique" &&
      currentPillars[0].id === destinationPillar.entity.id
    ) {
      return preview;
    }
    return {
      ...preview,
      kind: "pillar_conflict",
      pillarLabel: `The reused ${objectiveTerm} is already aligned to ${currentPillars.map((pillar) => pillar.name).join(", ")}; the source requests ${sourcePillar.name}. Resolve that alignment before importing.`,
    };
  };

  for (const objective of draft.objectives) {
    const target = objectiveTargetPlans.get(objective.sourceId);
    if (!target) continue;
    if (target.teamConflict) {
      matches.set(objective.sourceId, {
        kind: "team_conflict",
        teamLabel: target.teamLabel,
      });
      continue;
    }
    const sourceNameKey = `${target.teamKey}\0${normalizeImportReviewName(objective.name)}`;
    if (
      !excludedObjectives.has(objective.sourceId) &&
      (sourceNameCounts.get(sourceNameKey) ?? 0) > 1
    ) {
      matches.set(objective.sourceId, {
        kind: "source_conflict",
        teamLabel: target.teamLabel,
      });
      continue;
    }
    const cachedObjective = sourceObjectiveMappingsForReview.get(
      objective.sourceId,
    );
    if (target.teamId && cachedObjective?.teamId === target.teamId) {
      const existingObjective = (
        objectivesByTeamId.get(target.teamId) ?? []
      ).find((candidate) => candidate.id === cachedObjective.id);
      matches.set(
        objective.sourceId,
        withPillarReview(
          objective,
          {
            kind: "unique",
            locked: true,
            objectiveName: existingObjective?.name ?? objective.name,
            teamLabel: target.teamLabel,
          },
          cachedObjective.id,
        ),
      );
      continue;
    }
    if (!target.teamId) {
      matches.set(
        objective.sourceId,
        withPillarReview(objective, {
          kind: "none",
          teamLabel: target.teamLabel,
        }),
      );
      continue;
    }
    const exactNameCandidates = (
      objectivesByTeamId.get(target.teamId) ?? []
    ).filter(
      (candidate) =>
        normalizeImportReviewName(candidate.name) ===
        normalizeImportReviewName(objective.name),
    );
    const candidates = exactNameCandidates.filter(
      (candidate) => candidate.isPrivate === objective.isPrivate,
    );
    const resolution = resolveImportEntityNameMatch(objective.name, candidates);
    if (resolution.kind === "unique") {
      matches.set(
        objective.sourceId,
        withPillarReview(
          objective,
          {
            kind: "unique",
            objectiveName: resolution.entity.name,
            teamLabel: target.teamLabel,
          },
          resolution.entity.id,
        ),
      );
    } else if (resolution.kind === "ambiguous") {
      matches.set(objective.sourceId, {
        kind: "ambiguous",
        matchCount: resolution.entities.length,
        teamLabel: target.teamLabel,
      });
    } else if (exactNameCandidates.length > 0) {
      matches.set(objective.sourceId, {
        kind: "privacy_conflict",
        teamLabel: target.teamLabel,
      });
    } else {
      matches.set(
        objective.sourceId,
        withPillarReview(objective, {
          kind: "none",
          teamLabel: target.teamLabel,
        }),
      );
    }
  }
  return matches;
};
export const getStrategicPillarDestinationMatches = (
  draft: ImportDraft | null,
  excludedStrategicPillars: ReadonlySet<string>,
  strategyMapForReview: StrategyMap,
) => {
  const matches = new Map<string, StrategicPillarImportDestinationPreview>();
  if (!draft) return matches;
  const sourceNameCounts = new Map<string, number>();
  for (const pillar of draft.strategicPillars) {
    if (excludedStrategicPillars.has(pillar.sourceId)) continue;
    const name = normalizeImportReviewName(pillar.name);
    sourceNameCounts.set(name, (sourceNameCounts.get(name) ?? 0) + 1);
  }
  for (const pillar of draft.strategicPillars) {
    const normalizedName = normalizeImportReviewName(pillar.name);
    if (
      !excludedStrategicPillars.has(pillar.sourceId) &&
      (sourceNameCounts.get(normalizedName) ?? 0) > 1
    ) {
      matches.set(pillar.sourceId, { kind: "source_conflict" });
      continue;
    }
    const resolution = resolveImportEntityNameMatch(
      pillar.name,
      strategyMapForReview.pillars,
    );
    if (resolution.kind === "unique") {
      matches.set(pillar.sourceId, {
        kind: "unique",
        pillarName: resolution.entity.name,
      });
    } else if (resolution.kind === "ambiguous") {
      matches.set(pillar.sourceId, {
        kind: "ambiguous",
        matchCount: resolution.entities.length,
      });
    } else {
      matches.set(pillar.sourceId, { kind: "none" });
    }
  }
  return matches;
};
