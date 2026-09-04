import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types/api-response";
import type {
  GoogleDriveFileReference,
  GoogleDriveIntegration,
  GoogleDriveTarget,
} from "@/shared/google-drive/types";

export const getGoogleDriveIntegration = async (ctx: WorkspaceCtx) => {
  const response = await get<ApiResponse<GoogleDriveIntegration>>(
    "integrations/google-drive",
    ctx,
  );
  return response.data!;
};

export const getGoogleDriveFiles = async (
  ctx: WorkspaceCtx,
  target: GoogleDriveTarget,
) => {
  const searchParams = new URLSearchParams({
    targetId: target.id,
    targetType: target.type,
  });
  const response = await get<ApiResponse<GoogleDriveFileReference[]>>(
    `google-drive/files?${searchParams.toString()}`,
    ctx,
  );
  return response.data ?? [];
};
