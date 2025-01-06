"use server";
import { labelTags } from "@/constants/keys";
import { get } from "@/lib/http";
import { ApiResponse, Label } from "@/types";
import qs from "qs";
import { DURATION_FROM_SECONDS } from "@/constants/time";

export const getLabels = async (
  params: {
    teamId?: string;
  } = {},
) => {
  const query = qs.stringify(params, {
    skipNulls: true,
    addQueryPrefix: true,
    encodeValuesOnly: true,
  });
  const labels = await get<ApiResponse<Label[]>>(`labels${query}`, {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 10,
      tags: [
        params?.teamId ? labelTags.team(params.teamId) : labelTags.lists(),
      ],
    },
  });
  return labels.data;
};
