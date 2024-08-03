import { get } from "@/lib/http";
import { Sprint } from "@/modules/sprints/types";

export const getRunningSprints = async () => {
  const sprints = await get<Sprint[]>(`/sprints`);
  return sprints;
};
