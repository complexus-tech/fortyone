import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse, Member, MembersPage } from "@/types";

const buildMembersQuery = ({
  page,
  pageSize,
  search,
  teamId,
}: {
  page?: number;
  pageSize?: number;
  search?: string;
  teamId?: string;
}) => {
  const params = new URLSearchParams();
  if (page) {
    params.set("page", String(page));
  }
  if (pageSize) {
    params.set("pageSize", String(pageSize));
  }
  if (teamId) {
    params.set("teamId", teamId);
  }
  if (search?.trim()) {
    params.set("search", search.trim());
  }

  const query = params.toString();
  return query ? `?${query}` : "";
};

const emptyMembersPage = (page = 1, pageSize = 15): MembersPage => ({
  members: [],
  pagination: {
    page,
    pageSize,
    hasMore: false,
    nextPage: 0,
  },
});

export const getMembersPage = async (
  ctx: WorkspaceCtx,
  search = "",
  page = 1,
  pageSize = 15,
) => {
  const members = await get<ApiResponse<MembersPage>>(
    `members${buildMembersQuery({ page, pageSize, search })}`,
    ctx,
  );
  return members.data ?? emptyMembersPage(page, pageSize);
};

export const getMayaAssignee = async (ctx: WorkspaceCtx) => {
  const member = await get<ApiResponse<Member>>("members/maya", ctx);
  return member.data!;
};

export const getTeamMembersPage = async (
  teamId: string,
  ctx: WorkspaceCtx,
  search = "",
  page = 1,
  pageSize = 15,
) => {
  if (!teamId) return emptyMembersPage(page, pageSize);

  const members = await get<ApiResponse<MembersPage>>(
    `members${buildMembersQuery({ page, pageSize, search, teamId })}`,
    ctx,
  );
  return members.data ?? emptyMembersPage(page, pageSize);
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
