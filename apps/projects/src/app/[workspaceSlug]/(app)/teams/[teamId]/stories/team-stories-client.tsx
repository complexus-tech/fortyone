"use client";

import { ListStories } from "@/modules/teams/public/story-view";
import type {
  SavedViewsAction,
  ViewCreatorAction,
} from "@/modules/teams/public/story-view";
import { SavedViews, ViewCreator } from "@/modules/work-presets/public/views";

const renderSavedViews: SavedViewsAction = (props) => <SavedViews {...props} />;
const renderViewCreator: ViewCreatorAction = (props) => (
  <ViewCreator {...props} />
);

export const TeamStoriesClient = () => (
  <ListStories
    renderSavedViews={renderSavedViews}
    renderViewCreator={renderViewCreator}
  />
);
