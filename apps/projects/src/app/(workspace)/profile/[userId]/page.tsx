import { ListUserStories } from "@/modules/profile";
import { getStories } from "@/modules/stories/queries/get-stories";

export default async function Page(props: {
  params: Promise<{
    userId: string;
  }>;
}) {
  const params = await props.params;

  const { userId } = params;

  const stories = await getStories({
    reporterId: userId,
    assigneeId: userId,
  });

  return <ListUserStories stories={stories!} />;
}
