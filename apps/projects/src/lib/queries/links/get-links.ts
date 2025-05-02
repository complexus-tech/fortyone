import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse, Link } from "@/types";

export const getLinks = async (storyId: string, session: Session) => {
  const links = await get<ApiResponse<Link[]>>(
    `stories/${storyId}/links`,
    session,
  );
  return links.data!;
};
