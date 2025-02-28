"use server";

import ky from "ky";
import type { ApiResponse } from "@/types";
import { requestError } from "../fetch-error";

const apiUrl = process.env.NEXT_PUBLIC_API_URL;

export type Invitation = {
  id: string;
  workspaceId: string;
  inviterId: string;
  email: string;
  role: string;
  teamIds: string[];
  expiresAt: string;
  usedAt?: string;
  createdAt: string;
  updatedAt: string;
  workspaceName: string;
  workspaceSlug: string;
  workspaceColor: string;
};

export async function verifyInvitation(token: string) {
  try {
    const response = await ky.get(`${apiUrl}/invitations/${token}`);
    const data = await response.json<ApiResponse<Invitation>>();

    return data;
  } catch (error) {
    const res = await requestError<Invitation>(error);
    return {
      data: null,
      error: {
        message: res.error?.message || "Failed to verify invitation",
      },
    };
  }
}
