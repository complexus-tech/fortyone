import { auth } from "@/auth";
import { getStories } from "@/modules/stories/queries/get-stories";

export const getMyStories = async () => {
  const session = await auth();
  const stories = await getStories({
    assigneeId: session?.user?.id,
    createdById: session?.user?.id,
  });
  return stories;
};
