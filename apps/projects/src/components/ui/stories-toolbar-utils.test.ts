import { readFileSync } from "node:fs";
import { join } from "node:path";
import { formatBulkStoryDeadline } from "./stories-toolbar-utils";

describe("formatBulkStoryDeadline", () => {
  it("serializes a picked day using the API date-only contract", () => {
    const pickedDay = new Date(2026, 8, 1, 17, 45, 30);

    expect(formatBulkStoryDeadline(pickedDay)).toBe("2026-09-01");
  });

  it("puts every shared story-property picker into explicit bulk mode", () => {
    const toolbarSource = readFileSync(
      join(process.cwd(), "src/components/ui/stories-toolbar.tsx"),
      "utf8",
    );

    expect(toolbarSource.match(/selectionMode="bulk"/g)).toHaveLength(5);
    expect(toolbarSource).toContain("endDate: formatBulkStoryDeadline(day)");
  });
});
