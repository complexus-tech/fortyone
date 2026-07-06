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
  RuntimeOption,
  SlackActor,
} from "@/lib/runtime";

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
  initialTeam?: RuntimeOption;
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
  initialTeam,
  source,
  title,
}: {
  description?: string;
  initialTeam?: RuntimeOption;
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
        initialOption={initialTeam}
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
        minQueryLength={0}
        optional
        placeholder="Search people"
      />
      <ExternalSelect
        id={CREATE_STORY_FIELDS.objective}
        label="Objective"
        minQueryLength={0}
        optional
        placeholder="Search objectives"
      />
      <ExternalSelect
        id={CREATE_STORY_FIELDS.label}
        label="Label"
        minQueryLength={0}
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

const slackPlainText = (text: string) => ({
  text,
  type: "plain_text",
});

const slackTextInput = ({
  id,
  initialValue,
  label,
  maxLength,
  multiline,
  optional,
  placeholder,
}: {
  id: string;
  initialValue?: string;
  label: string;
  maxLength?: number;
  multiline?: boolean;
  optional?: boolean;
  placeholder?: string;
}) => ({
  block_id: id,
  element: {
    action_id: id,
    ...(initialValue ? { initial_value: initialValue } : {}),
    ...(maxLength ? { max_length: maxLength } : {}),
    ...(multiline ? { multiline } : {}),
    ...(placeholder ? { placeholder: slackPlainText(placeholder) } : {}),
    type: "plain_text_input",
  },
  label: slackPlainText(label),
  optional: optional ?? false,
  type: "input",
});

const slackExternalSelect = ({
  id,
  initialOption,
  label,
  optional,
  placeholder,
  type = "external_select",
}: {
  id: string;
  initialOption?: RuntimeOption;
  label: string;
  optional?: boolean;
  placeholder: string;
  type?: "external_select" | "multi_external_select";
}) => ({
  block_id: id,
  element: {
    action_id: id,
    ...(initialOption
      ? {
          initial_option: {
            text: slackPlainText(initialOption.label),
            value: initialOption.value,
          },
        }
      : {}),
    min_query_length: 0,
    placeholder: slackPlainText(placeholder),
    type,
  },
  label: slackPlainText(label),
  optional: optional ?? false,
  type: "input",
});

export const CreateStorySlackView = ({
  description,
  initialTeam,
  source,
  title,
}: {
  description?: string;
  initialTeam?: RuntimeOption;
  source: ModalMetadata["source"];
  title?: string;
}) => ({
  callback_id: CREATE_STORY_MODAL_ID,
  close: slackPlainText("Cancel"),
  private_metadata: JSON.stringify({
    m: encodeCreateStoryMetadata({ initialTeam, source }),
  }),
  submit: slackPlainText("Create"),
  title: slackPlainText("Create story"),
  type: "modal",
  blocks: [
    slackTextInput({
      id: CREATE_STORY_FIELDS.title,
      initialValue: title,
      label: "Title",
      maxLength: 120,
      placeholder: "What needs to be done?",
    }),
    slackTextInput({
      id: CREATE_STORY_FIELDS.description,
      initialValue: description,
      label: "Description",
      multiline: true,
      optional: true,
      placeholder: "Add context from Slack or write details here",
    }),
    slackExternalSelect({
      id: CREATE_STORY_FIELDS.team,
      initialOption: initialTeam,
      label: "Team",
      placeholder: "Search teams",
    }),
    slackExternalSelect({
      id: CREATE_STORY_FIELDS.status,
      label: "Status",
      optional: true,
      placeholder: "Search statuses",
    }),
    slackExternalSelect({
      id: CREATE_STORY_FIELDS.assignee,
      label: "Assignee",
      optional: true,
      placeholder: "Search people",
    }),
    slackExternalSelect({
      id: CREATE_STORY_FIELDS.objective,
      label: "Objective",
      optional: true,
      placeholder: "Search objectives",
    }),
    slackExternalSelect({
      id: CREATE_STORY_FIELDS.label,
      label: "Labels",
      optional: true,
      placeholder: "Search labels",
      type: "multi_external_select",
    }),
    {
      block_id: CREATE_STORY_FIELDS.priority,
      element: {
        action_id: CREATE_STORY_FIELDS.priority,
        initial_option: {
          text: slackPlainText("No Priority"),
          value: "No Priority",
        },
        options: PRIORITIES.map((priority) => ({
          text: slackPlainText(priority),
          value: priority,
        })),
        placeholder: slackPlainText("Select priority"),
        type: "static_select",
      },
      label: slackPlainText("Priority"),
      optional: true,
      type: "input",
    },
  ],
});

const selectedOptionValueFromRaw = (
  raw: unknown,
  actionId: string,
): string | undefined => {
  const stateValues = (raw as { view?: { state?: { values?: unknown } } }).view
    ?.state?.values;
  if (!stateValues || typeof stateValues !== "object") {
    return undefined;
  }

  for (const block of Object.values(stateValues)) {
    if (!block || typeof block !== "object") {
      continue;
    }

    const input = (block as Record<string, unknown>)[actionId] as
      | { selected_option?: { value?: unknown }; value?: unknown }
      | undefined;
    const selectedValue = input?.selected_option?.value;
    if (typeof selectedValue === "string" && selectedValue.trim()) {
      return selectedValue;
    }

    if (typeof input?.value === "string" && input.value.trim()) {
      return input.value;
    }
  }

  return undefined;
};

const selectedOptionValuesFromRaw = (
  raw: unknown,
  actionId: string,
): string[] => {
  const stateValues = (raw as { view?: { state?: { values?: unknown } } }).view
    ?.state?.values;
  if (!stateValues || typeof stateValues !== "object") {
    return [];
  }

  for (const block of Object.values(stateValues)) {
    if (!block || typeof block !== "object") {
      continue;
    }

    const input = (block as Record<string, unknown>)[actionId] as
      | { selected_options?: unknown }
      | undefined;
    const selectedOptions = input?.selected_options;
    if (Array.isArray(selectedOptions)) {
      return selectedOptions
        .map((option) => (option as { value?: unknown }).value)
        .filter((value): value is string => typeof value === "string");
    }
  }

  return [];
};

export const buildCreateStoryInput = (
  values: Record<string, string>,
  metadata: ModalMetadata,
  raw?: unknown,
): CreateStoryFromSlackInput => {
  const selectedLabelIds = selectedOptionValuesFromRaw(
    raw,
    CREATE_STORY_FIELDS.label,
  );
  const fallbackLabelId = values[CREATE_STORY_FIELDS.label];
  let labelIds: string[] | undefined;
  if (selectedLabelIds.length > 0) {
    labelIds = selectedLabelIds;
  } else if (fallbackLabelId) {
    labelIds = [fallbackLabelId];
  }

  const valueFor = (field: string) =>
    values[field] || selectedOptionValueFromRaw(raw, field);

  return {
    assigneeId: valueFor(CREATE_STORY_FIELDS.assignee) || undefined,
    description:
      valueFor(CREATE_STORY_FIELDS.description) || metadata.source.messageText,
    labelIds,
    objectiveId: valueFor(CREATE_STORY_FIELDS.objective) || undefined,
    priority: valueFor(CREATE_STORY_FIELDS.priority) || "No Priority",
    source: metadata.source,
    statusId: valueFor(CREATE_STORY_FIELDS.status) || undefined,
    teamId:
      valueFor(CREATE_STORY_FIELDS.team) ?? metadata.initialTeam?.value ?? "",
    title: valueFor(CREATE_STORY_FIELDS.title) ?? "",
  };
};

export const toSelectOptions = (
  options: RuntimeOption[],
): SelectOptionElement[] =>
  options.map((option) =>
    SelectOption({ label: option.label, value: option.value }),
  );
