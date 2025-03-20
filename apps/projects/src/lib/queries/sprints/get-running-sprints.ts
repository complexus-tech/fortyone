import { get } from "@/lib/http";
import type { Sprint } from "@/modules/sprints/types";
import type { ApiResponse } from "@/types";

export const getRunningSprints = async () => {
  const sprints = await get<ApiResponse<Sprint[]>>("sprints");
  return sprints.data!;
};
