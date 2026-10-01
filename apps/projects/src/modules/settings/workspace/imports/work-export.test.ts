import { createEmptyImportEntityCollections, importTaskSchema } from "./schema";
import { workExportSchema } from "./backup-format";
import { workExportToCSV } from "./work-export";

const task = importTaskSchema.parse({
  sourceId: "task-1",
  title: "Preview",
  description: "Preview description",
  status: "Won",
  statusCategory: "completed",
  priority: "No Priority",
  estimateValue: null,
  estimatedDurationMinutes: null,
  minimumFocusBlockMinutes: null,
  assigneeEmail: null,
  assigneeName: null,
  assigneePersonSourceId: null,
  collaboratorPersonSourceIds: [],
  teamSourceId: "team-1",
  parentSourceId: null,
  objectiveSourceId: null,
  keyResultSourceId: null,
  sprintSourceId: null,
  labelSourceIds: [],
  associations: [],
  links: [],
  startDate: null,
  endDate: null,
});

it("exports canonical content, exact decimals and option labels without spreadsheet formulas", () => {
  const backup = workExportSchema.parse({
    format: "fortyone-work-export",
    version: 1,
    analysis: {
      sourceType: "json",
      sourceNamespace: "fortyone:workspace:test",
      summary: "Work export",
      warnings: [],
      mapping: null,
      ...createEmptyImportEntityCollections(),
      teams: [
        {
          sourceId: "team-1",
          name: "Sales",
          code: "SAL",
          color: "#123456",
          description: null,
          isPrivate: false,
        },
      ],
      tasks: [task],
    },
    customFields: [
      {
        sourceId: "field-amount",
        name: "Amount",
        type: "money",
        currency: "USD",
        teamSourceId: "team-1",
        options: [],
      },
      {
        sourceId: "field-stage",
        name: "Stage",
        type: "select",
        currency: null,
        teamSourceId: "team-1",
        options: [{ sourceId: "option-won", name: "=Unsafe source label" }],
      },
    ],
    taskData: [
      {
        sourceId: "task-1",
        title: "=SUM(A1:A2)",
        description: 'Line one\nLine "two"',
        comments: [],
        customFieldValues: [
          { sourceFieldId: "field-amount", value: "9007199254740993.01" },
          { sourceFieldId: "field-stage", value: "option-won" },
        ],
      },
    ],
  });
  const csv = workExportToCSV(backup);
  expect(csv).toContain('"9007199254740993.01"');
  expect(csv).toContain('"\'=SUM(A1:A2)"');
  expect(csv).toContain('"\'=Unsafe source label"');
  expect(csv).toContain('"Line one\nLine ""two"""');
  expect(csv).not.toContain("Preview description");
  expect(csv).toContain("Amount (USD)");
});
