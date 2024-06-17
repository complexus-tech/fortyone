import { MainStory } from "@/modules/teams/story";
import { getStory } from "@/modules/teams/story/actions/get-story";

export default async function Page() {
  const story = await getStory("d00525f6-0f50-4540-9351-154fb69f051f");

  return <MainStory story={story} />;
}
