"use client";

import Link from "next/link";
import { useMemo } from "react";
import { Tooltip } from "ui";
import { useTerminology, useWorkspacePath } from "@/hooks";
import { useStatuses } from "@/lib/hooks/statuses";
import { PriorityIcon } from "../priority-icon";
import { Dot } from "../dot";
import { getStoryResults } from "./story-results-data";

export const StoryResults = ({ output }: { output: unknown }) => {
  const stories = getStoryResults(output);
  const { getTermDisplay } = useTerminology();
  const { withWorkspace } = useWorkspacePath();
  const { data: statuses = [] } = useStatuses();
  const statusesById = useMemo(
    () => new Map(statuses.map((status) => [status.id, status])),
    [statuses],
  );
  const storyLabel = getTermDisplay("storyTerm", {
    capitalize: true,
    variant: "plural",
  });

  return (
    <section aria-label={storyLabel} className="mt-3 mb-1 grid gap-2.5">
      <span className="text-text-muted text-[0.82rem] font-medium">
        {storyLabel}
      </span>
      <div className="border-border/70 grid border-t dark:border-white/[0.09]">
        {stories.length ? (
          stories.map((story) => {
            const status = statusesById.get(story.statusId);

            return (
              <Link
                aria-label={`Open ${story.title}. Priority: ${story.priority}. Status: ${status?.name ?? "Unknown"}.`}
                className="border-border/70 text-foreground focus-visible:ring-foreground/50 flex min-h-[50px] items-center gap-3 border-b px-px py-3 text-[0.95rem] leading-5 no-underline transition-[opacity,transform] duration-150 outline-none hover:opacity-70 focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.99] dark:border-white/[0.09]"
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
          <span className="border-border/70 text-text-muted border-b px-px py-3 text-sm dark:border-white/[0.09]">
            No {storyLabel.toLowerCase()} found.
          </span>
        )}
      </div>
    </section>
  );
};
