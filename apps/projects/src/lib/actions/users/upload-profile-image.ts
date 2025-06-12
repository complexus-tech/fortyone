"use server";

import { post } from "@/lib/http";
import type { ApiResponse, User } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const uploadProfileImageAction = async (file: File) => {
  try {
    const formData = new FormData();
    formData.append("image", file);
    const session = await auth();
    const res = await post<FormData, ApiResponse<User>>(
      "logo",
      formData,
      session!,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
