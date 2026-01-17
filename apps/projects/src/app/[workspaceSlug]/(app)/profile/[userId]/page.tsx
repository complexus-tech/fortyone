import type { Metadata } from "next";
import { ListUserStories } from "@/modules/profile";
import { getStories } from "@/modules/stories/queries/get-stories";
import { auth } from "@/auth";

export const metadata: Metadata = {
  title: "Profile",
};

export default async function Page(props: {
  params: Promise<{
    userId: string;
  }>;
}) {
  const params = await props.params;
  const session = await auth();

  const { userId } = params;

  const stories = await getStories(session!, {
    reporterId: userId,
    assigneeId: userId,
  });

  return <ListUserStories stories={stories} />;
}
