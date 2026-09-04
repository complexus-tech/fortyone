"use client";
import { cn } from "lib";
import { Box, Select, Text } from "ui";
import type { ImportDraft, ImportMapping } from "../schema";
import { isImportMappingFieldLocked } from "./import-wizard-safety";
import { DO_NOT_IMPORT_VALUE } from "./import-draft-model";
import { useImportTerms } from "./use-import-terms";

const MAPPING_FIELDS: {
  key: keyof ImportMapping;
  label: string;
  required?: boolean;
}[] = [
  { key: "title", label: "Title", required: true },
  { key: "description", label: "Description" },
  { key: "status", label: "Status" },
  { key: "priority", label: "Priority" },
  { key: "assigneeEmail", label: "Assignee email" },
  { key: "startDate", label: "Start date" },
  { key: "endDate", label: "Due date" },
  { key: "sourceId", label: "Source ID" },
];

export const ImportFieldMapping = ({
  draft,
  hasAttemptedImport,
  updateMapping,
}: {
  draft: ImportDraft;
  hasAttemptedImport: boolean;
  updateMapping: (field: keyof ImportMapping, value: string) => void;
}) => {
  const { storyTerm, storyTermPlural } = useImportTerms();
  return (
    <>
      {draft.mapping && draft.columns.length > 0 ? (
        <Box className="border-border bg-surface mt-5 rounded-xl border-[0.5px] p-5">
          <Text className="font-medium">Field mapping</Text>
          {hasAttemptedImport ? (
            <Box
              className="bg-warning/8 mt-3 rounded-xl p-3"
              id="import-source-id-lock-note"
            >
              <Text className="font-medium">
                Source ID is locked for safe retries
              </Text>
              <Text className="mt-1 leading-6" color="muted">
                Retries use stable source IDs to recognize {storyTermPlural}
                already created and complete only unfinished work. Previously
                created {storyTermPlural} are reused as-is;
                {storyTerm}-field edits apply only to {storyTermPlural}
                that have not been created yet. Upload a new file to change the
                Source ID mapping.
              </Text>
            </Box>
          ) : null}
          <Box className="mt-4 grid gap-x-5 gap-y-3 md:grid-cols-2">
            {MAPPING_FIELDS.map(({ key, label, required }) => {
              const mappingFieldLocked = isImportMappingFieldLocked(
                key,
                hasAttemptedImport,
              );
              return (
                <label key={key}>
                  <span className="mb-1.5 block font-medium">
                    {label}
                    {required ? <span className="text-danger"> *</span> : null}
                  </span>
                  <Select
                    disabled={mappingFieldLocked}
                    onValueChange={(value) => {
                      updateMapping(key, value);
                    }}
                    value={draft.mapping?.[key] ?? DO_NOT_IMPORT_VALUE}
                  >
                    <Select.Trigger
                      aria-describedby={
                        mappingFieldLocked
                          ? "import-source-id-lock-note"
                          : undefined
                      }
                      className={cn(
                        "bg-surface h-11 rounded-xl text-base",
                        mappingFieldLocked && "cursor-not-allowed opacity-60",
                      )}
                    >
                      <Select.Input />
                    </Select.Trigger>
                    <Select.Content>
                      <Select.Option
                        className="text-base"
                        value={DO_NOT_IMPORT_VALUE}
                      >
                        Do not import
                      </Select.Option>
                      {draft.columns.map((column) => (
                        <Select.Option
                          className="text-base"
                          key={column}
                          value={column}
                        >
                          {column}
                        </Select.Option>
                      ))}
                    </Select.Content>
                  </Select>
                </label>
              );
            })}
          </Box>
        </Box>
      ) : null}
    </>
  );
};
