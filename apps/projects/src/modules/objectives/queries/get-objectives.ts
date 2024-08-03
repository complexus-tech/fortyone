import { get } from "@/lib/http";
import { Objective } from "../types";

export const getObjectives = async () => {
  const objectives = await get<Objective[]>(`/objectives`);
  return objectives;
};
