import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { DocumentRelationType, WorkspaceDocument } from "./types";

export const getDocuments = async (
  ctx: WorkspaceCtx,
  search = "",
  scope = "all",
) => {
  const params = new URLSearchParams({ scope });
  if (search.trim()) params.set("search", search.trim());
  const response = await get<ApiResponse<WorkspaceDocument[]>>(
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
