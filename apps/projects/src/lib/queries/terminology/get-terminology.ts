"use server";

import { get } from "@/lib/http";
import type { ApiResponse, Terminology } from "@/types";

export const getTerminology = async () => {
  const terminology = await get<ApiResponse<Terminology>>("terminology");
  return terminology.data!;
};
