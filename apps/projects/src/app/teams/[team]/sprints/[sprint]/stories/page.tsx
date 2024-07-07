import { getStories } from "@/modules/stories/queries/get-stories";
import { ListSprintStories } from "@/modules/teams/sprints/stories/list-stories";

export default async function Page() {
  const stories = await getStories({
    sprintId: "456982345-4325-2345-23453245",
  });

  return <ListSprintStories stories={stories} />;
}
