"use client";

import { useState } from "react";
import { Box, Button, Command, Popover, Text } from "ui";
import { PlusIcon, WarningIcon } from "icons";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { useTeamCustomFields } from "./hooks";
import type { CustomFieldValueMap } from "./types";
import {
  prepareCustomFieldValues,
  getCustomFieldInputLabel,
} from "./value-utils";
import { CreateFieldChip } from "./create-field-chip";
import { CustomFieldIcon } from "./icons";

export const useCreateCustomFields = (teamId?: string) => {
  const query = useTeamCustomFields(teamId);
  const members = useTeamMembers(teamId);
  const [draft, setDraft] = useState<{
    teamId?: string;
    values: CustomFieldValueMap;
  }>({ teamId, values: {} });
  const values = draft.teamId === teamId ? draft.values : {};
  const setValues = (next: CustomFieldValueMap) => {
    setDraft({ teamId, values: next });
  };
  return {
    ...query,
    fields: query.data ?? [],
    values,
    setValues,
    reset: () => {
      setDraft({ teamId, values: {} });
    },
    prepareValues: () => {
      if (Object.keys(values).length && !query.data)
        throw new Error(
          "Custom fields have not loaded. Try again before creating work.",
        );
      const personFieldIds = new Set(
        (query.data ?? [])
          .filter((field) => field.type === "person")
          .map((field) => field.id),
      );
      const hasPersonValues = Object.entries(values).some(
        ([id, value]) => personFieldIds.has(id) && value,
      );
      if (hasPersonValues && !members.data)
        throw new Error(
          "Team members have not loaded. Try again before creating work.",
        );
      const memberIds = new Set(
        (members.data ?? [])
          .filter((member) => member.isActive && !member.isSystem)
          .map((member) => member.id),
      );
      const applicableValues = Object.fromEntries(
        Object.entries(values).filter(
          ([id, value]) =>
            !personFieldIds.has(id) || !value || memberIds.has(value),
        ),
      );
      return prepareCustomFieldValues(query.data ?? [], applicableValues);
    },
  };
};

export const CreateCustomFields = ({
  teamId,
  values,
  onChange,
  disabled,
}: {
  teamId?: string;
  values: CustomFieldValueMap;
  onChange: (values: CustomFieldValueMap) => void;
  disabled?: boolean;
}) => {
  const {
    data: definitions = [],
    isPending,
    isError,
    refetch,
  } = useTeamCustomFields(teamId);
  const [pickerOpen, setPickerOpen] = useState(false);
  const [revealed, setRevealed] = useState<{
    teamId?: string;
    ids: string[];
    openId?: string;
  }>({ teamId, ids: [] });
  const revealedIds = revealed.teamId === teamId ? revealed.ids : [];
  const active = definitions.filter((field) => !field.archivedAt);
  const fields = active.filter(
    (field) =>
      field.showOnCreate ||
      revealedIds.includes(field.id) ||
      Boolean(values[field.id]),
  );
  const visibleIds = new Set(fields.map((field) => field.id));
  const additional = active.filter((field) => !visibleIds.has(field.id));
  if (!teamId) return null;
  if (isError)
    return (
      <Box className="order-12">
        <Text className="sr-only" role="alert">
          Custom fields could not be loaded.
        </Text>
        <Button
          className="dark:bg-surface-elevated/90 gap-1.5 px-2"
          color="tertiary"
          leftIcon={<WarningIcon className="h-4.5 w-auto" />}
          onClick={() => void refetch()}
          size="sm"
          type="button"
          variant="outline"
        >
          Retry properties
        </Button>
      </Box>
    );
  if (isPending || !active.length) return null;
  return (
    <>
      {fields.map((field) => (
        <CreateFieldChip
          disabled={disabled}
          field={field}
          initiallyOpen={
            revealed.teamId === teamId && revealed.openId === field.id
          }
          key={`${teamId}:${field.id}`}
          onChange={(value) => {
            onChange({ ...values, [field.id]: value });
          }}
          value={values[field.id] ?? null}
        />
      ))}
      {additional.length ? (
        <Box className="order-12">
          <Popover onOpenChange={setPickerOpen} open={pickerOpen}>
            <Popover.Trigger asChild>
              <Button
                className="dark:bg-surface-elevated/90 gap-1.5 px-2"
                color="tertiary"
                disabled={disabled}
                leftIcon={<PlusIcon className="h-4 w-auto" />}
                size="sm"
                type="button"
                variant="outline"
              >
                Properties
              </Button>
            </Popover.Trigger>
            <Popover.Content
              align="start"
              className="w-80 max-w-[calc(100vw-2rem)]"
            >
              <Command>
                <Command.Input autoFocus placeholder="Find a property..." />
                <Command.Empty>
                  <Text color="muted">No properties found.</Text>
                </Command.Empty>
                <Command.Group className="max-h-64 overflow-y-auto">
                  {additional.map((field) => (
                    <Command.Item
                      key={field.id}
                      keywords={[field.name, field.currency ?? ""]}
                      onSelect={() => {
                        setRevealed({
                          teamId,
                          ids: [...revealedIds, field.id],
                          openId: field.id,
                        });
                        setPickerOpen(false);
                      }}
                      value={field.id}
                    >
                      <CustomFieldIcon
                        className="h-4 w-auto shrink-0"
                        field={field}
                      />
                      <Text>{getCustomFieldInputLabel(field)}</Text>
                    </Command.Item>
                  ))}
                </Command.Group>
              </Command>
            </Popover.Content>
          </Popover>
        </Box>
      ) : null}
    </>
  );
};
