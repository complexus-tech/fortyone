"use server";

import { auth } from "@/auth";
import { post, remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type {
  GoogleDriveFileReference,
  GoogleDriveFileType,
  GoogleDriveImportResult,
  GoogleDriveImportVisibility,
  GoogleDrivePickerFile,
  GoogleDrivePickerSession,
  GoogleDriveTarget,
} from "./types";

const context = async (workspaceSlug: string) => ({
  session: (await auth())!,
  workspaceSlug,
});

export const createGoogleDriveConnectSessionAction = async (
  workspaceSlug: string,
  returnUrl?: string,
) => {
  try {
    return await post<
      { returnUrl?: string },
      ApiResponse<{ authorizationUrl: string }>
    >(
      "integrations/google-drive/connect-session",
      returnUrl ? { returnUrl } : {},
      await context(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const disconnectGoogleDriveAction = async (workspaceSlug: string) => {
  try {
    return await remove<ApiResponse<null>>(
      "integrations/google-drive",
      await context(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const createGoogleDrivePickerSessionAction = async (
  workspaceSlug: string,
) => {
  try {
    return await post<
      Record<string, never>,
      ApiResponse<GoogleDrivePickerSession>
    >(
      "integrations/google-drive/picker-session",
      {},
      await context(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const attachGoogleDriveFilesAction = async (
  workspaceSlug: string,
  target: GoogleDriveTarget,
  files: GoogleDrivePickerFile[],
) => {
  try {
    return await post<
      {
        targetId: string;
        targetType: GoogleDriveTarget["type"];
        files: GoogleDrivePickerFile[];
      },
      ApiResponse<GoogleDriveFileReference[]>
    >(
      "google-drive/files",
      { files, targetId: target.id, targetType: target.type },
      await context(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const deleteGoogleDriveFileAction = async (
  workspaceSlug: string,
  referenceId: string,
) => {
  try {
    return await remove<ApiResponse<null>>(
      `google-drive/files/${referenceId}`,
      await context(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const refreshGoogleDriveFileAction = async (
  workspaceSlug: string,
  referenceId: string,
) => {
  try {
    return await post<
      Record<string, never>,
      ApiResponse<GoogleDriveFileReference>
    >(
      `google-drive/files/${referenceId}/refresh`,
      {},
      await context(workspaceSlug),
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const createGoogleDriveFileAction = async (
  workspaceSlug: string,
  target: GoogleDriveTarget,
  fileType: GoogleDriveFileType,
  title: string,
  idempotencyKey: string,
) => {
  try {
    return await post<
      {
        targetId: string;
        targetType: GoogleDriveTarget["type"];
        fileType: GoogleDriveFileType;
        title: string;
      },
      ApiResponse<GoogleDriveFileReference>
    >(
      "google-drive/files/create",
      { fileType, targetId: target.id, targetType: target.type, title },
      await context(workspaceSlug),
      { headers: { "Idempotency-Key": idempotencyKey } },
    );
  } catch (error) {
    return getApiError(error);
  }
};

export const importGoogleDriveFileAction = async (
  workspaceSlug: string,
  referenceId: string,
  visibility: GoogleDriveImportVisibility,
  idempotencyKey: string,
) => {
  try {
    return await post<
      { visibility: GoogleDriveImportVisibility },
      ApiResponse<GoogleDriveImportResult>
    >(
      `google-drive/files/${referenceId}/imports`,
      { visibility },
      await context(workspaceSlug),
      { headers: { "Idempotency-Key": idempotencyKey } },
    );
  } catch (error) {
    return getApiError(error);
  }
};
