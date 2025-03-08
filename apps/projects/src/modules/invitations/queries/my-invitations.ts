import ky from "ky";
import type { ApiResponse } from "@/types";
import { auth } from "@/auth";
import type { Invitation } from "../types";

const apiURL = process.env.NEXT_PUBLIC_API_URL;

export async function getMyInvitations() {
  const session = await auth();
  if (!session) {
    return [];
  }
  try {
    const response = await ky.get(`${apiURL}/users/me/invitations`, {
      headers: {
        Authorization: `Bearer ${session.token}`,
      },
    });
    const invitations = await response.json<ApiResponse<Invitation[]>>();
    return invitations.data ?? [];
  } catch (error) {
    return [];
  }
}
