import type { CustomField } from "@/modules/custom-fields/public/types";
import {
  createCustomField,
  getTeamCustomFields,
  updateCustomField,
} from "@/modules/custom-fields/public/import";
import { importCustomFields } from "./import-custom-fields";
import type { ImportDraft, ImportSourceCustomField } from "./schema";
import { createEmptyImportEntityCollections } from "./schema";

jest.mock("@/modules/custom-fields/public/import", () => ({
  archiveCustomField: jest.fn(),
  createCustomField: jest.fn(),
  getTeamCustomFields: jest.fn(),
  updateCustomField: jest.fn(),
}));
const field: CustomField = {
  id: "d350e64b-06cd-43e5-8bc8-66d756d54378",
  teamId: "c5f7e92d-3810-45a8-86fc-940835532701",
  name: "Budget",
  type: "money",
  icon: "goal",
  currency: "USD",
  options: [],
  showOnCreate: true,
  archivedAt: null,
  createdAt: "2026-10-01T12:00:00Z",
  updatedAt: "2026-10-01T12:00:00Z",
};
const source: ImportSourceCustomField = {
  sourceId: "source-budget",
  name: "Budget",
  type: "money",
  icon: "star",
  currency: "USD",
  teamSourceId: null,
  options: [],
};
const run = (definition: ImportSourceCustomField) => {
  const draft: ImportDraft = {
    ...createEmptyImportEntityCollections(),
    sourceType: "json",
    sourceNamespace: "fortyone:workspace:source",
    summary: "Backup",
    warnings: [],
    mapping: null,
    tasks: [],
    columns: [],
    rows: [],
    fileHash: "a".repeat(64),
    fileName: "backup.json",
    customFields: [definition],
  };
  return importCustomFields(
    { ctx: { workspaceSlug: "first" }, draft },
    { selectedTasks: [] },
    { getTargetTeamId: () => field.teamId },
    { peopleBySourceId: new Map(), resolveReviewedPerson: () => undefined },
  );
};
describe("imported custom field definitions", () => {
  beforeEach(() => {
    jest.mocked(getTeamCustomFields).mockReset().mockResolvedValue([]);
    jest.mocked(createCustomField).mockReset().mockResolvedValue(field);
    jest.mocked(updateCustomField).mockReset();
  });
  it.each(["star", "automation", "image", "workflow", null] as const)(
    "preserves the exported icon on a newly created definition (%s)",
    async (icon) => {
      const result = await run({ ...source, icon });
      expect(result.createdFields).toBe(1);
      expect(createCustomField).toHaveBeenCalledWith(
        field.teamId,
        expect.objectContaining({ icon }),
        { workspaceSlug: "first" },
      );
    },
  );
  it("keeps legacy automatic icons omitted and preserves matched destination icons", async () => {
    const { icon: _icon, ...legacy } = source;
    await run(legacy);
    expect(
      jest.mocked(createCustomField).mock.calls[0]?.[1],
    ).not.toHaveProperty("icon");
    jest.mocked(createCustomField).mockClear();
    jest.mocked(getTeamCustomFields).mockResolvedValue([field]);
    const result = await run(source);
    expect(result.createdFields).toBe(0);
    expect(createCustomField).not.toHaveBeenCalled();
    expect(updateCustomField).not.toHaveBeenCalled();
  });
});
