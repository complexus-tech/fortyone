"use client";

import Link from "next/link";
import { useState } from "react";
import { Tooltip } from "ui";
import { useTerminology, useWorkspacePath } from "@/hooks";
import { useStatuses } from "@/lib/hooks/statuses";
import { PriorityIcon } from "../priority-icon";
import { Dot } from "../dot";
import {
  getStoryResults,
  getVisibleStoryResults,
  STORY_RESULTS_PREVIEW_LIMIT,
} from "./story-results-data";

export const StoryResults = ({ output }: { output: unknown }) => {
  const stories = getStoryResults(output);
  const [showAll, setShowAll] = useState(false);
  const { getTermDisplay } = useTerminology();
  const { withWorkspace } = useWorkspacePath();
  const { data: statuses = [] } = useStatuses();
  const statusesById = new Map(statuses.map((status) => [status.id, status]));
  const storyLabel = getTermDisplay("storyTerm", {
    capitalize: true,
    variant: "plural",
  });
  const visibleStories = getVisibleStoryResults(stories, showAll);
  const hasAdditionalStories = stories.length > STORY_RESULTS_PREVIEW_LIMIT;
  const remainingStoryCount = stories.length - STORY_RESULTS_PREVIEW_LIMIT;

  return (
    <section
      aria-label={storyLabel}
      className="mt-3 grid w-full max-w-full min-w-0 gap-2.5 overflow-hidden"
    >
      <span className="text-text-muted text-base font-medium">
        {storyLabel}
      </span>
      <div className="border-border/70 grid w-full max-w-full min-w-0 overflow-hidden border-t dark:border-white/[0.09]">
        {stories.length ? (
          visibleStories.map((story) => {
            const status = statusesById.get(story.statusId);

            return (
              <Link
                aria-label={`Open ${story.title}. Priority: ${story.priority}. Status: ${status?.name ?? "Unknown"}.`}
                className="border-border/70 text-foreground focus-visible:ring-foreground/50 flex min-h-[52px] w-full max-w-full min-w-0 items-center gap-3 overflow-hidden border-b px-px py-3 text-base leading-6 no-underline transition-[opacity,transform] duration-150 outline-none hover:opacity-70 focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.99] dark:border-white/[0.09]"
                href={withWorkspace(`/story/${story.id}`)}
                key={story.id}
              >
                <Tooltip title={`Priority: ${story.priority}`}>
                  <span className="flex size-5 shrink-0 items-center justify-center">
                    <PriorityIcon
                      className="max-h-4 max-w-4"
                      priority={story.priority}
                    />
                  </span>
                </Tooltip>
                <span className="min-w-0 flex-1 truncate">{story.title}</span>
                <Tooltip title={`Status: ${status?.name ?? "Unknown"}`}>
                  <span className="flex size-5 shrink-0 items-center justify-center">
                    <Dot className="size-3" color={status?.color} />
                  </span>
                </Tooltip>
              </Link>
            );
          })
        ) : (
          <span className="border-border/70 text-text-muted border-b px-px py-3 text-base dark:border-white/[0.09]">
            No {storyLabel.toLowerCase()} found.
          </span>
        )}
        {hasAdditionalStories ? (
          <button
            className="border-border/70 text-foreground focus-visible:ring-foreground/50 min-h-12 w-full border-b px-px py-3 text-left text-base font-medium transition-opacity outline-none hover:opacity-70 focus-visible:ring-2 focus-visible:ring-offset-2 dark:border-white/[0.09]"
            onClick={() => {
              setShowAll((current) => !current);
            }}
            type="button"
          >
            {showAll
              ? "Show less"
              : `View ${remainingStoryCount} more ${remainingStoryCount === 1 ? getTermDisplay("storyTerm") : getTermDisplay("storyTerm", { variant: "plural" })}`}
          </button>
        ) : null}
      </div>
    </section>
  );
};
