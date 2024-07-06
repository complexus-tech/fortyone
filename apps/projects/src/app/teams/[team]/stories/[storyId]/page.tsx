import { MainStory } from "@/modules/teams/story";
import { getStory } from "@/modules/teams/story/queries/get-story";

type Props = {
  params: {
    storyId: string;
  };
};
export default async function Page({ params: { storyId } }: Props) {
  const story = await getStory(storyId);

  return <MainStory story={story} />;
}
