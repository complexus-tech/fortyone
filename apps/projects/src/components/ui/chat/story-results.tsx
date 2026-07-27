"use client";

import { Tooltip } from "ui";
import { useTerminology, useWorkspacePath } from "@/hooks";
import { useStatuses } from "@/lib/hooks/statuses";
import { getStoryPath } from "@/modules/story/utils/story-url";
import { PriorityIcon } from "../priority-icon";
import { Dot } from "../dot";
import { GenerativeList, GenerativeListItem } from "./generative-list";
import { getStoryResults } from "./story-results-data";

export const StoryResults = ({ output }: { output: unknown }) => {
  const stories = getStoryResults(output);
  const { getTermDisplay } = useTerminology();
  const { withWorkspace } = useWorkspacePath();
  const { data: statuses = [] } = useStatuses();
  const statusesById = new Map(statuses.map((status) => [status.id, status]));
  const storyLabel = getTermDisplay("storyTerm", {
    capitalize: true,
    variant: "plural",
  });

  return (
    <GenerativeList
      emptyMessage={`No ${storyLabel.toLowerCase()} found.`}
      title={storyLabel}
    >
      {stories.map((story) => {
        const status = statusesById.get(story.statusId);

        return (
          <GenerativeListItem
            ariaLabel={`Open ${story.title}. Priority: ${story.priority}. Status: ${status?.name ?? "Unknown"}.`}
            href={withWorkspace(getStoryPath(story))}
            key={story.id}
            leading={
              <Tooltip title={`Priority: ${story.priority}`}>
                <span className="flex size-5 shrink-0 items-center justify-center">
                  <PriorityIcon
                    className="max-h-4 max-w-4"
                    priority={story.priority}
                  />
                </span>
              </Tooltip>
            }
            title={story.title}
            trailing={
              <Tooltip title={`Status: ${status?.name ?? "Unknown"}`}>
                <span className="flex size-5 shrink-0 items-center justify-center">
                  <Dot className="size-3" color={status?.color} />
                </span>
              </Tooltip>
            }
          />
        );
      })}
    </GenerativeList>
  );
};
