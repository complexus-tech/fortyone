"use client";

import { ListStories } from "@/modules/teams/public/story-view";
import type { SavedViewsAction } from "@/modules/teams/public/story-view";
import { SavedViews } from "@/modules/work-presets/public/views";

const renderSavedViews: SavedViewsAction = (props) => <SavedViews {...props} />;

export const TeamStoriesClient = () => (
  <ListStories renderSavedViews={renderSavedViews} />
);
