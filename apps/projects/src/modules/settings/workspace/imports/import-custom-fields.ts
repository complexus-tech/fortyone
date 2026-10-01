import type {
  CustomField,
  CustomFieldValue,
} from "@/modules/custom-fields/public/types";
import {
  archiveCustomField,
  createCustomField,
  getTeamCustomFields,
  updateCustomField,
  validateCustomFieldValue,
} from "@/modules/custom-fields/public/import";
import type { RunImportInput } from "./import-run-model";
import type { ImportSelection } from "./import-selection";
import type { ImportTeamDestinations } from "./import-team-destinations";
import type { ImportPeople } from "./import-people";
import type { ImportSourceCustomField, ImportTask } from "./schema";
import { normalizeImportMatch } from "./import-entity-matching";

type FieldMapping = {
  field: CustomField;
  source: ImportSourceCustomField;
  options: Map<string, string>;
  created: boolean;
};
export const importCustomFields = async (
  { ctx, draft }: Pick<RunImportInput, "ctx" | "draft">,
  { selectedTasks }: Pick<ImportSelection, "selectedTasks">,
  { getTargetTeamId }: Pick<ImportTeamDestinations, "getTargetTeamId">,
  {
    peopleBySourceId,
    resolveReviewedPerson,
  }: Pick<ImportPeople, "peopleBySourceId" | "resolveReviewedPerson">,
) => {
  const sources = new Map(
    (draft.customFields ?? []).map((field) => [field.sourceId, field]),
  );
  const targets = new Map<
    string,
    { teamId: string; source: ImportSourceCustomField }
  >();
  for (const source of sources.values()) {
    const teamId = getTargetTeamId(source.teamSourceId);
    if (teamId) targets.set(`${teamId}:${source.sourceId}`, { teamId, source });
  }
  for (const task of selectedTasks) {
    const teamId = getTargetTeamId(task.teamSourceId);
    if (!teamId) continue;
    for (const value of task.customFieldValues ?? []) {
      const source = sources.get(value.sourceFieldId);
      if (source)
        targets.set(`${teamId}:${source.sourceId}`, { teamId, source });
    }
  }
  const definitionsByTeam = new Map<string, CustomField[]>();
  const mappings = new Map<string, FieldMapping>();
  let createdFields = 0;
  const issues: string[] = [];
  for (const [key, { teamId, source }] of targets) {
    try {
      let definitions = definitionsByTeam.get(teamId);
      if (!definitions) {
        // eslint-disable-next-line no-await-in-loop -- One live definition read per target team, then sequential creation prevents duplicate names on retry.
        definitions = await getTeamCustomFields(teamId, ctx);
        definitionsByTeam.set(teamId, definitions);
      }
      const sameName = definitions.filter(
        (field) =>
          normalizeImportMatch(field.name) ===
          normalizeImportMatch(source.name),
      );
      const compatible = sameName.filter(
        (field) =>
          !field.archivedAt &&
          field.type === source.type &&
          field.currency === source.currency,
      );
      if (
        sameName.length > 0 &&
        (sameName.length !== 1 || compatible.length !== 1)
      ) {
        issues.push(
          `${source.name}: an existing field has an incompatible or ambiguous definition.`,
        );
        continue;
      }
      let field = compatible.at(0);
      let created = false;
      if (!field) {
        // eslint-disable-next-line no-await-in-loop -- Definitions are reused by a unique live name/type/currency match before creating.
        field = await createCustomField(
          teamId,
          {
            name: source.name,
            type: source.type,
            ...(source.icon !== undefined ? { icon: source.icon } : {}),
            currency: source.currency,
            showOnCreate: false,
            options: source.options.map((option) => ({
              id: crypto.randomUUID(),
              name: option.name,
            })),
          },
          ctx,
        );
        definitions.push(field);
        created = true;
        createdFields += 1;
      }
      const matchedField = field;
      const missingOptions = source.options.filter(
        (sourceOption) =>
          !matchedField.options.some(
            (option) =>
              normalizeImportMatch(option.name) ===
              normalizeImportMatch(sourceOption.name),
          ),
      );
      if (!created && field.type === "select" && missingOptions.length) {
        // eslint-disable-next-line no-await-in-loop -- Add source options without replacing existing IDs, names or archived choices.
        field = await updateCustomField(
          teamId,
          field.id,
          {
            name: field.name,
            showOnCreate: field.showOnCreate,
            options: [
              ...field.options
                .filter((option) => !option.archivedAt)
                .map(({ id, name }) => ({ id, name })),
              ...missingOptions.map((option) => ({
                id: crypto.randomUUID(),
                name: option.name,
              })),
            ],
          },
          ctx,
        );
      }
      const options = new Map<string, string>();
      for (const option of source.options) {
        const matches = field.options.filter(
          (candidate) =>
            !candidate.archivedAt &&
            normalizeImportMatch(candidate.name) ===
              normalizeImportMatch(option.name),
        );
        if (matches.length === 1) options.set(option.sourceId, matches[0].id);
      }
      mappings.set(key, { field, source, options, created });
    } catch (error) {
      issues.push(
        `${source.name}: ${error instanceof Error ? error.message : "the field could not be prepared."}`,
      );
    }
  }
  let unresolvedValues = 0;
  const getValues = (task: ImportTask, teamId: string): CustomFieldValue[] =>
    (task.customFieldValues ?? []).flatMap((sourceValue) => {
      const mapping = mappings.get(`${teamId}:${sourceValue.sourceFieldId}`);
      if (!mapping) {
        unresolvedValues += 1;
        return [];
      }
      let value = sourceValue.value;
      if (value !== null && mapping.field.type === "select")
        value = mapping.options.get(value) ?? null;
      if (sourceValue.value !== null && mapping.field.type === "person") {
        const person = resolveReviewedPerson(
          peopleBySourceId.get(sourceValue.value),
          teamId,
          sourceValue.value,
        );
        value = person?.id ?? null;
      }
      if (
        (sourceValue.value !== null && value === null) ||
        validateCustomFieldValue(mapping.field, value) !== null
      ) {
        unresolvedValues += 1;
        return [];
      }
      return [{ fieldId: mapping.field.id, value }];
    });
  const finalizeArchives = async () => {
    for (const { field, source, created } of mappings.values()) {
      // Existing destination definitions are never archived by an import.
      if (!created) continue;
      try {
        if (source.archivedAt) {
          // eslint-disable-next-line no-await-in-loop -- Archive only after values were created under an active definition.
          await archiveCustomField(field.teamId, field.id, ctx);
        } else if (source.options.some((option) => option.archivedAt)) {
          const archivedNames = new Set(
            source.options
              .filter((option) => option.archivedAt)
              .map((option) => normalizeImportMatch(option.name)),
          );
          // eslint-disable-next-line no-await-in-loop -- Preserve historical select values first; omitted option IDs archive without rewriting values.
          await updateCustomField(
            field.teamId,
            field.id,
            {
              name: field.name,
              showOnCreate: field.showOnCreate,
              options: field.options
                .filter(
                  (option) =>
                    !archivedNames.has(normalizeImportMatch(option.name)),
                )
                .map(({ id, name }) => ({ id, name })),
            },
            ctx,
          );
        }
      } catch (error) {
        issues.push(
          `${source.name}: Values were saved, but archive state could not be restored. ${error instanceof Error ? error.message : "Retry the archive change."}`,
        );
      }
    }
  };
  return {
    createdFields,
    issues,
    getValues,
    getUnresolvedValues: () => unresolvedValues,
    finalizeArchives,
  };
};
