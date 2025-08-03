import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse, Workspace } from "@/types";

export const getWorkspace = async (session: Session) => {
  const workspace = await get<ApiResponse<Workspace>>("", session);
  return workspace.data!;
};
