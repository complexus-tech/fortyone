import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Invitation } from "../types";

export const getPendingInvitations = async (session: Session) => {
  const response = await get<ApiResponse<Invitation[]>>("invitations", session);
  return response.data ?? [];
};
