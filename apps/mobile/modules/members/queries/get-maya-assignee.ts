import type { ApiResponse, User } from "@/types";
import { get } from "@/lib/http";

type MayaAssignee = Pick<
  User,
  "id" | "username" | "fullName" | "avatarUrl" | "isActive"
> & { isSystem: boolean };

export async function getMayaAssignee(signal?: AbortSignal) {
  const response = await get<ApiResponse<MayaAssignee>>("members/maya", {
    signal,
  });
  return response.data ?? null;
}
