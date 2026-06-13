import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Objective, ObjectivesPage } from "../types";

const emptyObjectivesPage = (page = 1, pageSize = 15): ObjectivesPage => ({
  objectives: [],
  pagination: {
    page,
    pageSize,
    hasMore: false,
    nextPage: 0,
  },
});

export const getObjectivesPage = async (
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

  const objectives = await get<ApiResponse<ObjectivesPage>>(
    `objectives?${params.toString()}`,
    ctx,
  );
  return objectives.data ?? emptyObjectivesPage(page, pageSize);
};

export const getObjectives = async (ctx: WorkspaceCtx, search = "") => {
  const params = new URLSearchParams();
  if (search.trim()) {
    params.set("search", search.trim());
  }
  const query = params.toString();
  const objectives = await get<ApiResponse<Objective[]>>(
    `objectives${query ? `?${query}` : ""}`,
    ctx,
  );
  return objectives.data ?? [];
};
