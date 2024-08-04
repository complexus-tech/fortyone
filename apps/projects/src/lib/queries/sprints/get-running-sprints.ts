import { sprintTags } from "@/constants/keys";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import { get } from "@/lib/http";
import { Sprint } from "@/modules/sprints/types";

export const getRunningSprints = async () => {
  const sprints = await get<Sprint[]>("sprints", {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 5,
      tags: [sprintTags.lists()],
    },
  });
  return sprints;
};
