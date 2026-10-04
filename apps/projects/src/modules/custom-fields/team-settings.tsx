"use client";

import { useRef, useState } from "react";
import { Box, Button, Flex, Menu, Text } from "ui";
import { ArchiveIcon, EditIcon, MoreHorizontalIcon, PlusIcon } from "icons";
import { useUserRole } from "@/hooks/role";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { SectionHeader } from "@/components/ui/section-header";
import { CustomFieldEditor } from "./field-editor";
import { useCustomFieldMutations, useTeamCustomFields } from "./hooks";
import { CUSTOM_FIELD_TYPE_LABELS } from "./types";
import type { CustomField } from "./types";
import { CustomFieldIcon } from "./icons";

const CustomFieldActions = ({
  field,
  onEdit,
  onArchive,
}: {
  field: CustomField;
  onEdit: (trigger: HTMLButtonElement | null) => void;
  onArchive: () => void;
}) => {
  const triggerRef = useRef<HTMLButtonElement>(null);
  const openingEditor = useRef(false);
  return (
    <Menu>
      <Menu.Button asChild>
        <Button
          aria-label={`Actions for ${field.name}`}
          asIcon
          color="tertiary"
          ref={triggerRef}
          size="sm"
        >
          <MoreHorizontalIcon />
        </Button>
      </Menu.Button>
      <Menu.Items
        align="end"
        className="w-44"
        onCloseAutoFocus={(event) => {
          if (!openingEditor.current) return;
          event.preventDefault();
          openingEditor.current = false;
        }}
      >
        <Menu.Group>
          <Menu.Item
            onSelect={() => {
              openingEditor.current = true;
              onEdit(triggerRef.current);
            }}
          >
            <EditIcon aria-hidden="true" className="h-[1.15rem]" />
            Edit field
          </Menu.Item>
        </Menu.Group>
        <Menu.Separator />
        <Menu.Group>
          <Menu.Item onSelect={onArchive}>
            <ArchiveIcon aria-hidden="true" className="h-[1.15rem]" />
            Archive field
          </Menu.Item>
        </Menu.Group>
      </Menu.Items>
    </Menu>
  );
};

export const TeamCustomFieldSettings = ({ teamId }: { teamId: string }) => {
  const settingsRef = useRef<HTMLDivElement>(null);
  const {
    data: fields = [],
    isPending,
    isError,
    refetch,
  } = useTeamCustomFields(teamId);
  const { userRole } = useUserRole();
  const canManage = userRole === "admin";
  const [editor, setEditor] = useState<{
    field: CustomField | null;
    trigger: HTMLButtonElement | null;
  } | null>(null);
  const [archiving, setArchiving] = useState<CustomField | null>(null);
  const { archive } = useCustomFieldMutations(teamId);
  const active = fields.filter((field) => !field.archivedAt);
  const archived = fields.filter((field) => field.archivedAt);
  const confirmArchive = async () => {
    if (!archiving) return;
    try {
      await archive.mutateAsync(archiving.id);
      setArchiving(null);
    } catch {
      /* Keep the dialog and API error visible. */
    }
  };
  return (
    <Box
      aria-label="Custom fields"
      className="border-border bg-surface overflow-hidden rounded-2xl border"
      ref={settingsRef}
      role="region"
      tabIndex={-1}
    >
      <SectionHeader
        action={
          canManage ? (
            <Button
              className="shrink-0"
              color="tertiary"
              disabled={isPending || isError || active.length >= 50}
              leftIcon={<PlusIcon />}
              onClick={(event) => {
                setEditor({ field: null, trigger: event.currentTarget });
              }}
            >
              Create field
            </Button>
          ) : undefined
        }
        description="Add team properties for tasks, views and reports."
        title="Custom fields"
      />
      <Box>
        {!canManage ? (
          <Text className="px-6 pt-4" color="muted">
            Workspace admins manage fields.
          </Text>
        ) : null}
        {isPending ? (
          <Text aria-live="polite" className="px-6 py-6" color="muted">
            Loading fields…
          </Text>
        ) : null}
        {isError ? (
          <Box className="px-6 py-6">
            <Text color="danger" role="alert">
              Custom fields could not be loaded.
            </Text>
            <Button
              className="mt-3"
              color="tertiary"
              onClick={() => void refetch()}
              variant="outline"
            >
              Try again
            </Button>
          </Box>
        ) : null}
        {!isPending && !isError && active.length > 0 ? (
          <Box className="divide-border divide-y-[0.5px]">
            {active.map((field) => (
              <Flex
                align="center"
                className="hover:bg-state-hover/50 gap-4 px-6 py-4"
                justify="between"
                key={field.id}
                wrap
              >
                <Flex align="center" className="min-w-0 flex-1 gap-3">
                  <CustomFieldIcon
                    className="text-icon h-5 w-auto shrink-0"
                    field={field}
                  />
                  <Box className="min-w-0">
                    <Text className="break-words" fontWeight="medium">
                      {field.name}
                    </Text>
                    <Text className="mt-1" color="muted">
                      {CUSTOM_FIELD_TYPE_LABELS[field.type]}
                      {field.currency ? ` · ${field.currency}` : ""}
                      {field.showOnCreate ? " · On creation" : ""}
                    </Text>
                  </Box>
                </Flex>
                {canManage ? (
                  <CustomFieldActions
                    field={field}
                    onArchive={() => {
                      archive.reset();
                      setArchiving(field);
                    }}
                    onEdit={(trigger) => {
                      setEditor({ field, trigger });
                    }}
                  />
                ) : null}
              </Flex>
            ))}
          </Box>
        ) : null}
        {!isPending && !isError && active.length === 0 ? (
          <Box className="px-6 py-6">
            <Text color="muted">
              No custom fields yet. Add details such as customers, amounts,
              categories, or dates to your work.
            </Text>
          </Box>
        ) : null}
        {archived.length ? (
          <details className="border-border border-t-[0.5px] px-6 py-4">
            <summary className="cursor-pointer font-medium">
              Archived fields ({archived.length})
            </summary>
            <Box className="mt-3 space-y-3">
              {archived.map((field) => (
                <Flex className="gap-3" justify="between" key={field.id}>
                  <Flex align="center" className="min-w-0 gap-3">
                    <CustomFieldIcon
                      className="text-icon h-5 w-auto shrink-0"
                      field={field}
                    />
                    <Text className="min-w-0 break-words">{field.name}</Text>
                  </Flex>
                  <Text color="muted">
                    {CUSTOM_FIELD_TYPE_LABELS[field.type]}
                  </Text>
                </Flex>
              ))}
            </Box>
            <Text className="mt-3 leading-6" color="muted">
              Existing values stay available on work and in historical reports.
            </Text>
          </details>
        ) : null}
        {active.length >= 50 ? (
          <Text className="px-6 pb-4" color="muted">
            This team has reached the 50-field limit.
          </Text>
        ) : null}
      </Box>
      {editor ? (
        <CustomFieldEditor
          field={editor.field}
          key={editor.field?.id ?? "new"}
          onClose={() => {
            setEditor(null);
          }}
          onReturnFocus={() => {
            if (editor.trigger?.isConnected && !editor.trigger.disabled)
              editor.trigger.focus();
            else settingsRef.current?.focus();
          }}
          teamId={teamId}
        />
      ) : null}
      <ConfirmDialog
        confirmText="Archive field"
        description="Stop new values while keeping existing history and reports."
        errorMessage={archive.error?.message}
        isLoading={archive.isPending}
        isOpen={Boolean(archiving)}
        loadingText="Archiving…"
        onClose={() => {
          setArchiving(null);
        }}
        onConfirm={() => void confirmArchive()}
        title={`Archive ${archiving?.name ?? "field"}?`}
      />
    </Box>
  );
};
