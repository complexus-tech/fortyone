import { stringify } from "qs";
import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse, Label } from "@/types";

export const getLabels = async (
  session: Session,
  params: {
    teamId?: string;
  } = {},
) => {
  const query = stringify(params, {
    skipNulls: true,
    addQueryPrefix: true,
    encodeValuesOnly: true,
  });
  const labels = await get<ApiResponse<Label[]>>(`labels${query}`, session);
  return labels.data!;
};
