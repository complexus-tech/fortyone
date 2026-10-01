"use client";

import type { Ref } from "react";
import { useId, useImperativeHandle, useRef, useState } from "react";
import { ApiError } from "api-client";
import { Button, Flex, Popover, Text } from "ui";
import type { CustomField } from "./types";
import { useUpdateStoryCustomFields } from "./hooks";
import { CustomFieldIcon } from "./icons";
import { CustomFieldValueInput } from "./value-input";
import {
  formatCustomFieldValue,
  getCustomFieldInputLabel,
  validateCustomFieldValue,
} from "./value-utils";

type FieldValueSnapshot = { value: string | null; version?: number };
type FieldValueEditorHandle = { isBusy: boolean };
type FieldValuePickerProps = FieldValueSnapshot & {
  field: CustomField;
  storyId: string;
  disabled?: boolean;
  people?: { id: string; fullName: string }[];
  variant?: "solid" | "naked" | "outline";
  size?: "sm" | "xs";
  align?: "start" | "end";
  onOpenChange?: (open: boolean) => void;
  onReload: () => Promise<FieldValueSnapshot>;
};

/** Mounted only while editing; the opening snapshot stays fixed during refetches. */
const FieldValueEditor = ({
  field,
  storyId,
  value,
  version,
  people,
  disabled,
  align,
  onReload,
  onClose,
  editorRef,
}: FieldValuePickerProps & {
  onClose: () => void;
  editorRef: Ref<FieldValueEditorHandle>;
}) => {
  const [draft, setDraft] = useState(value);
  const [baseline, setBaseline] = useState<FieldValueSnapshot>({
    value,
    version,
  });
  const [validation, setValidation] = useState<string | null>(null);
  const [hasConflict, setHasConflict] = useState(false);
  const [isReloading, setIsReloading] = useState(false);
  const [reviewedValue, setReviewedValue] = useState<string | null>(null);
  const mutation = useUpdateStoryCustomFields(storyId);
  const labelId = useId();
  const errorId = useId();
  const isBusy = mutation.isPending || isReloading;
  const error = validation ?? mutation.error?.message;
  useImperativeHandle(editorRef, () => ({ isBusy }), [isBusy]);

  const save = async () => {
    if (isBusy || disabled || hasConflict) return;
    const validationError = validateCustomFieldValue(field, draft);
    setValidation(validationError);
    if (validationError) return;
    try {
      await mutation.mutateAsync({
        values: [{ fieldId: field.id, value: draft?.trim() || null }],
        expectedVersion: baseline.version,
      });
      onClose();
    } catch (cause) {
      if (cause instanceof ApiError && cause.status === 409) {
        setHasConflict(true);
      }
      // React Query retains the API error; the user's draft remains editable.
    }
  };

  const reviewLatest = async () => {
    if (isBusy || disabled) return;
    setIsReloading(true);
    try {
      const latest = await onReload();
      setBaseline(latest);
      setReviewedValue(formatCustomFieldValue(field, latest.value, people));
      setHasConflict(false);
      setValidation(null);
      mutation.reset();
    } catch (cause) {
      setValidation(
        cause instanceof Error
          ? cause.message
          : "The value could not be reloaded.",
      );
    } finally {
      setIsReloading(false);
    }
  };

  return (
    <Popover.Content
      align={align}
      aria-busy={isBusy}
      aria-labelledby={labelId}
      className="w-80 max-w-[calc(100vw-2rem)] p-3"
      onClick={(event) => {
        event.stopPropagation();
      }}
      onEscapeKeyDown={(event) => {
        if (isBusy) event.preventDefault();
      }}
      onInteractOutside={(event) => {
        if (isBusy) event.preventDefault();
      }}
      onKeyDown={(event) => {
        event.stopPropagation();
      }}
      onPointerDown={(event) => {
        event.stopPropagation();
      }}
    >
      <Text className="mb-3" fontWeight="medium" id={labelId}>
        {getCustomFieldInputLabel(field)}
      </Text>
      <CustomFieldValueInput
        aria-describedby={error ? errorId : undefined}
        aria-invalid={Boolean(error)}
        disabled={isBusy || disabled}
        field={field}
        onChange={(next) => {
          setDraft(next);
          setValidation(null);
        }}
        value={draft}
      />
      {error ? (
        <Text
          className="mt-3 leading-6"
          color="danger"
          id={errorId}
          role="alert"
        >
          {hasConflict && !validation
            ? "This value changed elsewhere. Review the latest saved value before saving again."
            : error}
        </Text>
      ) : null}
      {hasConflict ? (
        <Button
          className="mt-3"
          color="tertiary"
          disabled={isBusy || disabled}
          onClick={() => void reviewLatest()}
          variant="outline"
        >
          {isReloading ? "Loading…" : "Review latest value"}
        </Button>
      ) : null}
      {reviewedValue !== null ? (
        <Text className="mt-3 leading-6" color="muted" role="status">
          Latest saved value: {reviewedValue}. Your changes are ready to review.
        </Text>
      ) : null}
      <Flex className="mt-3" gap={2} justify="end">
        <Button
          color="tertiary"
          disabled={isBusy}
          onClick={onClose}
          variant="outline"
        >
          Cancel
        </Button>
        <Button
          disabled={
            isBusy || disabled || hasConflict || draft === baseline.value
          }
          onClick={() => void save()}
        >
          {mutation.isPending ? "Saving…" : "Save"}
        </Button>
      </Flex>
    </Popover.Content>
  );
};

export const FieldValuePicker = ({
  field,
  value,
  disabled = false,
  people = [],
  variant = "solid",
  size = "sm",
  align = "start",
  onOpenChange,
  ...editorProps
}: FieldValuePickerProps) => {
  const [open, setOpen] = useState(false);
  const editorRef = useRef<FieldValueEditorHandle | null>(null);
  const formatted = formatCustomFieldValue(field, value, people);
  const label = `${getCustomFieldInputLabel(field)}${field.archivedAt ? " (archived)" : ""}`;
  const readOnly = disabled || Boolean(field.archivedAt);
  const changeOpen = (next: boolean) => {
    setOpen(next);
    onOpenChange?.(next);
  };
  const trigger = (
    <Button
      aria-label={`${label}: ${formatted}`}
      className="w-fit max-w-[13rem] justify-start gap-1 px-2 font-medium"
      color="tertiary"
      disabled={readOnly}
      leftIcon={
        <CustomFieldIcon className="h-4 w-auto shrink-0" field={field} />
      }
      onClick={(event) => {
        event.stopPropagation();
      }}
      onKeyDown={(event) => {
        event.stopPropagation();
      }}
      onPointerDown={(event) => {
        event.stopPropagation();
      }}
      size={size}
      title={`${label}: ${formatted}`}
      type="button"
      variant={variant}
    >
      <span className="min-w-0 truncate">{formatted}</span>
    </Button>
  );
  if (readOnly && !open) return trigger;
  return (
    <Popover
      onOpenChange={(next) => {
        if (!editorRef.current?.isBusy) changeOpen(next);
      }}
      open={open}
    >
      <Popover.Trigger asChild>{trigger}</Popover.Trigger>
      {open ? (
        <FieldValueEditor
          {...editorProps}
          align={align}
          disabled={readOnly}
          editorRef={editorRef}
          field={field}
          onClose={() => {
            changeOpen(false);
          }}
          people={people}
          value={value}
        />
      ) : null}
    </Popover>
  );
};
