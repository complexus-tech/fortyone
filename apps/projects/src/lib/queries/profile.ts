import ky from "ky";
import type { Session } from "next-auth";
import { getApiUrl } from "@/lib/api-url";
import type { ApiResponse, User } from "@/types";

const apiURL = getApiUrl();

export async function getProfile(session: Session) {
  const res = await ky.get(`${apiURL}/users/profile`, {
    headers: {
      Authorization: `Bearer ${session.token}`,
    },
  });
  const data = await res.json<ApiResponse<User>>();
  return data.data!;
}
