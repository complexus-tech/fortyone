import type { Story } from "@/modules/stories/types";
import { ListStories } from "@/modules/teams/stories/list-stories";
import { getTeamStories } from "@/modules/teams/stories/queries/get-stories";
// import { getStories } from "@/actions/stories/get-stories";
// import { getStates } from "@/actions/states/get-states";

export default async function Page() {
  // const stories = await getStories();
  // const states = await getStates();
  // console.log(states);
  const stories = await getTeamStories();
  return <ListStories stories={stories} />;
}
