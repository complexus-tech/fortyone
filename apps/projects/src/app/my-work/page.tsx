import { ListMyStories } from "@/modules/my-work";
import { getMyStories } from "@/modules/my-work/queries/get-stories";

export default async function Page() {
  const stories = await getMyStories();

  return <ListMyStories stories={stories} />;
}
