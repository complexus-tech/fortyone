import { readFileSync } from "node:fs";
import { join } from "node:path";

const readSource = (path: string) =>
  readFileSync(join(process.cwd(), path), "utf8");

const dialogSource = readSource("src/components/ui/new-objective/index.tsx");
const contentSource = readSource(
  "src/components/ui/new-objective/new-objective-dialog-content.tsx",
);
const headerSource = readSource(
  "src/components/ui/new-objective/new-objective-dialog-header.tsx",
);
const controlsSource = readSource(
  "src/components/ui/new-objective/new-objective-dialog-controls.tsx",
);
const keyResultsSource = readSource(
  "src/components/ui/new-objective/new-objective-key-results.tsx",
);
const keyResultsHookSource = readSource(
  "src/components/ui/new-objective/use-new-objective-key-results.ts",
);
const coordinatorSource = dialogSource.slice(
  dialogSource.indexOf("export const NewObjectiveDialog"),
);

describe("NewObjectiveDialog composition", () => {
  it("keeps the public dialog API and state coordinator focused", () => {
    expect(dialogSource).toContain("export const NewObjectiveDialog");
    expect(dialogSource).toContain("NewObjectiveDialogContent");
    expect(dialogSource).toContain("useNewObjectiveKeyResults");
    expect(dialogSource).toContain("handleSaveKeyResult");
    expect(coordinatorSource.split(/\r?\n/).length).toBeLessThanOrEqual(300);
  });

  it("keeps accessibility-bearing controls and key-result context with their sections", () => {
    expect(contentSource).toContain("<NewObjectiveDialogHeader");
    expect(contentSource).toContain("<NewObjectiveDialogControls");
    expect(contentSource).toContain("<NewObjectiveKeyResults");
    expect(headerSource).toContain('className="sr-only"');
    expect(controlsSource).toContain('aria-label="Remove date"');
    expect(controlsSource).toContain("event.stopPropagation()");
    expect(keyResultsSource).toContain("existingKeyResults");
    expect(keyResultsSource).toContain("qualityContext");
    expect(keyResultsHookSource).toContain("onKeyResultsChange");
    expect(keyResultsHookSource).toContain("editingIndex");
  });
});
