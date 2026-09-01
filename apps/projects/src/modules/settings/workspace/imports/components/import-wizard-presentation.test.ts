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
const graphSource = readFileSync(
  join(
    process.cwd(),
    "src/modules/settings/workspace/imports/components/import-graph-review.tsx",
  ),
  "utf8",
);
const memberPickerSource = readFileSync(
  join(
    process.cwd(),
    "src/modules/settings/workspace/imports/components/import-member-picker.tsx",
  ),
  "utf8",
);
const runImportSource = readFileSync(
  join(process.cwd(), "src/modules/settings/workspace/imports/run-import.ts"),
  "utf8",
);
const thinkingSource = readFileSync(
  join(process.cwd(), "src/components/ui/chat/thinking.tsx"),
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

  it("only dismisses the importer through an explicit action", () => {
    expect(source).toMatch(
      /onInteractOutside=\{\(event\) => \{\s*event\.preventDefault\(\);/,
    );
    expect(source).toMatch(
      /onEscapeKeyDown=\{\(event\) => \{\s*event\.preventDefault\(\);/,
    );
    expect(source).not.toContain("hideClose");
  });

  it("uses the segmented Art Circles analysis treatment without a spinner", () => {
    expect(source).not.toContain("LoadingIcon");
    expect(source).not.toContain("animate-spin");
    expect(source).toContain("IMPORT_ANALYSIS_PHASES");
    expect(source).toContain("Uploading your file");
    expect(source).toContain("Reading your export");
    expect(source).toContain("Mapping your work");
    expect(source).toContain("Ready to choose teams");
    expect(source).toContain("Ready with standard mapping");
    expect(source).toContain('activeLabel: "Uploading"');
    expect(source).toContain('completeLabel: "Uploaded"');
    expect(source).toContain('activeLabel: "Reading"');
    expect(source).toContain('completeLabel: "Read"');
    expect(source).toContain('activeLabel: "Mapping"');
    expect(source).toContain('completeLabel: "Mapped"');
    expect(source).toContain('activeLabel: "Preparing"');
    expect(source).toContain('completeLabel: "Prepared"');
    expect(source).toContain("<Thinking");
    expect(source).toContain('<CheckIcon className="h-3.5 w-auto"');
    expect(source).toContain('className="text-foreground"');
    expect(source).toContain('className="text-base font-medium" color="muted"');
    expect(source).toContain('className="min-h-6 gap-1 text-base font-medium"');
    expect(source).not.toContain(
      '<Text className="text-xs font-medium">{item.label}</Text>',
    );
    expect(thinkingSource).toContain("motion-reduce:animate-none");
    expect(source).toContain(
      "from-secondary/10 via-info/10 to-primary/10 pointer-events-none absolute inset-0 animate-pulse bg-linear-to-r motion-reduce:animate-none",
    );
    expect(source).toContain(
      "from-secondary via-info to-primary bg-linear-to-r bg-clip-text text-transparent",
    );
    expect(source).toContain(
      '"text-[0.8125rem] font-semibold tracking-[0.08em] uppercase"',
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

  it("uses compact team controls", () => {
    expect(source).toContain('selected && "border-primary bg-primary/5');
    expect(source).toContain('selected && "bg-primary/10 text-primary"');
    expect(source).toContain("relative min-h-20 rounded-xl");
    expect(source).toContain("flex h-9 w-9 shrink-0");
    expect(source).toContain("icon={<PlusIcon />}");
    expect(source).toContain("<DestinationTeamPicker");
    expect(source).toContain('className="mt-3 max-w-full"');
    expect(source).toContain('label="Destination teams"');
    expect(source).toContain(
      'className="z-[60] w-80 max-w-[calc(100vw-2rem)]"',
    );
    expect(source.match(/"Destination team"/g)).toHaveLength(1);
    expect(source).not.toContain('className="h-11 w-full"');
    expect(source).toContain("<Command.Input");
    expect(source).toContain('<Command.Empty className="py-3 text-base">');
    expect(source).toContain("<ColorPicker");
    expect(source).not.toContain('type="color"');
    expect(source).not.toContain("<select");
  });

  it("keeps team and member setup concise", () => {
    expect(source).toContain(
      "Reuse or create source teams, joining matched teams when needed.",
    );
    expect(source).toContain("Map members");
    expect(source).toContain("Source members");
    expect(source).toContain("No source members need mapping");
    expect(source).not.toContain("<ImportPeoplePolicy");
    expect(graphSource).not.toContain("ImportPeoplePolicy");
    expect(source).not.toContain(
      "Add matched workspace members to imported teams",
    );
    expect(graphSource).not.toContain(
      "Matched members can join destination teams when needed.",
    );
    expect(graphSource).not.toContain(
      "No accounts or invitations are created.",
    );
    expect(source).toContain("border-border/70 bg-surface/60 mt-4");
    expect(source).not.toContain("Email matches are automatic");
    expect(source).not.toContain("No safe workspace match");
    expect(graphSource).not.toContain("When on, exact source-email matches");
    expect(source).not.toContain(
      "Final assignment is checked in each destination team.",
    );
  });

  it("separates teams, members, review, and import into five steps", () => {
    expect(source).toContain(
      'const STEPS = ["Upload", "Teams", "Members", "Review", "Import"]',
    );
    expect(source).toContain("step === WIZARD_STEP.teams");
    expect(source).toContain("step === WIZARD_STEP.members");
    expect(source).toContain("step === WIZARD_STEP.review && draft");
    expect(source).toContain("step === WIZARD_STEP.import");
    expect(source).toContain("step < WIZARD_STEP.review");
    expect(source).toContain("setStep(WIZARD_STEP.import)");
    expect(source).toContain(
      "!hasImportIdentities ||\n      (!peoplePreflightPending && !peoplePreflightError)",
    );
  });

  it("always adds safely matched members without a policy toggle", () => {
    expect(source).not.toContain("addMatchedMembers");
    expect(runImportSource).not.toContain("addMatchedMembers");
    expect(runImportSource).toContain(
      "const addedMembershipKeys = new Set<string>();",
    );
    expect(runImportSource).toContain("for (const use of identityUses)");
  });

  it("keeps review focused and gives both review groups a border", () => {
    expect(source).toContain("Review your import");
    expect(source).toContain("DO_NOT_IMPORT_VALUE");
    expect(source).toContain("<Select.Trigger");
    expect(source).toContain("{storyTermCapitalized} review");
    expect(source).toContain("<ImportPlanSummary");
    expect(source).toContain("<StrategicPillarImportReview");
    expect(source).toContain("<ObjectiveImportReview");
    expect(graphSource).toContain('label: "Strategic pillars"');
    expect(graphSource).toContain("draft.strategicPillars.length");
    expect(source).toContain("sm:flex-row sm:items-center");
    expect(source).not.toContain("Review before importing");
    expect(source).not.toContain("AnalysisWarnings");
    expect(source).not.toContain("How destination fields match");
  });

  it("keeps mapping notes light and limited to four visible lines", () => {
    expect(source).toContain(
      'className="text-foreground/90 font-medium dark:text-white/90"',
    );
    expect(source).toContain(
      'className="text-foreground/75 mt-2 max-h-24 overflow-y-auto pr-2 pl-5 dark:text-white/70"',
    );
    expect(source).not.toContain("max-h-64");
  });

  it("renders task review rows as a checkbox and editable title only", () => {
    expect(source).toMatch(/aria-label=\{`Import \$\{task\.title\}`\}/);
    expect(source).toMatch(/aria-label=\{`Title for \$\{task\.sourceId\}`\}/);
    expect(source).toContain("updateTaskTitle(taskIndex, event.target.value)");
    expect(source).toContain('placeholder="Add a title"');
    expect(source).toContain('"gap-2.5 px-4 py-2",');
    expect(source).toContain('className="min-w-0 flex-1"');
    expect(source).not.toContain("{task.description}");
    expect(source).not.toContain("<span>Source: {task.sourceId}</span>");
    expect(source).not.toContain("<span>Status: {task.status}</span>");
    expect(source).not.toContain("<span>Priority: {task.priority}</span>");
    expect(source).not.toContain("Assignee:");
    expect(source).not.toContain("getImportTaskRelationshipPreview");
    expect(source).not.toContain("relationshipPreview.map");
  });

  it("uses icon-free inverted primary actions", () => {
    expect(source).toContain('color="invert"');
    expect(source).not.toContain('color="gradient"');
    expect(source).not.toMatch(
      /rightIcon=\{<ArrowRight2Icon \/>}[^]*?Continue/,
    );
  });

  it("accepts JSON exports through the universal importer", () => {
    expect(source).toContain('"application/json": [".json"]');
    expect(source).toContain("CSV, JSON, Excel");
    expect(source).toContain('current?.sourceType === "csv"');
    expect(source).toContain("...completedAnalysis");
    expect(source).toContain("response.analysis.teams");
    expect(source).toContain("mergeAnalyzedTaskGraph");
    expect(source).toContain("prepareCompletedAIImportAnalysis");
    expect(source).toContain("setDraft(response.analysis);");
  });

  it("locks source identity after an attempt while preserving safe retries", () => {
    expect(source).toContain(
      "const [hasAttemptedImport, setHasAttemptedImport] = useState(false)",
    );
    expect(source).toContain("setHasAttemptedImport(true);");
    expect(source.match(/setHasAttemptedImport\(false\);/g)).toHaveLength(2);
    expect(source).toContain(
      "if (isImportMappingFieldLocked(field, hasAttemptedImport)) return;",
    );
    expect(source).toContain("disabled={mappingFieldLocked}");
    expect(source).toContain("Source ID is locked for safe retries");
    expect(source).toContain(
      "already created and complete only unfinished work.",
    );
    expect(source).toContain("Previously created {storyTermPlural} are");
    expect(source).toContain("{storyTerm}-field edits apply only to");
    expect(source).toContain("Import setup is locked for safe retries");
    expect(source).toContain(
      "The destination, team structure, and privacy stay fixed",
    );
    expect(source).toContain("Member mappings are locked for safe retries.");
    expect(source).toContain("disabled={hasAttemptedImport}");
    expect(source).toContain(
      "Import already attempted; retries reuse completed work",
    );
  });

  it("seeds an empty workspace fallback from immediate and polled source teams", () => {
    expect(source).toContain("getImportSourceTeamDestination(sourceTeam)");
    expect(source).toContain('setStructureMode("preserve")');
    expect(source).toContain("response.analysis.teams.at(0)");
    expect(source).toContain("teams.length === 0");
  });

  it("matches preserved source teams against the whole workspace", () => {
    expect(source).toContain("useJoinedTeams, useTeams");
    expect(source).toContain("knownWorkspaceTeams");
    expect(source).toContain("existingTeams={knownWorkspaceTeams}");
    expect(source).toContain("existingTeams: knownWorkspaceTeams");
    expect(source).toContain("actorUserId: session.user.id");
    expect(source).toContain("const joinedTeamIds = useMemo");
    expect(source).toContain("joinedTeamIds,");
    expect(source).toContain(
      "Reuse or create source teams, joining matched teams when needed.",
    );
    expect(source).toContain("Checking workspace teams for safe reuse");
    expect(source).toContain("workspaceTeamsError");
    expect(graphSource).toContain("resolveImportSourceTeam");
    expect(graphSource).toContain('destinationLabel = "Needs review"');
    expect(graphSource).toContain("Reuse + join");
  });

  it("scopes a created fallback team to the current file and destination", () => {
    expect(source).toContain("const createdFallbackTeam = useRef");
    expect(source).toContain("getNewTeamImportSignature");
    expect(source).toContain("JSON.stringify([");
    expect(source).toContain("fileHash,");
    expect(source).toContain("createdFallbackTeam.current?.signature");
    expect(source).toContain("createdFallbackTeam.current = null");
  });

  it("refreshes every possibly mutated domain after partial failures", () => {
    expect(source).toContain("void Promise.allSettled([");
    expect(source).toContain("storyKeys.all(workspaceSlug)");
    expect(source).toContain("objectiveKeys.all(workspaceSlug)");
    expect(source).toContain("memberKeys.all(workspaceSlug)");
    expect(source).toContain("strategyKeys.map(workspaceSlug)");
  });

  it("reviews pillars and keeps strategy-only imports team-free", () => {
    expect(source).toContain("selectedStrategicPillarSourceIds");
    expect(source).toContain("strategicPillarDestinationMatches");
    expect(source).toContain("hasSelectedTeamScopedImport");
    expect(source).toContain("No team changes");
    expect(source).toContain(
      'withWorkspace(viewOutcomeInStrategy ? "/strategy"',
    );
    expect(graphSource).toContain("Strategic pillar review");
    expect(graphSource).toContain("Create strategic pillar");
  });

  it("supports explicit member mapping and prevents privacy widening", () => {
    expect(source).toContain("confirmedMemberIdsByIdentityKey");
    expect(source).toContain("suggestImportPersonMember");
    expect(source).toContain("<ImportMemberPicker");
    expect(source).toContain(": preview.suggestedMember?.id");
    expect(source).toContain(
      "lockedMemberIdsByIdentityKey ?? confirmedMemberIdsByIdentityKey",
    );
    expect(source).toContain(
      "setLockedMemberIdsByIdentityKey(memberIdsByIdentityKey)",
    );
    expect(source).not.toContain("No safe workspace match");
    expect(source).not.toContain("Source details conflict; choose one");
    expect(memberPickerSource).toContain("Choose member");
    expect(memberPickerSource).toContain("Unassigned");
    expect(memberPickerSource).toContain("<Avatar");
    expect(memberPickerSource).toContain(
      'className="justify-between opacity-70"',
    );
    expect(memberPickerSource).not.toContain("<AiIcon");
    expect(memberPickerSource).not.toContain("AI suggested match");
    expect(memberPickerSource).not.toContain("Suggested");
    expect(memberPickerSource).not.toMatch(
      /<Text[^>]*>\s*\{member\.email\}\s*<\/Text>/,
    );
    expect(memberPickerSource).toContain(
      'className="z-[60] w-80 max-w-[calc(100vw-2rem)]"',
    );
    expect(memberPickerSource).toContain('className="max-w-full"');
    expect(source).toContain("Private source work needs a private destination");
    expect(source).toContain("!hasPrivacyWideningRisk");
  });

  it("uses the workspace terminology throughout import review and results", () => {
    expect(source).toContain("useTerminology, useWorkspacePath");
    expect(source).toContain('getTermDisplay("storyTerm")');
    expect(source).toContain("{storyTermCapitalized} review");
    expect(source).toContain("stable {storyTerm} source IDs");
    expect(graphSource).toContain("useTerminology");
    expect(graphSource).toContain('getTermDisplay("storyTerm"');
    expect(graphSource).toContain('getTermDisplay("objectiveTerm"');
    expect(graphSource).toContain('getTermDisplay("sprintTerm"');
    expect(graphSource).toContain('getTermDisplay("keyResultTerm"');
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
