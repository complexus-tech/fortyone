import { ListSprintStories } from "@/modules/sprints/stories/list-stories";

export default async function Page(props: {
  params: Promise<{
    teamId: string;
    sprintId: string;
  }>;
}) {
  const params = await props.params;

  return <ListSprintStories sprintId={params.sprintId} />;
}
