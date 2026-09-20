import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type {
  DocumentRelationType,
  DocumentCommentThread,
  DocumentRevision,
  WorkspaceDocument,
  WorkspaceDocumentSummary,
} from "./types";

export const getDocuments = async (
  ctx: WorkspaceCtx,
  search = "",
  scope = "all",
  limit?: number,
) => {
  const params = new URLSearchParams({ scope });
  if (search.trim()) params.set("search", search.trim());
  if (limit !== undefined) params.set("limit", String(limit));
  const response = await get<ApiResponse<WorkspaceDocumentSummary[]>>(
    `documents?${params.toString()}`,
    ctx,
  );
  return response.data ?? [];
};

export const getDocument = async (documentId: string, ctx: WorkspaceCtx) => {
  const response = await get<ApiResponse<WorkspaceDocument>>(
    `documents/${documentId}`,
    ctx,
  );
  return response.data!;
};

export const getDocumentComments = async (
  documentId: string,
  ctx: WorkspaceCtx,
) => {
  const response = await get<ApiResponse<DocumentCommentThread[]>>(
    `documents/${documentId}/comments`,
    ctx,
  );
  return response.data ?? [];
};

export const getRelatedDocuments = async (
  entityType: DocumentRelationType,
  entityId: string,
  ctx: WorkspaceCtx,
) => {
  const params = new URLSearchParams({ entityType, entityId });
  const response = await get<ApiResponse<WorkspaceDocument[]>>(
    `documents/related?${params.toString()}`,
    ctx,
  );
  return response.data ?? [];
};

export const getDocumentRevisions = async (
  documentId: string,
  ctx: WorkspaceCtx,
  before = 0,
) => {
  const response = await get<ApiResponse<DocumentRevision[]>>(
    `documents/${documentId}/revisions${before ? `?before=${before}` : ""}`,
    ctx,
  );
  return response.data ?? [];
};

export const getDocumentRevision = async (
  documentId: string,
  revision: number,
  ctx: WorkspaceCtx,
) => {
  const response = await get<ApiResponse<DocumentRevision>>(
    `documents/${documentId}/revisions/${revision}`,
    ctx,
  );
  return response.data!;
};
