import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import { sprintTags } from "@/constants/keys";
import type { Sprint } from "../types";

export const getSprints = async () => {
  const sprints = await get<ApiResponse<Sprint[]>>("sprints", {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 10,
      tags: [sprintTags.lists()],
    },
  });
  return sprints.data!;
};
