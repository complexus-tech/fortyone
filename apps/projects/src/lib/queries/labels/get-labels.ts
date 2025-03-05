"use server";
import { stringify } from "qs";
import { get } from "@/lib/http";
import type { ApiResponse, Label } from "@/types";

export const getLabels = async (
  params: {
    teamId?: string;
  } = {},
) => {
  const query = stringify(params, {
    skipNulls: true,
    addQueryPrefix: true,
    encodeValuesOnly: true,
  });
  const labels = await get<ApiResponse<Label[]>>(`labels${query}`);
  return labels.data!;
};
