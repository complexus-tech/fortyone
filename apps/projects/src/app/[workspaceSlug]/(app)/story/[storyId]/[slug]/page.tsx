import { LegacyStoryRoute } from "@/modules/story/server/legacy-story-route";

type Props = {
  params: Promise<{
    storyId: string;
    workspaceSlug: string;
  }>;
};
export default async function Page(props: Props) {
  const { storyId, workspaceSlug } = await props.params;
  return (
    <LegacyStoryRoute identifier={storyId} workspaceSlug={workspaceSlug} />
  );
}
