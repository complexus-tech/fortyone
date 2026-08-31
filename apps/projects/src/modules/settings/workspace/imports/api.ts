import type { WorkspaceCtx } from "@/lib/http";
import { get, post } from "@/lib/http";
import type { ApiResponse, Member } from "@/types";
import type { State } from "@/types/states";
import type { Team, CreateTeamInput } from "@/modules/teams/types";
import type {
  ImportAnalysisPollResponse,
  ImportAnalysisStartResponse,
  ImportPriority,
} from "./schema";

export type ImportStoryPayload = {
  title: string;
  description: string;
  teamId: string;
  statusId?: string;
  assigneeId?: string;
  priority: ImportPriority;
  startDate?: string;
  endDate?: string;
};

export type ImportStoriesRequest = {
  provider: "jira_csv" | "file";
  sourceDigest: string;
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
}: ImportStoriesRequest): ImportStoriesRequest[] => {
  const requests: ImportStoriesRequest[] = [];
  let batch: ImportStoriesRequest["items"] = [];

  for (const item of items) {
    const candidate = [...batch, item];
    const candidateRequest = { items: candidate, provider, sourceDigest };
    const candidateBytes = new TextEncoder().encode(
      JSON.stringify(candidateRequest),
    ).byteLength;

    if (
      batch.length > 0 &&
      (candidate.length > IMPORT_BATCH_MAX_ITEMS ||
        candidateBytes > IMPORT_BATCH_TARGET_BYTES)
    ) {
      requests.push({ items: batch, provider, sourceDigest });
      batch = [item];
    } else {
      batch = candidate;
    }
  }

  if (batch.length > 0) {
    requests.push({ items: batch, provider, sourceDigest });
  }

  return requests;
};

const readError = async (response: Response, fallback: string) => {
  const message = await response.text();
  return new Error(message.trim() || fallback);
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
  if (response.error?.message || !response.data) {
    throw new Error(
      response.error?.message || "Unable to load the destination workflow",
    );
  }
  return response.data;
};

export const getImportTeamMembers = async (
  teamId: string,
  ctx: WorkspaceCtx,
) => {
  const response = await get<ApiResponse<Member[]>>(
    `members?teamId=${teamId}`,
    ctx,
  );
  if (response.error?.message || !response.data) {
    throw new Error(
      response.error?.message || "Unable to load the destination team members",
    );
  }
  return response.data;
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
