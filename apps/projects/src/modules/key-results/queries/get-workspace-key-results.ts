import type { Session } from "next-auth";
import { stringify } from "qs";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { KeyResultListResponse, KeyResultFilters } from "../types";

export const getWorkspaceKeyResults = async (
  session: Session,
  filters?: KeyResultFilters,
) => {
  const query = stringify(filters, {
    skipNulls: true,
    addQueryPrefix: true,
    encodeValuesOnly: true,
    arrayFormat: "comma",
  });

  const keyResults = await get<ApiResponse<KeyResultListResponse>>(
    `key-results${query}`,
    session,
  );

  return keyResults.data!;
};
