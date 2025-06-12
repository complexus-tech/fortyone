"use server";

import ky from "ky";
import type { ApiResponse, User } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

const apiURL = process.env.NEXT_PUBLIC_API_URL;

export const deleteProfileImageAction = async () => {
  try {
    const session = await auth();
    const res = await ky
      .delete(`${apiURL}/users/profile/image`, {
        headers: {
          Authorization: `Bearer ${session?.token}`,
        },
      })
      .json<ApiResponse<User>>();
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
