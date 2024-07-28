import { ListStories } from "@/modules/teams/stories/list-stories";
import { getStories } from "@/modules/stories/queries/get-stories";

export default async function Page() {
  const stories = await getStories();
  return <ListStories stories={stories} />;
}
