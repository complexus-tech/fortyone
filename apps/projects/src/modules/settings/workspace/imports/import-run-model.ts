import type { WorkspaceCtx } from "@/lib/http";
import type { Team } from "@/modules/teams/public/types";
import type { ImportDraft } from "./schema";
import type { ImportStoryResult } from "./api";

export type ImportStructureMode = "preserve" | "single";

export type ImportRunResult = {
  created: number;
  replayed: number;
  failed: number;
  items: ImportStoryResult[];
  teamId: string | null;
  createdTeams: number;
  createdStrategicPillars: number;
  createdObjectives: number;
  createdKeyResults: number;
  createdSprints: number;
  createdLabels: number;
  createdLinks: number;
  addedMemberships: number;
  appliedCollaborators: number;
  createdAssociations: number;
  alignedObjectives: number;
  destinationConflicts: number;
  unresolvedAssociations: number;
  unresolvedLinks: number;
  unresolvedPeople: number;
};

export type RunImportInput = {
  actorUserId: string;
  draft: ImportDraft;
  selectedTaskIndexes: ReadonlySet<number>;
  selectedObjectiveSourceIds: ReadonlySet<string>;
  selectedStrategicPillarSourceIds: ReadonlySet<string>;
  structureMode: ImportStructureMode;
  fallbackTeamId: string | null;
  fallbackTeamIsPrivate: boolean;
  fallbackTeamName: string;
  fallbackTeamCode: string;
  fallbackTeamIsNew: boolean;
  fallbackTeamCreated: boolean;
  existingTeams: readonly Team[];
  forceCreateObjectiveSourceIds: ReadonlySet<string>;
  joinedTeamIds: ReadonlySet<string>;
  confirmedMemberIdsByIdentityKey: ReadonlyMap<string, string | null>;
  sourceTeamCache: Map<string, string>;
  sourceObjectiveCache: Map<string, { id: string; teamId: string }>;
  ctx: WorkspaceCtx;
  onProgress: (progress: number) => void;
};
