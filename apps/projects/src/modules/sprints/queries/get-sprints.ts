import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Sprint, SprintsPage } from "../types";

const emptySprintsPage = (page = 1, pageSize = 15): SprintsPage => ({
  sprints: [],
  pagination: {
    page,
    pageSize,
    hasMore: false,
    nextPage: 0,
  },
});

export const getSprintsPage = async (
  ctx: WorkspaceCtx,
  search = "",
  page = 1,
  pageSize = 15,
) => {
  const params = new URLSearchParams({
    page: String(page),
    pageSize: String(pageSize),
  });
  if (search.trim()) {
    params.set("search", search.trim());
  }

  const sprints = await get<ApiResponse<SprintsPage>>(
    `sprints?${params.toString()}`,
    ctx,
  );
  return sprints.data ?? emptySprintsPage(page, pageSize);
};

export const getSprints = async (ctx: WorkspaceCtx) => {
  const sprints = await get<ApiResponse<Sprint[]>>("sprints", ctx);
  return sprints.data!;
};
