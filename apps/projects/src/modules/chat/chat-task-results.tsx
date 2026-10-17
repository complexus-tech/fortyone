"use client";

import { StoryResults } from "@/components/ui/chat/story-results";

type ChatTaskResult = {
  id: string;
  title: string;
  reference?: string;
  priority?: string;
  statusId?: string;
};

// eslint-disable-next-line prefer-regex-literals -- Named literals fail the ES2017 TypeScript target.
const TASK_REFERENCE_PATTERN = new RegExp(
  "^(?<teamCode>.+)-(?<sequenceId>\\d+)$",
);

/** Use the assistant's shared renderer so task previews keep one visual contract. */
export const ChatTaskResults = ({ tasks }: { tasks: ChatTaskResult[] }) => (
  <StoryResults
    output={{
      success: true,
      stories: tasks.map((task) => {
        const reference = TASK_REFERENCE_PATTERN.exec(
          task.reference ?? "",
        )?.groups;
        return {
          ...task,
          statusId: task.statusId ?? "",
          priority: task.priority ?? "No Priority",
          teamCode: reference?.teamCode,
          sequenceId: reference ? Number(reference.sequenceId) : undefined,
        };
      }),
    }}
  />
);
