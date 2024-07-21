import { StoryPage } from "@/modules/story";
import { getStory } from "@/modules/story/queries/get-story";

type Props = {
  params: {
    storyId: string;
  };
};
export default async function Page({ params: { storyId } }: Props) {
  const story = await getStory(storyId);

  return <StoryPage story={story} />;
}
