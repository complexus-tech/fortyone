import { ResourceNotFoundState } from "@/components/ui/resource-not-found-state";
import { withWorkspacePath } from "@/utils";

export const StoryRouteNotFound = ({
  workspaceSlug,
}: {
  workspaceSlug: string;
}) => (
  <ResourceNotFoundState
    description="This item might not exist or you do not have access to it."
    href={withWorkspacePath("/my-work", workspaceSlug)}
    title="404: Item not found"
  />
);
