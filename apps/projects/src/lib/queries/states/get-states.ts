import { get } from "@/lib/http";
import { State } from "@/types/states";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import { TAGS } from "@/constants/tags";
import { auth } from "@/auth";

export const getStates = async () => {
  const session = await auth();
  const states = await get<State[]>(`/states`, {
    // next: {
    //   revalidate: DURATION_FROM_SECONDS.SECOND * 1,
    //   tags: [TAGS.states],
    // },
  });
  return states;
};
