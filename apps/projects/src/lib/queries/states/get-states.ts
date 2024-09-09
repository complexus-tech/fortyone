"use server";
import { statusTags } from "@/constants/keys";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import { get } from "@/lib/http";
import { State } from "@/types/states";

export const getStatuses = async () => {
  const statuses = await get<State[]>("states", {
    next: {
      revalidate: DURATION_FROM_SECONDS.MINUTE * 5,
      tags: [statusTags.lists()],
    },
  });
  return statuses;
};
