import { DURATION_FROM_SECONDS } from "@/constants/time";
import { getStories } from "@/modules/stories/queries/get-stories";
import { ListSprintStories } from "@/modules/teams/sprints/stories/list-stories";

export default async function Page(props: {
  params: Promise<{
    teamId: string;
    sprintId: string;
  }>;
}) {
  const params = await props.params;

  const { teamId, sprintId } = params;

  const stories = await getStories(
    {
      teamId,
      sprintId,
    },
    {
      next: {
        revalidate: DURATION_FROM_SECONDS.MINUTE * 5,
        tags: [`stories-team-${teamId}-sprint-${sprintId}`],
      },
    },
  );

  return <ListSprintStories stories={stories!} />;
}
