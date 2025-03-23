import { ListStories } from "@/modules/objectives/stories/list-stories";

export default async function Page(props: {
  params: Promise<{
    teamId: string;
    objectiveId: string;
  }>;
}) {
  const params = await props.params;

  const { objectiveId } = params;

  return <ListStories objectiveId={objectiveId} />;
}
