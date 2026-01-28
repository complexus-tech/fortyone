"use server";

import type { ApiResponse, User } from "@/types";
import ky from "ky";
import { getApiUrl } from "@/lib/api-url";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

const apiURL = getApiUrl();

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
