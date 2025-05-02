import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse, WorkspaceSettings } from "@/types";

export const getWorkspaceSettings = async (session: Session) => {
  const settings = await get<ApiResponse<WorkspaceSettings>>(
    "settings",
    session,
  );
  return settings.data!;
};
