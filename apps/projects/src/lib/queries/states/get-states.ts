import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { State } from "@/types/states";

export const getStatuses = async () => {
  const statuses = await get<ApiResponse<State[]>>("states");
  return statuses.data!;
};
