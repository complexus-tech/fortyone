import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Memory } from "../types";

export const getMemory = async (session: Session) => {
  const memory = await get<ApiResponse<Memory>>("users/memory", session);
  return memory.data!;
};
