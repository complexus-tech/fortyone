import { Box, Select } from "ui";
import {
  TEAM_SIZES,
  WORK_TYPES,
  type WorkspaceOnboardingDraft,
} from "../workspace-onboarding-model";

type ChoiceProps = {
  draft: WorkspaceOnboardingDraft;
  onChange: (updates: Partial<WorkspaceOnboardingDraft>) => void;
};

const choiceClassName =
  "border-border flex cursor-pointer items-start rounded-lg border px-4 py-3 has-checked:border-foreground has-checked:bg-surface-muted/50 has-checked:ring-1 has-checked:ring-foreground has-checked:ring-inset has-[:focus-visible]:outline-2 has-[:focus-visible]:outline-offset-2 has-[:focus-visible]:outline-foreground";

export const WorkspaceWorkStep = ({ draft, onChange }: ChoiceProps) => (
  <>
    <fieldset className="space-y-2">
      <legend className="sr-only">What kind of work will you manage?</legend>
      {WORK_TYPES.map((workType) => (
        <label
          className={choiceClassName}
          htmlFor={`work-type-${workType.value}`}
          key={workType.value}
        >
          <input
            checked={draft.workType === workType.value}
            className="sr-only"
            id={`work-type-${workType.value}`}
            name="workType"
            onChange={() => {
              onChange({ workType: workType.value });
            }}
            type="radio"
            value={workType.value}
          />
          <span className="font-medium">{workType.label}</span>
        </label>
      ))}
    </fieldset>
    <Box>
      <label className="mb-2 block font-medium" htmlFor="team-size">
        How many people will use this workspace?
      </label>
      <Select
        onValueChange={(value) => {
          const size = TEAM_SIZES.find((teamSize) => teamSize === value);
          if (size) onChange({ teamSize: size });
        }}
        value={draft.teamSize}
      >
        <Select.Trigger
          className="border-border bg-surface/70 h-[2.7rem] w-full rounded-lg"
          id="team-size"
        >
          <Select.Input />
        </Select.Trigger>
        <Select.Content>
          <Select.Group>
            {TEAM_SIZES.map((size) => (
              <Select.Option
                className="h-10 rounded-lg text-[0.9rem]"
                key={size}
                value={size}
              >
                {size}
              </Select.Option>
            ))}
          </Select.Group>
        </Select.Content>
      </Select>
    </Box>
  </>
);

export const WorkspaceStartStep = ({ draft, onChange }: ChoiceProps) => {
  const choices = [
    {
      value: "task",
      label: "Create my first task",
      description: "Start with one real piece of work.",
    },
    {
      value: "import",
      label: "Import existing work",
      description: "Upload an export and review it before importing.",
    },
    {
      value: "examples",
      label: "Explore with examples",
      description: "Add three example tasks for your kind of work.",
    },
  ] as const;

  return (
    <fieldset className="space-y-2">
      <legend className="sr-only">How would you like to get started?</legend>
      {choices.map((choice) => (
        <label
          className={choiceClassName}
          htmlFor={`start-${choice.value}`}
          key={choice.value}
        >
          <input
            checked={draft.start === choice.value}
            className="sr-only"
            id={`start-${choice.value}`}
            name="start"
            onChange={() => {
              onChange({ start: choice.value });
            }}
            type="radio"
            value={choice.value}
          />
          <span className="font-medium">
            {choice.label}
            <span className="text-text-muted mt-0.5 block font-normal">
              {choice.description}
            </span>
          </span>
        </label>
      ))}
    </fieldset>
  );
};
