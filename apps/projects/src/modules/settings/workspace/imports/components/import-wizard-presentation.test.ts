/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

const source = readFileSync(
  join(
    process.cwd(),
    "src/modules/settings/workspace/imports/components/import-wizard.tsx",
  ),
  "utf8",
);
const selectSource = readFileSync(
  join(process.cwd(), "../../packages/ui/src/select.tsx"),
  "utf8",
);

describe("ImportWizard presentation", () => {
  it("keeps the whole dialog inside the viewport with one scrollable body", () => {
    expect(source).toContain(
      'className="mt-0 flex max-h-[calc(100dvh-2rem)] max-w-3xl flex-col md:mt-0"',
    );
    expect(source).toContain('overlayClassName="items-center py-4"');
    expect(source).toContain('className="shrink-0 px-6 pt-5 pb-2"');
    expect(source).toContain(
      'className="min-h-0 flex-1 overflow-y-auto overscroll-contain px-6 pt-4 pb-6"',
    );
    expect(source).toContain(
      'className="bg-surface-muted/35 shrink-0 gap-3 px-6 py-4"',
    );
    expect(source).not.toContain("max-h-[min(68dvh,700px)]");
  });

  it("uses the segmented Art Circles analysis treatment without a spinner", () => {
    expect(source).not.toContain("LoadingIcon");
    expect(source).not.toContain("animate-spin");
    expect(source).toContain("IMPORT_ANALYSIS_PHASES");
    expect(source).toContain("Uploading your file");
    expect(source).toContain("Reading your export");
    expect(source).toContain("Mapping your work");
    expect(source).toContain("Ready to organize");
    expect(source).toContain("Ready with standard mapping");
    expect(source).toContain(
      "bg-[linear-gradient(100deg,rgba(102,121,248,0.07)_0%,rgba(243,90,168,0.07)_100%)]",
    );
    expect(source).toContain("animate-pulse motion-reduce:animate-none");
    expect(source).toContain("motion-reduce:transition-none");
    expect(source).toContain(
      "flex-col items-stretch sm:flex-row sm:items-start",
    );
    expect(source).toContain('role={analysisError ? "alert" : "status"}');
    expect(source).toContain('aria-busy="true"');
    expect(source).toContain("<ProgressBar");
    expect(source).toContain("% complete");
    expect(source).not.toContain("analysisIndicator");
  });

  it("uses established bordered team controls", () => {
    expect(source).toContain('selected && "border-primary bg-primary/5');
    expect(source).toContain('selected && "bg-primary/10 text-primary"');
    expect(source).toContain("icon={<PlusIcon />}");
    expect(source).toContain("<DestinationTeamPicker");
    expect(source).toContain("<Command.Input");
    expect(source).toContain('<Command.Empty className="py-3 text-base">');
    expect(source).toContain("<ColorPicker");
    expect(source).not.toContain('type="color"');
    expect(source).not.toContain("<select");
  });

  it("keeps review focused and gives both review groups a border", () => {
    expect(source).toContain("Review your import");
    expect(source).toContain("DO_NOT_IMPORT_VALUE");
    expect(source).toContain("<Select.Trigger");
    expect(source).toContain("Task review");
    expect(source).toContain("sm:flex-row sm:items-center");
    expect(source).not.toContain("Review before importing");
    expect(source).not.toContain("AnalysisWarnings");
    expect(source).not.toContain("How destination fields match");
  });

  it("uses icon-free inverted primary actions", () => {
    expect(source).toContain('color="invert"');
    expect(source).not.toContain('color="gradient"');
    expect(source).not.toMatch(
      /rightIcon=\{<ArrowRight2Icon \/>}[^]*?Continue/,
    );
  });
});

describe("Select presentation", () => {
  it("reserves space and anchors the chevron away from the edge", () => {
    expect(selectSource).toContain("pr-8 pl-3");
    expect(selectSource).toContain("min-w-0 flex-1 truncate pr-6 text-left");
    expect(selectSource).toContain(
      "absolute top-1/2 right-3.5 h-3.5 w-auto shrink-0",
    );
  });
});
