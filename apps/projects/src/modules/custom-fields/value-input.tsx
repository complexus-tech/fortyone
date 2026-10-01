import { Box, Input, Select, Text } from "ui";
import type { AriaAttributes } from "react";
import { useTeamMembers } from "@/lib/hooks/team-members";
import type { CustomField } from "./types";
import { getCustomFieldInputLabel } from "./value-utils";

const EMPTY_VALUE = "__not_set__";

const PersonInput = ({
  field,
  value,
  onChange,
  disabled,
  ...accessibility
}: CustomFieldValueInputProps) => {
  const {
    data: members = [],
    isPending,
    isError,
  } = useTeamMembers(field.teamId);
  const activeMembers = members.filter(
    (member) => member.isActive && !member.isSystem,
  );
  return (
    <Box>
      <Select
        disabled={disabled || isPending || isError}
        onValueChange={(next) => {
          onChange(next === EMPTY_VALUE ? null : next);
        }}
        value={value || EMPTY_VALUE}
      >
        <Select.Trigger
          aria-label={field.name}
          {...accessibility}
          className="h-11 w-full text-base"
        >
          <Select.Input placeholder="Choose a team member" />
        </Select.Trigger>
        <Select.Content>
          <Select.Option className="text-base" value={EMPTY_VALUE}>
            Not set
          </Select.Option>
          {value && !activeMembers.some((member) => member.id === value) ? (
            <Select.Option className="text-base" disabled value={value}>
              Unavailable member
            </Select.Option>
          ) : null}
          {activeMembers.map((member) => (
            <Select.Option
              className="text-base"
              key={member.id}
              value={member.id}
            >
              {member.fullName || member.username}
            </Select.Option>
          ))}
        </Select.Content>
      </Select>
      {isError ? (
        <Text className="mt-2" color="danger" role="alert">
          Team members could not be loaded.
        </Text>
      ) : null}
    </Box>
  );
};

type CustomFieldValueInputProps = Pick<
  AriaAttributes,
  "aria-describedby" | "aria-invalid"
> & {
  field: CustomField;
  value: string | null;
  onChange: (value: string | null) => void;
  disabled?: boolean;
};

export const CustomFieldValueInput = ({
  field,
  value,
  onChange,
  disabled,
  ...accessibility
}: CustomFieldValueInputProps) => {
  if (field.type === "person")
    return (
      <PersonInput
        {...accessibility}
        disabled={disabled}
        field={field}
        onChange={onChange}
        value={value}
      />
    );
  if (field.type === "select") {
    return (
      <Select
        disabled={disabled}
        onValueChange={(next) => {
          onChange(next === EMPTY_VALUE ? null : next);
        }}
        value={value || EMPTY_VALUE}
      >
        <Select.Trigger
          aria-label={field.name}
          {...accessibility}
          className="h-11 w-full text-base"
        >
          <Select.Input placeholder="Choose an option" />
        </Select.Trigger>
        <Select.Content>
          <Select.Option className="text-base" value={EMPTY_VALUE}>
            Not set
          </Select.Option>
          {field.options
            .filter((option) => !option.archivedAt || option.id === value)
            .map((option) => (
              <Select.Option
                className="text-base"
                disabled={Boolean(option.archivedAt)}
                key={option.id}
                value={option.id}
              >
                {option.name}
                {option.archivedAt ? " (archived)" : ""}
              </Select.Option>
            ))}
        </Select.Content>
      </Select>
    );
  }
  return (
    <Input
      {...accessibility}
      aria-label={getCustomFieldInputLabel(field)}
      className="h-11 text-base"
      disabled={disabled}
      inputMode={
        field.type === "money" || field.type === "number"
          ? "decimal"
          : undefined
      }
      maxLength={field.type === "text" ? 10_000 : 200}
      onChange={(event) => {
        onChange(event.target.value || null);
      }}
      placeholder={
        field.type === "money" ? `${field.currency ?? ""} 0.00` : "Not set"
      }
      type={field.type === "date" ? "date" : "text"}
      value={value ?? ""}
    />
  );
};
