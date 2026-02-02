"use server";

import type { ApiResponse, User } from "@/types";
import ky from "ky";
import { getApiUrl } from "@/lib/api-url";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

const apiURL = getApiUrl();

export const uploadProfileImageAction = async (file: File) => {
  try {
    const formData = new FormData();
    formData.append("image", file);
    const session = await auth();
    const res = await ky
      .post(`${apiURL}/users/profile/image`, {
        body: formData,
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
