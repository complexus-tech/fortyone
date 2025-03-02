import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Invitation } from "../types";

export const getPendingInvitations = async () => {
  const response = await get<ApiResponse<Invitation[]>>("invitations");
  return response.data ?? [];
};
