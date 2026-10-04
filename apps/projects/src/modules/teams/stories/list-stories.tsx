"use client";
import type { ReactNode } from "react";
import { useState } from "react";
import { useParams, useSearchParams } from "next/navigation";
import { Box, Button, Text } from "ui";
import type { StoriesLayout } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { StoriesFilterBar } from "@/components/ui/stories-filter-bar";
import { BoardSkeleton } from "@/components/ui/board-skeleton";
import type {
  SavedViewConfiguration,
  SavedViewLoadState,
} from "@/shared/story/view-configuration";
import { TeamOptionsProvider, useTeamOptions } from "./provider";
import { Header } from "./header";
import type { SavedViewsAction } from "./header";
import { AllStories } from "./all-stories";

export type ViewCreatorAction = (props: {
  teamId: string;
  configuration: SavedViewConfiguration;
}) => ReactNode;

const ActiveViewCreator = ({
  layout,
  teamId,
  renderViewCreator,
}: {
  layout: StoriesLayout;
  teamId: string;
  renderViewCreator: ViewCreatorAction;
}) => {
  const { filters, viewOptions, viewMetadata } = useTeamOptions();
  return renderViewCreator({
    teamId,
    configuration: {
      ...viewMetadata,
      version: 1,
      layout,
      filters,
      viewOptions,
    },
  });
};

const savedViewMessage = (status: SavedViewLoadState["status"]) => {
  if (status === "loading") return "Loading saved view...";
  if (status === "unavailable")
    return "This view is no longer available. Choose another view or browse Views.";
  return "This view could not be loaded. Try again from the view menu.";
};

const ActiveStoriesFilterBar = ({
  layout,
  teamId,
}: {
  layout: StoriesLayout;
  teamId: string;
}) => {
  const { filters, resetFilters, setFilters, viewOptions, viewMetadata } =
    useTeamOptions();

  return (
    <StoriesFilterBar
      filters={filters}
      resetFilters={resetFilters}
      saveView={{
        teamId,
        configuration: {
          ...viewMetadata,
          version: 1,
          layout,
          filters,
          viewOptions,
        },
      }}
      setFilters={setFilters}
    />
  );
};

export const ListStories = ({
  renderSavedViews,
  renderViewCreator,
}: {
  renderSavedViews?: SavedViewsAction;
  renderViewCreator?: ViewCreatorAction;
}) => {
  const { teamId } = useParams<{ teamId: string }>();
  const { data: session } = useSession();
  const { workspaceSlug, withWorkspace } = useWorkspacePath();
  const params = useSearchParams();
  const viewId = params.get("view");
  const creating = params.get("createView") === "true";
  const [viewState, setViewState] = useState<SavedViewLoadState | null>(null);
  const currentState =
    viewState?.teamId === teamId &&
    viewState.viewId === viewId &&
    viewState.workspaceSlug === workspaceSlug &&
    viewState.userId === session?.user.id
      ? viewState.status
      : "loading";
  const viewBlocked = Boolean(
    renderSavedViews && viewId && currentState !== "ready",
  );
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "teams:stories:layout",
    "list",
    { initializeWithValue: false },
  );

  return (
    <TeamOptionsProvider layout={layout} viewId={viewId}>
      <Box className="flex h-full min-h-0 flex-col">
        <Header
          layout={layout}
          onViewLoadStateChange={setViewState}
          renderSavedViews={renderSavedViews}
          setLayout={setLayout}
        />
        {creating && renderViewCreator ? (
          <ActiveViewCreator
            layout={layout}
            renderViewCreator={renderViewCreator}
            teamId={teamId}
          />
        ) : null}
        {!viewBlocked ? (
          <ActiveStoriesFilterBar layout={layout} teamId={teamId} />
        ) : null}
        <Box className="min-h-0 flex-1">
          {viewBlocked && currentState === "loading" ? (
            <>
              <Text className="sr-only" role="status">
                Loading saved view...
              </Text>
              <Box aria-hidden className="h-full overflow-hidden" inert>
                <BoardSkeleton className="h-full" layout={layout} />
              </Box>
            </>
          ) : null}
          {viewBlocked && currentState !== "loading" ? (
            <Box className="space-y-3 p-6">
              <Text color="muted" role="alert">
                {savedViewMessage(currentState)}
              </Text>
              <Button
                color="tertiary"
                href={withWorkspace(`/teams/${teamId}/stories`)}
                variant="outline"
              >
                Back to team tasks
              </Button>
            </Box>
          ) : null}
          {!viewBlocked ? <AllStories layout={layout} /> : null}
        </Box>
      </Box>
    </TeamOptionsProvider>
  );
};
