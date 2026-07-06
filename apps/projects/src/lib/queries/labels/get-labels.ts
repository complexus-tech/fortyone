import { stringify } from "qs";
import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse, Label, LabelsPage } from "@/types";

const emptyLabelsPage = (page = 1, pageSize = 15): LabelsPage => ({
  labels: [],
  pagination: {
    page,
    pageSize,
    hasMore: false,
    nextPage: 0,
  },
});

export const getLabels = async (
  ctx: WorkspaceCtx,
  params: {
    search?: string;
    teamId?: string;
  } = {},
) => {
  const query = stringify(params, {
    skipNulls: true,
    addQueryPrefix: true,
    encodeValuesOnly: true,
  });
  const labels = await get<ApiResponse<Label[]>>(`labels${query}`, ctx);
  return labels.data!;
};

export const getLabelsPage = async (
  ctx: WorkspaceCtx,
  params: {
    page?: number;
    pageSize?: number;
    search?: string;
    teamId?: string;
  } = {},
) => {
  const page = params.page ?? 1;
  const pageSize = params.pageSize ?? 15;
  const query = stringify(
    {
      ...params,
      page,
      pageSize,
      search: params.search?.trim() || undefined,
    },
    {
      skipNulls: true,
      addQueryPrefix: true,
      encodeValuesOnly: true,
    },
  );
  const labels = await get<ApiResponse<LabelsPage>>(`labels${query}`, ctx);
  return labels.data ?? emptyLabelsPage(page, pageSize);
};
