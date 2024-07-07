import { MainStory } from "@/modules/stories/detail";
import { getStory } from "@/modules/stories/queries/get-story";

type Props = {
  params: {
    storyId: string;
  };
};
export default async function Page({ params: { storyId } }: Props) {
  const story = await getStory(storyId);

  return <MainStory story={story} />;
}
