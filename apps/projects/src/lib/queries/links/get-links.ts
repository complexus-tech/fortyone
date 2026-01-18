import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse, Link } from "@/types";

export const getLinks = async (storyId: string, ctx: WorkspaceCtx) => {
  const links = await get<ApiResponse<Link[]>>(
    `stories/${storyId}/links`,
    ctx,
  );
  return links.data!;
};
