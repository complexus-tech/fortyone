export type StoryChecklistItem = {
  id: string;
  storyId: string;
  title: string;
  completed: boolean;
  position: number;
  convertedAt: string | null;
  convertedStory: {
    id: string;
    sequenceId: number;
    title: string;
    teamId: string;
    statusId: string | null;
  } | null;
  createdAt: string;
  updatedAt: string;
};

export type ChecklistCommand =
  | { type: "create"; title: string }
  | {
      type: "update";
      itemId: string;
      updates: { title?: string; completed?: boolean };
    }
  | { type: "delete"; itemId: string }
  | { type: "reorder"; itemIds: string[] }
  | { type: "clear" }
  | { type: "convert"; itemId: string };

export type ChecklistMutationResult =
  | StoryChecklistItem
  | StoryChecklistItem[]
  | null;
