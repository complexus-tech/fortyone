const UNCERTAIN_STORY_DELETION_MESSAGE =
  "The story deletion request did not return a definitive result.";

/**
 * Signals that a story-deletion request crossed the mutation boundary but its
 * response could not prove whether the server committed the deletion. Approved
 * tool execution must let this reach the mutation ledger for reconciliation.
 */
export class StoryDeletionOutcomeUncertainError extends Error {
  constructor(cause: unknown) {
    super(UNCERTAIN_STORY_DELETION_MESSAGE, { cause });
    this.name = "StoryDeletionOutcomeUncertainError";
  }
}

export const isStoryDeletionOutcomeUncertainError = (
  error: unknown,
): error is StoryDeletionOutcomeUncertainError =>
  error instanceof StoryDeletionOutcomeUncertainError;
