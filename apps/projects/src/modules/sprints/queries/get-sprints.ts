import { get } from "@/lib/http";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import { TAGS } from "@/constants/tags";
import { auth } from "@/auth";
import { Sprint } from "../types";

export const getSprints = async () => {
  const session = await auth();
  const sprints = await get<Sprint[]>(`/sprints`, {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 30,
      tags: [TAGS.sprints],
    },
  });
  return sprints;
};
