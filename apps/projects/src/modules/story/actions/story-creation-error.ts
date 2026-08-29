const UNCERTAIN_STORY_CREATION_MESSAGE =
  "The story creation request did not return a definitive result.";

/**
 * Signals that a story-creation request crossed the mutation boundary but its
 * response could not prove whether the server committed the story. Approved
 * tool execution must let this reach the mutation ledger so it can quarantine
 * the operation until the idempotency key is reconciled.
 */
export class StoryCreationOutcomeUncertainError extends Error {
  constructor(cause: unknown) {
    super(UNCERTAIN_STORY_CREATION_MESSAGE, { cause });
    this.name = "StoryCreationOutcomeUncertainError";
  }
}

export const isStoryCreationOutcomeUncertainError = (
  error: unknown,
): error is StoryCreationOutcomeUncertainError =>
  error instanceof StoryCreationOutcomeUncertainError;
