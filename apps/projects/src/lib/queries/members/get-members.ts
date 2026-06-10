import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse, Member } from "@/types";

const buildMembersQuery = ({
  search,
  teamId,
}: {
  search?: string;
  teamId?: string;
}) => {
  const params = new URLSearchParams();
  if (teamId) {
    params.set("teamId", teamId);
  }
  if (search?.trim()) {
    params.set("search", search.trim());
  }

  const query = params.toString();
  return query ? `?${query}` : "";
};

export const getMembers = async (ctx: WorkspaceCtx, search?: string) => {
  const members = await get<ApiResponse<Member[]>>(
    `members${buildMembersQuery({ search })}`,
    ctx,
  );
  return members.data!;
};

export const getTeamMembers = async (
  teamId: string,
  ctx: WorkspaceCtx,
  search?: string,
) => {
  const members = await get<ApiResponse<Member[]>>(
    `members${buildMembersQuery({ search, teamId })}`,
    ctx,
  );
  return members.data!;
};
