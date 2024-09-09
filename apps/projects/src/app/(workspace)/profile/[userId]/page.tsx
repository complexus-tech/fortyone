import { ListUserStories } from "@/modules/profile";
import { getStories } from "@/modules/stories/queries/get-stories";

export default async function Page({
  params: { userId },
}: {
  params: {
    userId: string;
  };
}) {
  const stories = await getStories({
    reporterId: userId,
    assigneeId: userId,
  });

  return <ListUserStories stories={stories} />;
}
