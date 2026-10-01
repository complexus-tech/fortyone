"use client";

import { useState } from "react";
import { PlusIcon } from "icons";
import { Box, Button, Text } from "ui";
import { PropertyOption } from "@/components/ui/property-option";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { useStoryCustomFields } from "./hooks";
import type { CustomField } from "./types";
import { getCustomFieldInputLabel, isCustomFieldValueSet } from "./value-utils";
import { FieldValuePicker } from "./field-value-picker";

type PropertyLayoutProps = {
  isCompact: boolean;
  isNotifications: boolean;
};

const FieldProperty = ({
  field,
  isCompact,
  isNotifications,
  ...pickerProps
}: PropertyLayoutProps & {
  field: CustomField;
  storyId: string;
  value: string | null;
  version?: number;
  disabled: boolean;
  onOpenChange: (open: boolean) => void;
  onReload: () => Promise<{ value: string | null; version?: number }>;
}) => {
  const { data: people = [] } = useTeamMembers(
    field.type === "person" ? field.teamId : undefined,
  );
  const label = `${getCustomFieldInputLabel(field)}${field.archivedAt ? " (archived)" : ""}`;
  return (
    <PropertyOption
      isCompact={isCompact}
      isNotifications={isNotifications}
      label={label}
      value={
        <FieldValuePicker
          {...pickerProps}
          align={isCompact ? "start" : "end"}
          field={field}
          people={people}
          variant={isCompact ? "solid" : "naked"}
        />
      }
    />
  );
};

type StoryCustomFieldPropertiesProps = {
  storyId: string;
  disabled?: boolean;
  isCompact?: boolean;
  isNotifications?: boolean;
};

const StoryCustomFieldPropertiesContent = ({
  storyId,
  disabled = false,
  isCompact = false,
  isNotifications = false,
}: StoryCustomFieldPropertiesProps) => {
  const { data, isError, refetch } = useStoryCustomFields(storyId);
  const [showEmpty, setShowEmpty] = useState(false);
  const [editingFields, setEditingFields] = useState(new Set<string>());
  const loadFeedback =
    isError || !data ? (
      <PropertyOption
        isCompact={isCompact}
        isNotifications={isNotifications}
        label="Custom fields"
        value={
          isError ? (
            <Box>
              <Text color="danger" role="alert">
                Fields could not be loaded.
              </Text>
              <Button
                className="mt-2"
                color="tertiary"
                onClick={() => void refetch()}
                size="sm"
                variant="outline"
              >
                Try again
              </Button>
            </Box>
          ) : (
            <Text color="muted" role="status">
              Loading fields…
            </Text>
          )
        }
      />
    ) : null;
  if (!data?.fields.length) return loadFeedback;
  const values = new Map(
    data.values.map((value) => [value.fieldId, value.value]),
  );
  const fields = data.fields.filter(
    (field) =>
      showEmpty ||
      editingFields.has(field.id) ||
      isCustomFieldValueSet(values.get(field.id)),
  );
  const emptyCount = data.fields.filter(
    (field) => !isCustomFieldValueSet(values.get(field.id)),
  ).length;
  return (
    <>
      {loadFeedback}
      {fields.map((field) => (
        <FieldProperty
          disabled={disabled}
          field={field}
          isCompact={isCompact}
          isNotifications={isNotifications}
          key={field.id}
          onOpenChange={(open) => {
            setEditingFields((current) => {
              const next = new Set(current);
              if (open) next.add(field.id);
              else next.delete(field.id);
              return next;
            });
          }}
          onReload={async () => {
            const latest = await refetch({ throwOnError: true });
            if (!latest.data) throw new Error("Fields could not be reloaded.");
            return {
              value:
                latest.data.values.find((item) => item.fieldId === field.id)
                  ?.value ?? null,
              version: latest.data.version,
            };
          }}
          storyId={storyId}
          value={values.get(field.id) ?? null}
          version={data.version}
        />
      ))}
      {emptyCount > 0 ? (
        <PropertyOption
          isCompact={isCompact}
          isNotifications={isNotifications}
          label="Optional fields"
          value={
            <Button
              className="w-fit justify-start font-medium"
              color="tertiary"
              leftIcon={<PlusIcon className="h-4 w-auto" />}
              onClick={() => {
                setShowEmpty((current) => !current);
              }}
              size="sm"
              type="button"
              variant={isCompact ? "solid" : "naked"}
            >
              {showEmpty ? "Hide empty fields" : `Add details (${emptyCount})`}
            </Button>
          }
        />
      ) : null}
    </>
  );
};

export const StoryCustomFieldProperties = (
  props: StoryCustomFieldPropertiesProps,
) => <StoryCustomFieldPropertiesContent key={props.storyId} {...props} />;
