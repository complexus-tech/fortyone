import ky from "ky";
import type { Session } from "next-auth";
import type { ApiResponse, User } from "@/types";

const apiURL = process.env.NEXT_PUBLIC_API_URL;

export async function getProfile(session: Session) {
  const res = await ky.get(`${apiURL}/users/profile`, {
    headers: {
      Authorization: `Bearer ${session.token}`,
    },
  });
  const data = await res.json<ApiResponse<User>>();
  return data.data!;
}
