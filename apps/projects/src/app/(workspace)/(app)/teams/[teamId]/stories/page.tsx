import type { Metadata } from "next";
import { ListStories } from "@/modules/teams/stories/list-stories";

export const metadata: Metadata = {
  title: "Stories",
};

export default async function Page(props: {
  params: Promise<{
    teamId: string;
  }>;
}) {
  const params = await props.params;

  const { teamId } = params;

  return <ListStories teamId={teamId} />;
}
