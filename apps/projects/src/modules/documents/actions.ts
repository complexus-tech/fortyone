import { auth } from "@/auth";
import { getApiUrl } from "@/lib/api-url";
import { post, put, remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type {
  DocumentAccessUpdate,
  DocumentCreate,
  DocumentMedia,
  DocumentRelationType,
  DocumentUpdate,
  RelatedWork,
  WorkspaceDocument,
} from "./types";

const DOCUMENT_MEDIA_UPLOAD_TIMEOUT_MS = 5 * 60 * 1000;

const resolveDocumentMediaUrl = (url: string) => {
  if (/^https?:\/\//i.test(url)) return url;
  return `${getApiUrl()}${url.startsWith("/") ? url : `/${url}`}`;
};

const workspaceContext = async (workspaceSlug: string) => {
  const session = await auth();
  return { session: session!, workspaceSlug };
};

export const createDocumentAction = async (
  workspaceSlug: string,
  input: DocumentCreate = {},
) => {
  try {
    return await post<DocumentCreate, ApiResponse<WorkspaceDocument>>(
      "documents",
      {
        title: input.title ?? "Untitled document",
        visibility: input.visibility ?? "workspace",
        contentHtml: input.contentHtml ?? "",
        contentText: input.contentText ?? "",
      },
      await workspaceContext(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const duplicateDocumentAction = async (
  documentId: string,
  workspaceSlug: string,
) => {
  try {
    return await post<Record<string, never>, ApiResponse<WorkspaceDocument>>(
      `documents/${documentId}/duplicate`,
      {},
      await workspaceContext(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const updateDocumentAction = async (
  documentId: string,
  payload: DocumentUpdate,
  workspaceSlug: string,
) => {
  try {
    return await put<DocumentUpdate, ApiResponse<WorkspaceDocument>>(
      `documents/${documentId}`,
      payload,
      await workspaceContext(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const updateDocumentAccessAction = async (
  documentId: string,
  payload: DocumentAccessUpdate,
  workspaceSlug: string,
) => {
  try {
    return await put<DocumentAccessUpdate, ApiResponse<WorkspaceDocument>>(
      `documents/${documentId}/access`,
      payload,
      await workspaceContext(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const archiveDocumentAction = async (
  documentId: string,
  workspaceSlug: string,
) => {
  try {
    return await remove<ApiResponse<null>>(
      `documents/${documentId}`,
      await workspaceContext(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const deleteDocumentAction = async (
  documentId: string,
  workspaceSlug: string,
) => {
  try {
    return await remove<ApiResponse<null>>(
      `documents/${documentId}/permanent`,
      await workspaceContext(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const uploadDocumentMediaAction = async (
  documentId: string,
  file: File,
  workspaceSlug: string,
) => {
  try {
    const formData = new FormData();
    formData.append("file", file);
    const response = await post<FormData, ApiResponse<DocumentMedia>>(
      `documents/${documentId}/media`,
      formData,
      await workspaceContext(workspaceSlug),
      { timeout: DOCUMENT_MEDIA_UPLOAD_TIMEOUT_MS },
    );
    if (response.data) {
      response.data.url = resolveDocumentMediaUrl(response.data.url);
    }
    return response;
  } catch (error) {
    return getApiError(error);
  }
};

export const deleteDocumentMediaAction = async (
  documentId: string,
  attachmentId: string,
  workspaceSlug: string,
) => {
  try {
    return await remove<ApiResponse<null>>(
      `documents/${documentId}/media/${attachmentId}`,
      await workspaceContext(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const addDocumentRelationshipAction = async (
  documentId: string,
  payload: { entityType: DocumentRelationType; entityId: string },
  workspaceSlug: string,
) => {
  try {
    return await post<typeof payload, ApiResponse<RelatedWork>>(
      `documents/${documentId}/relationships`,
      payload,
      await workspaceContext(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const removeDocumentRelationshipAction = async (
  documentId: string,
  entityType: DocumentRelationType,
  entityId: string,
  workspaceSlug: string,
) => {
  try {
    return await remove<ApiResponse<null>>(
      `documents/${documentId}/relationships/${entityType}/${entityId}`,
      await workspaceContext(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};
