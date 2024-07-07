import { ListUserStories } from "@/modules/profile";
import { getProfileStories } from "@/modules/profile/queries/get-stories";

export default async function Page() {
  const stories = await getProfileStories();

  return <ListUserStories stories={stories} user="josemukorivo" />;
}
