"use server";
import type { ApiResponse, User } from "@/types";
import { get } from "@/lib/http";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import { userTags } from "@/constants/keys";

export async function getProfile() {
  const res = await get<ApiResponse<User>>("profile", {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 10,
      tags: [userTags.profile()],
    },
  });

  return res.data;
}
