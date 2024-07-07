"use client";
import type { StoriesLayout } from "@/components/ui";
import { StoriesBoard } from "@/components/ui";
import type { Story } from "@/modules/stories/types";
import { useSprintStories } from "./provider";

export const AllStories = ({
  layout,
  stories,
}: {
  layout: StoriesLayout;
  stories: Story[];
}) => {
  const { viewOptions } = useSprintStories();
  return (
    <StoriesBoard layout={layout} stories={stories} viewOptions={viewOptions} />
  );
};
