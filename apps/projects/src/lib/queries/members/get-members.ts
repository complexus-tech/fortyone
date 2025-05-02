import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse, Member } from "@/types";

export const getMembers = async (session: Session) => {
  const members = await get<ApiResponse<Member[]>>("members", session);
  return members.data!;
};

export const getTeamMembers = async (session: Session, teamId: string) => {
  const members = await get<ApiResponse<Member[]>>(
    `members?teamId=${teamId}`,
    session,
  );
  return members.data!;
};
