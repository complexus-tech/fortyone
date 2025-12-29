import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Memory } from "../types";

export const getMemories = async (session: Session) => {
  const memories = await get<ApiResponse<Memory[]>>("users/memory", session);
  return memories.data!;
};
