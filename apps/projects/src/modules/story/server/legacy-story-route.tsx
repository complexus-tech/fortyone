import { redirect } from "next/navigation";
import { StoryRouteNotFound } from "@/modules/story/components/story-route-not-found";
import { getStoryPath } from "@/shared/routing/story";
import { withWorkspacePath } from "@/utils";
import { getStoryRouteData } from "./get-story-route-data";

export const LegacyStoryRoute = async ({
  identifier,
  workspaceSlug,
}: {
  identifier: string;
  workspaceSlug: string;
}) => {
  const story = await getStoryRouteData(identifier, workspaceSlug);

  if (!story) {
    return <StoryRouteNotFound workspaceSlug={workspaceSlug} />;
  }

  redirect(withWorkspacePath(getStoryPath(story), workspaceSlug));
};
