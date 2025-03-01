import ky from "ky";
import type { ApiResponse, Invitation } from "@/types";
import { requestError } from "../fetch-error";

const apiUrl = process.env.NEXT_PUBLIC_API_URL;

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
