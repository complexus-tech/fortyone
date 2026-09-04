import type { WorkspaceCtx } from "@/lib/http";
import { get, post, put } from "@/lib/http";
import type { ApiResponse } from "@/types/api-response";
import type { Label } from "@/types/label";
import type { Link } from "@/types/link";
import type { Member } from "@/types/member";
import type { State } from "@/types/states";
import type { Team, CreateTeamInput } from "@/modules/teams/public/types";
import type {
  NewStrategicPillar,
  StrategicPillar,
} from "@/modules/strategy/public/types";
import type {
  KeyResult,
  NewKeyResult,
  NewObjective,
  Objective,
  ObjectiveStatus,
} from "@/modules/objectives/public/types";
import type { NewSprint, Sprint } from "@/modules/sprints/public/types";
import type {
  StoryAssociation,
  StoryAssociationType,
} from "@/shared/story/types";
import {
  normalizeStrategyMap,
  type StrategyMapResponse,
} from "@/shared/strategy-map/normalize-strategy-map";
import type {
  ImportAnalysisPollResponse,
  ImportAnalysisStartResponse,
  ImportEstimateValue,
  ImportPriority,
} from "./schema";

export type ImportStoryPayload = {
  title: string;
  description: string;
  descriptionHTML?: string;
  teamId: string;
  statusId?: string;
  assigneeId?: string;
  objectiveId?: string;
  keyResultId?: string;
  sprintId?: string;
  parentId?: string;
  labelIds?: string[];
  priority: ImportPriority;
  estimateValue?: ImportEstimateValue;
  estimatedDurationMinutes?: number;
  minimumFocusBlockMinutes?: number;
  startDate?: string;
  endDate?: string;
};

export type ImportObjectiveCreateResult = {
  objective: Objective;
  keyResults?: KeyResult[];
};

export type ImportLabelCreateInput = {
  name: string;
  color: string;
  teamId?: string;
};

export type ImportStoryLinkCreateInput = {
  storyId: string;
  url: string;
  title?: string;
};

type ImportStoryRelationshipState = {
  associations?:
    | Pick<StoryAssociation, "fromStoryId" | "id" | "toStoryId" | "type">[]
    | null;
  collaboratorIds?: string[] | null;
};

export type ImportStoriesRequest = {
  provider: "jira_csv" | "file";
  sourceDigest: string;
  sourceNamespace?: string;
  items: {
    sourceKey: string;
    story: ImportStoryPayload;
  }[];
};

export type ImportStoryResult = {
  sourceKey: string;
  storyId: string | null;
  created: boolean;
  error: { code: string; message: string } | null;
};

export type ImportStoriesResult = {
  counts: {
    total: number;
    created: number;
    replayed: number;
    failed: number;
  };
  items: ImportStoryResult[];
};

const IMPORT_BATCH_MAX_ITEMS = 50;
const IMPORT_BATCH_TARGET_BYTES = 850 * 1024;

export const buildImportStoryRequests = ({
  items,
  provider,
  sourceDigest,
  sourceNamespace,
}: ImportStoriesRequest): ImportStoriesRequest[] => {
  const requests: ImportStoriesRequest[] = [];
  let batch: ImportStoriesRequest["items"] = [];

  for (const item of items) {
    const candidate = [...batch, item];
    const candidateRequest = {
      items: candidate,
      provider,
      sourceDigest,
      ...(sourceNamespace ? { sourceNamespace } : {}),
    };
    const candidateBytes = new TextEncoder().encode(
      JSON.stringify(candidateRequest),
    ).byteLength;

    if (
      batch.length > 0 &&
      (candidate.length > IMPORT_BATCH_MAX_ITEMS ||
        candidateBytes > IMPORT_BATCH_TARGET_BYTES)
    ) {
      requests.push({
        items: batch,
        provider,
        sourceDigest,
        ...(sourceNamespace ? { sourceNamespace } : {}),
      });
      batch = [item];
    } else {
      batch = candidate;
    }
  }

  if (batch.length > 0) {
    requests.push({
      items: batch,
      provider,
      sourceDigest,
      ...(sourceNamespace ? { sourceNamespace } : {}),
    });
  }

  return requests;
};

const readError = async (response: Response, fallback: string) => {
  const message = await response.text();
  return new Error(message.trim() || fallback);
};

const requireImportData = <T>(
  response: ApiResponse<T>,
  fallback: string,
): T => {
  if (response.error?.message) {
    throw new Error(response.error.message);
  }
  if (response.data === undefined || response.data === null) {
    throw new Error(fallback);
  }
  return response.data;
};

export const startImportAnalysis = async (
  file: File,
  workspaceSlug: string,
) => {
  const formData = new FormData();
  formData.set("file", file);
  const search = new URLSearchParams({ workspaceSlug });
  const response = await fetch(`/api/imports/analyze?${search.toString()}`, {
    body: formData,
    method: "POST",
  });
  if (!response.ok)
    throw await readError(response, "Unable to analyze the file");
  return response.json() as Promise<ImportAnalysisStartResponse>;
};

export const pollImportAnalysis = async ({
  fileHash,
  responseId,
  workspaceSlug,
}: {
  fileHash: string;
  responseId: string;
  workspaceSlug: string;
}) => {
  const search = new URLSearchParams({ fileHash, responseId, workspaceSlug });
  const response = await fetch(`/api/imports/analyze?${search.toString()}`);
  if (!response.ok)
    throw await readError(response, "Unable to load the analysis");
  return response.json() as Promise<ImportAnalysisPollResponse>;
};

export const createImportTeam = async (
  input: CreateTeamInput,
  ctx: WorkspaceCtx,
) => post<CreateTeamInput, ApiResponse<Team>>("teams", input, ctx);

export const getImportTeamStatuses = async (
  teamId: string,
  ctx: WorkspaceCtx,
) => {
  const response = await get<ApiResponse<State[]>>(
    `states?teamId=${teamId}`,
    ctx,
  );
  return requireImportData(response, "Unable to load the destination workflow");
};

export const getImportWorkspaceMembers = async (ctx: WorkspaceCtx) => {
  const response = await get<ApiResponse<Member[]>>("members", ctx);
  return requireImportData(response, "Unable to load workspace members");
};

export const getImportTeamMembers = async (
  teamId: string,
  ctx: WorkspaceCtx,
) => {
  const response = await get<ApiResponse<Member[]>>(
    `members?teamId=${teamId}`,
    ctx,
  );
  return requireImportData(
    response,
    "Unable to load the destination team members",
  );
};

export const addExistingImportMemberToTeam = async (
  teamId: string,
  memberId: string,
  ctx: WorkspaceCtx,
) => {
  const response = await post<
    { userId: string },
    ApiResponse<{ teamId: string }>
  >(`teams/${teamId}/members`, { userId: memberId }, ctx);
  return requireImportData(response, "Unable to add the workspace member");
};

export const getImportObjectiveStatuses = async (ctx: WorkspaceCtx) => {
  const response = await get<ApiResponse<ObjectiveStatus[]>>(
    "objective-statuses",
    ctx,
  );
  return requireImportData(response, "Unable to load objective statuses");
};

export const getImportTeamObjectives = async (
  teamId: string,
  ctx: WorkspaceCtx,
) => {
  const search = new URLSearchParams({ teamId });
  const response = await get<ApiResponse<Objective[]>>(
    `objectives?${search.toString()}`,
    ctx,
  );
  return requireImportData(response, "Unable to load team objectives");
};

export const getImportObjectiveKeyResults = async (
  objectiveId: string,
  ctx: WorkspaceCtx,
) => {
  const response = await get<ApiResponse<KeyResult[]>>(
    `objectives/${objectiveId}/key-results`,
    ctx,
  );
  return requireImportData(response, "Unable to load objective key results");
};

export const createImportObjective = async (
  input: NewObjective,
  ctx: WorkspaceCtx,
) => {
  const response = await post<
    NewObjective,
    ApiResponse<ImportObjectiveCreateResult>
  >("objectives", input, ctx);
  return requireImportData(response, "Unable to create the objective");
};

export const createImportKeyResults = async (
  objectiveId: string,
  keyResults: NewKeyResult[],
  ctx: WorkspaceCtx,
) => {
  const response = await post<
    { keyResults: NewKeyResult[] },
    ApiResponse<KeyResult[]>
  >(`objectives/${objectiveId}/key-results`, { keyResults }, ctx);
  return requireImportData(response, "Unable to create objective key results");
};

export const getImportStrategyMap = async (ctx: WorkspaceCtx) => {
  const response = await get<ApiResponse<StrategyMapResponse>>(
    "strategy-map",
    ctx,
  );
  return normalizeStrategyMap(
    requireImportData(response, "Unable to load the strategy map"),
  );
};

export const createImportStrategicPillar = async (
  input: NewStrategicPillar,
  ctx: WorkspaceCtx,
) => {
  const response = await post<NewStrategicPillar, ApiResponse<StrategicPillar>>(
    "strategy-map/pillars",
    input,
    ctx,
  );
  return requireImportData(response, "Unable to create the strategic pillar");
};

export const alignImportObjectiveToPillar = async (
  objectiveId: string,
  pillarId: string,
  ctx: WorkspaceCtx,
) => {
  const response = await put<{ pillarId: string }, ApiResponse<null>>(
    `strategy-map/objectives/${objectiveId}`,
    { pillarId },
    ctx,
  );
  if (response.error?.message) throw new Error(response.error.message);
};

export const getImportTeamLabels = async (
  teamId: string,
  ctx: WorkspaceCtx,
) => {
  const search = new URLSearchParams({ teamId });
  const response = await get<ApiResponse<Label[]>>(
    `labels?${search.toString()}`,
    ctx,
  );
  return requireImportData(response, "Unable to load team labels");
};

export const getImportWorkspaceLabels = async (ctx: WorkspaceCtx) => {
  const response = await get<ApiResponse<Label[]>>("labels", ctx);
  return requireImportData(response, "Unable to load workspace labels");
};

export const createImportLabel = async (
  input: ImportLabelCreateInput,
  ctx: WorkspaceCtx,
) => {
  const response = await post<ImportLabelCreateInput, ApiResponse<Label>>(
    "labels",
    input,
    ctx,
  );
  return requireImportData(response, "Unable to create the label");
};

export const getImportTeamSprints = async (
  teamId: string,
  ctx: WorkspaceCtx,
) => {
  const search = new URLSearchParams({ teamId });
  const response = await get<ApiResponse<Sprint[]>>(
    `sprints?${search.toString()}`,
    ctx,
  );
  return requireImportData(response, "Unable to load team sprints");
};

export const createImportSprint = async (
  input: NewSprint,
  ctx: WorkspaceCtx,
) => {
  const response = await post<NewSprint, ApiResponse<Sprint>>(
    "sprints",
    input,
    ctx,
  );
  return requireImportData(response, "Unable to create the sprint");
};

export const importStoriesBatch = async (
  input: ImportStoriesRequest,
  ctx: WorkspaceCtx,
) =>
  post<ImportStoriesRequest, ApiResponse<ImportStoriesResult>>(
    "stories/import",
    input,
    ctx,
  );

export const getImportStoryCollaboratorIds = async (
  storyId: string,
  ctx: WorkspaceCtx,
) => {
  const response = await get<ApiResponse<ImportStoryRelationshipState>>(
    `stories/${storyId}`,
    ctx,
  );
  return (
    requireImportData(response, "Unable to load work item collaborators")
      .collaboratorIds ?? []
  );
};

export const getImportStoryAssociations = async (
  storyId: string,
  ctx: WorkspaceCtx,
) => {
  const response = await get<ApiResponse<ImportStoryRelationshipState>>(
    `stories/${storyId}`,
    ctx,
  );
  return (
    requireImportData(response, "Unable to load work item relationships")
      .associations ?? []
  );
};

export const getImportStoryLinks = async (
  storyId: string,
  ctx: WorkspaceCtx,
) => {
  const response = await get<ApiResponse<Link[]>>(
    `stories/${storyId}/links`,
    ctx,
  );
  return requireImportData(response, "Unable to load work item links");
};

export const createImportStoryLink = async (
  input: ImportStoryLinkCreateInput,
  ctx: WorkspaceCtx,
) => {
  const response = await post<ImportStoryLinkCreateInput, ApiResponse<Link>>(
    "links",
    input,
    ctx,
  );
  return requireImportData(response, "Unable to create work item link");
};

export const createImportStoryAssociation = async (
  fromStoryId: string,
  input: { toStoryId: string; type: StoryAssociationType },
  ctx: WorkspaceCtx,
) => {
  const response = await post<
    { toStoryId: string; type: StoryAssociationType },
    ApiResponse<StoryAssociation>
  >(`stories/${fromStoryId}/associations`, input, ctx);
  return requireImportData(response, "Unable to create work item relationship");
};

export const updateImportStoryCollaborators = async (
  storyId: string,
  collaboratorIds: string[],
  ctx: WorkspaceCtx,
) => {
  const response = await put<{ collaboratorIds: string[] }, ApiResponse<null>>(
    `stories/${storyId}/collaborators`,
    { collaboratorIds },
    ctx,
  );
  if (response.error?.message) throw new Error(response.error.message);
};
