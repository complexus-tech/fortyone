"use server";
import ky from "ky";
import type { ApiResponse } from "@/types";
import { auth } from "@/auth";
import { requestError } from "../fetch-error";
import type { Invitation } from "@/modules/invitations/types";

const apiUrl = process.env.NEXT_PUBLIC_API_URL;

export async function getMyInvitations() {
  const session = await auth();
  if (!session) {
    return {
      data: [],
    } as ApiResponse<Invitation[]>;
  }
  try {
    const response = await ky.get(`${apiUrl}/users/me/invitations`, {
      headers: {
        Authorization: `Bearer ${session?.token}`,
      },
    });
    const data = await response.json<ApiResponse<Invitation[]>>();
    return data;
  } catch (error) {
    const res = await requestError<Invitation[]>(error);
    return {
      data: null,
      error: {
        message: res.error?.message || "Failed to verify invitation",
      },
    };
  }
}
