import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse, Member } from "@/types";

export const getMembers = async (ctx: WorkspaceCtx) => {
  const members = await get<ApiResponse<Member[]>>("members", ctx);
  return members.data!;
};

export const getTeamMembers = async (teamId: string, ctx: WorkspaceCtx) => {
  const members = await get<ApiResponse<Member[]>>(
    `members?teamId=${teamId}`,
    ctx,
  );
  return members.data!;
};
