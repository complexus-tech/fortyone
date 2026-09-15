import type { RichTextValue } from "../../components/rich-text/content";
import type { FormState } from "./form-state";
import { isRichTextValue } from "../../components/rich-text/content";

export type StorySubmissionState = {
  isSubmitting: boolean;
  isFinalized: boolean;
};

type SubmissionOptions = {
  getDraft: () => FormState;
  description: RichTextValue;
  availableTeamIds: string[];
  createIdempotencyKey: () => string;
  persistDraft: (snapshot: FormState) => Promise<void>;
  createStory: (snapshot: FormState) => Promise<string>;
  clearDraft: () => Promise<void>;
  assertActive: () => void;
};

export class CreatedStoryDraftCleanupError extends Error {
  constructor() {
    super(
      "Your task was created, but this device could not remove its draft. Try again to open the same task safely.",
    );
    this.name = "CreatedStoryDraftCleanupError";
  }
}

/** Owns one creation attempt, including durable cleanup after server success. */
export const createStoryFinalizer = (
  onStateChange: (state: StorySubmissionState) => void,
) => {
  let inFlight: Promise<string> | null = null;
  let isSubmitting = false;
  let createdStoryId: string | null = null;
  let cleared = false;

  const publish = () =>
    onStateChange({ isSubmitting, isFinalized: createdStoryId !== null });
  const canEdit = () => !isSubmitting && createdStoryId === null;

  const persistEdit = (write: () => Promise<void>): Promise<void> => {
    if (!canEdit()) {
      return Promise.reject(
        new Error(
          "This task is being submitted. Its draft can no longer be changed.",
        ),
      );
    }
    try {
      return write();
    } catch (cause) {
      return Promise.reject(cause);
    }
  };

  const submit = (options: SubmissionOptions): Promise<string> => {
    if (inFlight) return inFlight;
    // Close the write gate synchronously, before the first asynchronous operation.
    isSubmitting = true;
    publish();

    inFlight = Promise.resolve()
      .then(async () => {
        options.assertActive();
        if (!createdStoryId) {
          // The DOM flush has acknowledged this description. Read the native
          // fields now, rather than using a render snapshot from before the flush.
          const latest = options.getDraft();
          const teamId = latest.teamId || options.availableTeamIds[0] || "";
          if (!latest.title.trim())
            throw new Error("Add a title before creating a task.");
          if (!teamId || !options.availableTeamIds.includes(teamId)) {
            throw new Error(
              "Choose an available team before creating this task.",
            );
          }
          if (!isRichTextValue(options.description)) {
            throw new Error(
              "The editor could not provide your latest description. Please try again.",
            );
          }
          const snapshot: FormState = {
            ...latest,
            title: latest.title.trim(),
            teamId,
            description: {
              ...options.description,
              mentions: [...options.description.mentions],
            },
            labelIds: [...latest.labelIds],
            activeSheet: null,
            idempotencyKey:
              latest.idempotencyKey || options.createIdempotencyKey(),
          };
          // Persist the complete payload and its retry identity before issuing a POST.
          await options.persistDraft(snapshot);
          options.assertActive();
          const storyId = await options.createStory(snapshot);
          if (!storyId)
            throw new Error("The server did not return the created task.");
          createdStoryId = storyId;
          publish();
        }

        options.assertActive();
        if (!cleared) {
          try {
            await options.clearDraft();
          } catch {
            // Keep the server result. A retry must only repeat local cleanup.
            throw new CreatedStoryDraftCleanupError();
          }
          cleared = true;
        }
        options.assertActive();
        return createdStoryId;
      })
      .finally(() => {
        isSubmitting = false;
        inFlight = null;
        publish();
      });
    return inFlight;
  };

  return { submit, canEdit, persistEdit };
};
