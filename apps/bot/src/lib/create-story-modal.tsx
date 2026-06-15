/* @jsxImportSource chat */

import {
  ExternalSelect,
  Modal,
  Select,
  SelectOption,
  TextInput,
  type SelectOptionElement,
} from "chat";
import type {
  CreateStoryFromSlackInput,
  SlackActor,
  RuntimeOption,
} from "@/lib/fortyone-client";

export const CREATE_STORY_MODAL_ID = "fortyone_create_story";

export const CREATE_STORY_ACTION_ID = "fortyone_create_story_from_message";

export const CREATE_STORY_FIELDS = {
  assignee: "assignee",
  description: "description",
  label: "label",
  objective: "objective",
  priority: "priority",
  status: "status",
  team: "team",
  title: "title",
} as const;

const PRIORITIES = ["No Priority", "Low", "Medium", "High", "Urgent"];

interface ModalMetadata {
  source: SlackActor & {
    messageText?: string;
  };
}

export const encodeCreateStoryMetadata = (metadata: ModalMetadata) =>
  JSON.stringify(metadata);

export const decodeCreateStoryMetadata = (
  value: string | undefined,
): ModalMetadata => {
  if (!value) {
    return { source: { userId: "" } };
  }
  return JSON.parse(value) as ModalMetadata;
};

export function CreateStoryModal({
  description,
  source,
  title,
}: {
  description?: string;
  source: ModalMetadata["source"];
  title?: string;
}) {
  return (
    <Modal
      callbackId={CREATE_STORY_MODAL_ID}
      closeLabel="Cancel"
      privateMetadata={encodeCreateStoryMetadata({ source })}
      submitLabel="Create"
      title="Create story"
    >
      <TextInput
        id={CREATE_STORY_FIELDS.title}
        initialValue={title}
        label="Title"
        maxLength={120}
        placeholder="What needs to be done?"
      />
      <TextInput
        id={CREATE_STORY_FIELDS.description}
        initialValue={description}
        label="Description"
        multiline
        optional
        placeholder="Add context from Slack or write details here"
      />
      <ExternalSelect
        id={CREATE_STORY_FIELDS.team}
        label="Team"
        minQueryLength={0}
        placeholder="Search teams"
      />
      <ExternalSelect
        id={CREATE_STORY_FIELDS.status}
        label="Status"
        minQueryLength={0}
        optional
        placeholder="Search statuses"
      />
      <ExternalSelect
        id={CREATE_STORY_FIELDS.assignee}
        label="Assignee"
        minQueryLength={1}
        optional
        placeholder="Search people"
      />
      <ExternalSelect
        id={CREATE_STORY_FIELDS.objective}
        label="Objective"
        minQueryLength={1}
        optional
        placeholder="Search objectives"
      />
      <ExternalSelect
        id={CREATE_STORY_FIELDS.label}
        label="Label"
        minQueryLength={1}
        optional
        placeholder="Search labels"
      />
      <Select
        id={CREATE_STORY_FIELDS.priority}
        initialOption="No Priority"
        label="Priority"
        optional
        placeholder="Select priority"
      >
        {PRIORITIES.map((priority) => (
          <SelectOption key={priority} label={priority} value={priority} />
        ))}
      </Select>
    </Modal>
  );
}

export const buildCreateStoryInput = (
  values: Record<string, string>,
  metadata: ModalMetadata,
): CreateStoryFromSlackInput => ({
  assigneeId: values[CREATE_STORY_FIELDS.assignee] || undefined,
  description:
    values[CREATE_STORY_FIELDS.description] || metadata.source.messageText,
  labelIds: values[CREATE_STORY_FIELDS.label]
    ? [values[CREATE_STORY_FIELDS.label]]
    : undefined,
  objectiveId: values[CREATE_STORY_FIELDS.objective] || undefined,
  priority: values[CREATE_STORY_FIELDS.priority] || "No Priority",
  source: metadata.source,
  statusId: values[CREATE_STORY_FIELDS.status] || undefined,
  teamId: values[CREATE_STORY_FIELDS.team] ?? "",
  title: values[CREATE_STORY_FIELDS.title] ?? "",
});

export const toSelectOptions = (
  options: RuntimeOption[],
): SelectOptionElement[] =>
  options.map((option) =>
    SelectOption({ label: option.label, value: option.value }),
  );
