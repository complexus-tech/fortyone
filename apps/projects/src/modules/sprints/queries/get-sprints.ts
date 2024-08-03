import { get } from "@/lib/http";
import { Sprint } from "../types";

export const getSprints = async () => {
  const sprints = await get<Sprint[]>(`/sprints`);
  return sprints;
};
