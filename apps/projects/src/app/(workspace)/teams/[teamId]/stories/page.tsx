import { ListStories } from "@/modules/teams/stories/list-stories";
import { getStories } from "@/modules/stories/queries/get-stories";

export default async function Page({
  params: { teamId },
}: {
  params: {
    teamId: string;
  };
}) {
  const stories = await getStories({ teamId });
  return <ListStories stories={stories} />;
}
