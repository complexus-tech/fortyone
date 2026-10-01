"use client";

import { useId, useState } from "react";
import { Box, Button, Flex, Popover, Text } from "ui";
import { useTeamMembers } from "@/lib/hooks/team-members";
import type { CustomField } from "./types";
import { CustomFieldIcon } from "./icons";
import { CustomFieldValueInput } from "./value-input";
import {
  formatCustomFieldValue,
  getCustomFieldInputLabel,
  validateCustomFieldValue,
} from "./value-utils";

export const CreateFieldChip = ({
  field,
  value,
  onChange,
  disabled,
  initiallyOpen = false,
}: {
  field: CustomField;
  value: string | null;
  onChange: (value: string | null) => void;
  disabled?: boolean;
  initiallyOpen?: boolean;
}) => {
  const [open, setOpen] = useState(initiallyOpen);
  const [draft, setDraft] = useState(value);
  const [validation, setValidation] = useState<string | null>(null);
  const labelId = useId();
  const errorId = useId();
  const { data: people = [] } = useTeamMembers(
    field.type === "person" ? field.teamId : undefined,
  );
  const label = getCustomFieldInputLabel(field);
  const chipText = formatCustomFieldValue(field, value, people);
  const chipLabel = `${field.name}: ${chipText}`;
  const apply = () => {
    const error = validateCustomFieldValue(field, draft);
    setValidation(error);
    if (error) return;
    onChange(draft?.trim() || null);
    setOpen(false);
  };
  return (
    <Box className="order-12 max-w-full">
      <Popover
        onOpenChange={(next) => {
          setOpen(next);
          if (next) {
            setDraft(value);
            setValidation(null);
          }
        }}
        open={open}
      >
        <Popover.Trigger asChild>
          <Button
            aria-label={chipLabel}
            className="dark:bg-surface-elevated/90 gap-1.5 px-2"
            color="tertiary"
            disabled={disabled}
            leftIcon={
              <CustomFieldIcon
                className="h-4.5 w-auto shrink-0"
                field={field}
              />
            }
            size="sm"
            title={chipLabel}
            type="button"
            variant="outline"
          >
            <span className="inline-block max-w-[24ch] truncate">
              {chipText}
            </span>
          </Button>
        </Popover.Trigger>
        <Popover.Content
          align="start"
          aria-labelledby={labelId}
          className="w-80 max-w-[calc(100vw-2rem)] p-3"
        >
          <Text className="mb-3" fontWeight="medium" id={labelId}>
            {label}
          </Text>
          <CustomFieldValueInput
            aria-describedby={validation ? errorId : undefined}
            aria-invalid={Boolean(validation)}
            disabled={disabled}
            field={field}
            onChange={(next) => {
              setDraft(next);
              setValidation(null);
            }}
            value={draft}
          />
          {validation ? (
            <Text className="mt-3" color="danger" id={errorId} role="alert">
              {validation}
            </Text>
          ) : null}
          <Flex className="mt-3" gap={2} justify="between">
            <Button
              color="tertiary"
              disabled={disabled || !draft}
              onClick={() => {
                onChange(null);
                setOpen(false);
              }}
              type="button"
              variant="naked"
            >
              Clear
            </Button>
            <Button disabled={disabled} onClick={apply} type="button">
              Apply
            </Button>
          </Flex>
        </Popover.Content>
      </Popover>
    </Box>
  );
};
