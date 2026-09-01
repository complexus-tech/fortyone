"use client";

import type { ReactNode } from "react";
import { useEffect, useMemo, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useDropzone } from "react-dropzone";
import {
  ArrowDown2Icon,
  ArrowLeft2Icon,
  ArrowRight2Icon,
  CheckIcon,
  DownloadIcon,
  PlusIcon,
  TeamIcon,
  WarningIcon,
} from "icons";
import { cn } from "lib";
import { toast } from "sonner";
import {
  Box,
  Button,
  Checkbox,
  ColorPicker,
  Command,
  Dialog,
  DropZone,
  Flex,
  Input,
  Popover,
  ProgressBar,
  Select,
  Text,
} from "ui";
import { useSession } from "@/lib/auth/client";
import { useTerminology, useWorkspacePath } from "@/hooks";
import { TeamColor } from "@/components/ui/team-color";
import { Thinking } from "@/components/ui/chat/thinking";
import { storyKeys } from "@/modules/stories/constants";
import {
  labelKeys,
  memberKeys,
  sprintKeys,
  statusKeys,
  teamKeys,
} from "@/constants/keys";
import { objectiveKeys } from "@/shared/objectives/keys";
import { strategyKeys } from "@/shared/strategy-map/hooks";
import type { StrategyMap } from "@/shared/strategy-map/types";
import type { Objective } from "@/modules/objectives/types";
import { useJoinedTeams, useTeams } from "@/modules/teams/hooks/teams";
import type { Team } from "@/modules/teams/types";
import type { Member } from "@/types";
import type { ImportDraft, ImportMapping } from "../schema";
import { IMPORT_MAX_FILE_BYTES, importDestinationSchema } from "../schema";
import { mapRowsToImportTasks, sanitizeAIImportMapping } from "../csv";
import {
  getImportParentCycleSourceIds,
  isValidImportDateRange,
  resolveImportEntityNameMatch,
  resolveImportPerson,
  suggestImportPersonMember,
} from "../execution";
import {
  createImportTeam,
  getImportTeamObjectives,
  getImportStrategyMap,
  getImportWorkspaceMembers,
  pollImportAnalysis,
  startImportAnalysis,
} from "../api";
import type { ImportRunResult, ImportStructureMode } from "../run-import";
import {
  getCanonicalImportAssociation,
  getImportAssociationKey,
  getImportSourceTeamDestination,
  resolveImportSourceTeam,
  runImport,
} from "../run-import";
import { mergeAnalyzedTaskGraph } from "../task-graph";
import {
  ImportPlanSummary,
  ObjectiveImportReview,
  type ObjectiveImportDestinationPreview,
  SourceTeamPreview,
  StrategicPillarImportReview,
  type StrategicPillarImportDestinationPreview,
} from "./import-graph-review";
import {
  isImportMappingFieldLocked,
  prepareCompletedAIImportAnalysis,
} from "./import-wizard-safety";
import { collectImportIdentities } from "./import-identity-review";
import { ImportMemberPicker } from "./import-member-picker";

type ImportWizardProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

type DestinationChoice =
  | { kind: "existing"; teamId: string }
  | {
      kind: "new";
      name: string;
      code: string;
      color: string;
      isPrivate: boolean;
    };

type ObjectiveTargetPlan = {
  teamConflict: boolean;
  teamId: string | null;
  teamKey: string;
  teamLabel: string;
};

const WIZARD_STEP = {
  upload: 1,
  teams: 2,
  members: 3,
  review: 4,
  import: 5,
} as const;
const STEPS = ["Upload", "Teams", "Members", "Review", "Import"] as const;
const REVIEW_PAGE_SIZE = 50;
const IMPORT_ANALYSIS_POLL_TIMEOUT_MS = 7 * 60 * 1000;
const IMPORT_ANALYSIS_INITIAL_POLL_DELAY_MS = 700;
const IMPORT_ANALYSIS_MIN_POLL_DELAY_MS = 1_500;
const IMPORT_ANALYSIS_MAX_POLL_DELAY_MS = 10_000;
const IMPORT_ANALYSIS_POLL_ERROR_RETRIES = 3;
const DO_NOT_IMPORT_VALUE = "__do_not_import__";
const TRELLO_SOURCE_NAMESPACE_PREFIX = "trello:board:";
const EMPTY_IMPORT_SOURCE_IDS = new Set<string>();
const EMPTY_IMPORT_STRATEGY_MAP: StrategyMap = {
  description: null,
  pillars: [],
  ultimateGoal: "",
};

const normalizeImportReviewName = (value: string) =>
  value.normalize("NFKC").trim().toLocaleLowerCase().replace(/\s+/g, " ");

const normalizeImportColumnName = (value: string) =>
  value.trim().toLocaleLowerCase().replace(/[_-]+/g, " ").replace(/\s+/g, " ");

const getImportAnalysisPollDelay = (
  pollAttempts: number,
  consecutiveErrors: number,
) => {
  const regularDelay = Math.min(
    IMPORT_ANALYSIS_MAX_POLL_DELAY_MS,
    IMPORT_ANALYSIS_MIN_POLL_DELAY_MS * 1.35 ** Math.min(pollAttempts, 8),
  );
  const errorDelay = IMPORT_ANALYSIS_MIN_POLL_DELAY_MS * 2 ** consecutiveErrors;
  return Math.min(
    IMPORT_ANALYSIS_MAX_POLL_DELAY_MS,
    Math.max(regularDelay, errorDelay),
  );
};

const isDeterministicTrelloDraft = (draft: ImportDraft | null) =>
  Boolean(
    draft?.sourceType === "json" &&
      (draft.sourceMetadata?.platform === "trello" ||
        draft.sourceNamespace?.startsWith(TRELLO_SOURCE_NAMESPACE_PREFIX)),
  );

const getTrelloArchivedTaskSourceIds = (draft: ImportDraft | null) => {
  if (!isDeterministicTrelloDraft(draft) || !draft) return new Set<string>();
  if (draft.sourceMetadata?.platform === "trello") {
    return new Set(
      draft.sourceMetadata.archivedTaskSourceIds.flatMap((sourceId) => {
        const normalizedSourceId = sourceId.trim();
        return normalizedSourceId ? [normalizedSourceId] : [];
      }),
    );
  }
  const sourceIdColumn =
    draft.mapping?.sourceId ??
    draft.columns.find((column) => normalizeImportColumnName(column) === "id");
  const closedColumn = draft.columns.find(
    (column) => normalizeImportColumnName(column) === "closed",
  );
  if (!sourceIdColumn || !closedColumn) return new Set<string>();

  return new Set(
    draft.rows.flatMap((row) => {
      const sourceId = (row[sourceIdColumn] ?? "").trim();
      const closed =
        (row[closedColumn] ?? "").trim().toLocaleLowerCase() === "true";
      return sourceId && closed ? [sourceId] : [];
    }),
  );
};

const getTaskIndexesBySourceId = (
  draft: ImportDraft | null,
  sourceIds: ReadonlySet<string>,
) =>
  new Set(
    draft?.tasks.flatMap((task, index) =>
      sourceIds.has(task.sourceId) ? [index] : [],
    ) ?? [],
  );

const mergeDeterministicImportEntities = <T extends { sourceId: string }>(
  deterministicEntities: readonly T[],
  analyzedEntities: readonly T[],
) => {
  const analyzedBySourceId = new Map(
    analyzedEntities.map((entity) => [entity.sourceId, entity]),
  );
  const merged = deterministicEntities.map((entity) => {
    const analyzedEntity = analyzedBySourceId.get(entity.sourceId);
    return analyzedEntity ? { ...analyzedEntity, ...entity } : entity;
  });
  const sourceIds = new Set(
    deterministicEntities.map((entity) => entity.sourceId),
  );
  for (const entity of analyzedEntities) {
    if (sourceIds.has(entity.sourceId)) continue;
    sourceIds.add(entity.sourceId);
    merged.push(entity);
  }
  return merged;
};

const getNewTeamImportSignature = (
  fileHash: string,
  destination: Extract<DestinationChoice, { kind: "new" }>,
) =>
  JSON.stringify([
    fileHash,
    destination.name.trim(),
    destination.code.trim().toUpperCase(),
    destination.color,
    destination.isPrivate,
  ]);

const IMPORT_ANALYSIS_PHASES = [
  {
    activeLabel: "Uploading",
    completeLabel: "Uploaded",
    detail: "Securely uploading your export.",
    label: "Upload",
    title: "Uploading your file",
  },
  {
    activeLabel: "Reading",
    completeLabel: "Read",
    detail: "Finding rows, headers, and source fields.",
    label: "Read",
    title: "Reading your export",
  },
  {
    activeLabel: "Mapping",
    completeLabel: "Mapped",
    detail: "Matching source objects and relationships to FortyOne work.",
    label: "Map",
    title: "Mapping your work",
  },
  {
    activeLabel: "Preparing",
    completeLabel: "Prepared",
    detail: "Your mapping is ready for team setup.",
    label: "Prepare",
    title: "Ready to choose teams",
  },
] as const;

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

const fileAccept = {
  "text/csv": [".csv"],
  "text/tab-separated-values": [".tsv"],
  "application/json": [".json"],
  "application/vnd.ms-excel": [".xls"],
  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": [
    ".xlsx",
  ],
  "application/pdf": [".pdf"],
  "image/jpeg": [".jpg", ".jpeg"],
  "image/png": [".png"],
  "image/webp": [".webp"],
};

const initialNewTeam = {
  kind: "new" as const,
  name: "",
  code: "",
  color: "#4A90E2",
  isPrivate: false,
};

const formatTeamCode = (value: string) =>
  value
    .toUpperCase()
    .replace(/[^A-Z0-9]/g, "")
    .slice(0, 3);

const fileNameToTeamName = (fileName: string) =>
  fileName
    .replace(/\.[^.]+$/, "")
    .replace(/[-_]+/g, " ")
    .replace(/\s+/g, " ")
    .trim()
    .slice(0, 24);

const getSuggestedTeamName = (
  sourceType: ImportDraft["sourceType"] | undefined,
  fileName: string,
) => {
  if (sourceType === "jira_csv") return "Jira Import";
  return fileNameToTeamName(fileName);
};

const SelectionCard = ({
  description,
  disabled = false,
  icon,
  label,
  onClick,
  selected,
}: {
  description: string;
  disabled?: boolean;
  icon: ReactNode;
  label: string;
  onClick: () => void;
  selected: boolean;
}) => (
  <button
    aria-pressed={selected}
    className={cn(
      "border-border bg-surface hover:border-border-strong focus-visible:ring-ring relative min-h-20 rounded-xl border-[0.5px] p-3 text-left transition-[border-color,background-color,box-shadow] focus-visible:ring-2 focus-visible:outline-none",
      selected && "border-primary bg-primary/5 ring-primary/15 ring-1",
      disabled && "cursor-not-allowed opacity-60",
    )}
    disabled={disabled}
    onClick={onClick}
    type="button"
  >
    <Flex align="start" gap={3}>
      <Box
        className={cn(
          "bg-surface-muted text-text-muted flex h-9 w-9 shrink-0 items-center justify-center rounded-lg transition-colors",
          selected && "bg-primary/10 text-primary",
        )}
      >
        {icon}
      </Box>
      <Box className="min-w-0">
        <Text className="font-medium">{label}</Text>
        <Text className="mt-1 leading-5" color="muted">
          {description}
        </Text>
      </Box>
    </Flex>
  </button>
);

const ImportAnalysisBanner = ({
  analysisError,
  analysisNotice,
  analysisPending,
  fileName,
  hasAttemptedImport,
  onReplace,
  uploadPending,
}: {
  analysisError: string;
  analysisNotice: string;
  analysisPending: boolean;
  fileName: string;
  hasAttemptedImport: boolean;
  onReplace: () => void;
  uploadPending: boolean;
}) => {
  const [analysisPhase, setAnalysisPhase] = useState(0);
  const busy = uploadPending || analysisPending;

  useEffect(() => {
    if (uploadPending) {
      setAnalysisPhase(0);
      return;
    }
    if (!analysisPending) return;

    setAnalysisPhase(1);
    const mapTimer = window.setTimeout(() => {
      setAnalysisPhase(2);
    }, 1_200);
    const prepareTimer = window.setTimeout(() => {
      setAnalysisPhase(3);
    }, 3_200);

    return () => {
      window.clearTimeout(mapTimer);
      window.clearTimeout(prepareTimer);
    };
  }, [analysisPending, fileName, uploadPending]);

  let visiblePhase = 3;
  if (analysisError) visiblePhase = -1;
  else if (uploadPending) visiblePhase = 0;
  else if (analysisPending) visiblePhase = Math.max(1, analysisPhase);

  let phase: { detail: string; label: string; title: string };
  if (analysisNotice) {
    phase = {
      detail:
        "AI suggestions were unavailable, so FortyOne kept the safe file mapping for your review.",
      label: "Prepare",
      title: "Ready with standard mapping",
    };
  } else if (visiblePhase >= 0) {
    phase = IMPORT_ANALYSIS_PHASES[visiblePhase] ?? IMPORT_ANALYSIS_PHASES[3];
  } else {
    phase = {
      detail: "Choose Replace to try a different export.",
      label: "Upload",
      title: "This file needs attention",
    };
  }

  return (
    <Box
      aria-busy={busy}
      className={cn(
        "border-border bg-surface relative mt-5 overflow-hidden rounded-xl border-[0.5px] p-5",
        busy && "border-transparent",
        analysisError && "border-danger/30 bg-danger/5",
      )}
    >
      {busy ? (
        <Box
          aria-hidden="true"
          className="from-secondary/10 via-info/10 to-primary/10 pointer-events-none absolute inset-0 animate-pulse bg-linear-to-r motion-reduce:animate-none"
        />
      ) : null}
      <Box className="relative z-10">
        <Flex
          className="flex-col items-stretch sm:flex-row sm:items-start"
          gap={4}
          justify="between"
        >
          <Box
            aria-live={analysisError ? "assertive" : "polite"}
            className="min-w-0 flex-1"
            role={analysisError ? "alert" : "status"}
          >
            <Text
              className={cn(
                "text-[0.8125rem] font-semibold tracking-[0.08em] uppercase",
                analysisError
                  ? "text-danger"
                  : "from-secondary via-info to-primary bg-linear-to-r bg-clip-text text-transparent",
              )}
            >
              FortyOne analysis
            </Text>
            <Text className="mt-1 font-medium">{phase.title}</Text>
            <Text className="mt-0.5 leading-6" color="muted">
              {analysisError || phase.detail}
            </Text>
          </Box>
          <Button
            className="shrink-0 self-start"
            color="tertiary"
            disabled={busy}
            onClick={onReplace}
            size="sm"
            variant="outline"
          >
            Replace
          </Button>
        </Flex>

        {analysisError ? null : (
          <Box aria-hidden="true" className="mt-4 grid grid-cols-4 gap-2">
            {IMPORT_ANALYSIS_PHASES.map((item, index) => {
              const active = busy && index === visiblePhase;
              const complete =
                visiblePhase >= 0 &&
                (busy ? index < visiblePhase : index <= visiblePhase);
              let phaseStatus: ReactNode = (
                <Text className="text-base font-medium" color="muted">
                  {item.label}
                </Text>
              );
              if (complete) {
                phaseStatus = (
                  <Flex align="center" className="text-foreground" gap={1}>
                    <CheckIcon className="h-3.5 w-auto" strokeWidth={2.4} />
                    <Text className="text-base font-medium">
                      {item.completeLabel}
                    </Text>
                  </Flex>
                );
              }
              if (active) {
                phaseStatus = (
                  <Thinking
                    className="min-h-6 gap-1 text-base font-medium"
                    message={item.activeLabel}
                  />
                );
              }
              return (
                <Box key={item.label}>
                  <Box className="bg-border h-1 overflow-hidden rounded-full">
                    <Box
                      className={cn(
                        "from-secondary via-info to-primary h-full origin-left scale-x-0 rounded-full bg-linear-to-r transition-transform duration-200 motion-reduce:transition-none",
                        (complete || active) && "scale-x-100",
                        active && "animate-pulse motion-reduce:animate-none",
                      )}
                    />
                  </Box>
                  <Box className="mt-1.5 min-h-6">{phaseStatus}</Box>
                </Box>
              );
            })}
          </Box>
        )}
        <Text className="mt-3 truncate" color="muted">
          {fileName} ·{" "}
          {hasAttemptedImport
            ? "Import already attempted; retries reuse completed work"
            : "Nothing has been imported yet"}
        </Text>
      </Box>
    </Box>
  );
};

const DestinationTeamPicker = ({
  disabled = false,
  onChange,
  teams,
  value,
}: {
  disabled?: boolean;
  onChange: (teamId: string) => void;
  teams: Team[];
  value: string;
}) => {
  const [open, setOpen] = useState(false);
  const selectedTeam = teams.find((team) => team.id === value);

  if (teams.length === 0) {
    return (
      <Text className="mt-3" color="muted">
        No existing teams. Choose “Create a new team” to continue.
      </Text>
    );
  }

  return (
    <Popover onOpenChange={setOpen} open={open}>
      <Popover.Trigger asChild>
        <Button
          align="between"
          className="mt-3 max-w-full"
          color="tertiary"
          disabled={disabled}
          rightIcon={<ArrowDown2Icon className="h-4 shrink-0" />}
          size="sm"
          variant="outline"
        >
          <Flex align="center" className="min-w-0" gap={2}>
            <TeamColor className="shrink-0" color={selectedTeam?.color} />
            <Text className="truncate">
              {selectedTeam?.name ?? "Choose a team"}
            </Text>
            {selectedTeam ? (
              <Text className="shrink-0" color="muted">
                {selectedTeam.code}
              </Text>
            ) : null}
          </Flex>
        </Button>
      </Popover.Trigger>
      <Popover.Content
        align="start"
        className="z-[60] w-80 max-w-[calc(100vw-2rem)]"
      >
        <Command label="Destination teams">
          <Command.Input autoFocus placeholder="Search teams…" />
          <Command.Separator />
          <Command.Empty className="py-3 text-base">
            <Text color="muted">No teams found.</Text>
          </Command.Empty>
          <Command.Group className="max-h-60 overflow-y-auto">
            {teams.map((team) => (
              <Command.Item
                active={value === team.id}
                className="justify-between"
                key={team.id}
                onSelect={() => {
                  onChange(team.id);
                  setOpen(false);
                }}
              >
                <Flex align="center" className="min-w-0" gap={2}>
                  <TeamColor className="shrink-0" color={team.color} />
                  <Text className="truncate">{team.name}</Text>
                </Flex>
                <Flex align="center" className="shrink-0" gap={2}>
                  <Text color="muted">{team.code}</Text>
                  {value === team.id ? (
                    <CheckIcon className="h-4 w-auto" strokeWidth={2.2} />
                  ) : null}
                </Flex>
              </Command.Item>
            ))}
          </Command.Group>
        </Command>
      </Popover.Content>
    </Popover>
  );
};

const WizardProgress = ({ step }: { step: number }) => (
  <Box className="mt-5">
    <Flex align="center" justify="between">
      {STEPS.map((label, index) => {
        const stepNumber = index + 1;
        const active = stepNumber <= step;
        return (
          <Flex align="center" className="min-w-0" gap={2} key={label}>
            <Box
              className={cn(
                "bg-surface-muted text-text-muted flex h-6 w-6 shrink-0 items-center justify-center rounded-full font-medium",
                active && "bg-foreground text-background",
              )}
            >
              {stepNumber < step ? <CheckIcon className="h-3.5" /> : stepNumber}
            </Box>
            <Text
              className={cn(
                "hidden font-medium md:block",
                !active && "text-text-muted",
              )}
            >
              {label}
            </Text>
          </Flex>
        );
      })}
    </Flex>
    <Box className="bg-border mt-3 h-1.5 overflow-hidden rounded-full">
      <Box
        className="bg-foreground h-full rounded-full transition-[width] duration-300"
        style={{ width: `${(step / STEPS.length) * 100}%` }}
      />
    </Box>
  </Box>
);

const formatEntityCount = (
  count: number,
  singular: string,
  plural = `${singular}s`,
) => `${count} ${count === 1 ? singular : plural}`;

const ImportRunStep = ({
  importPending,
  importProgress,
  outcome,
  runError,
}: {
  importPending: boolean;
  importProgress: number;
  outcome: ImportRunResult | null;
  runError: string;
}) => {
  const { getTermDisplay } = useTerminology();
  const storyTerm = getTermDisplay("storyTerm");
  const storyTermPlural = getTermDisplay("storyTerm", { variant: "plural" });
  const storyTermPluralCapitalized = getTermDisplay("storyTerm", {
    capitalize: true,
    variant: "plural",
  });
  const objectiveTerm = getTermDisplay("objectiveTerm");
  const objectiveTermPlural = getTermDisplay("objectiveTerm", {
    variant: "plural",
  });
  const keyResultTerm = getTermDisplay("keyResultTerm");
  const keyResultTermPlural = getTermDisplay("keyResultTerm", {
    variant: "plural",
  });
  const sprintTerm = getTermDisplay("sprintTerm");
  const sprintTermPlural = getTermDisplay("sprintTerm", { variant: "plural" });

  if (importPending) {
    const preparing = importProgress === 0;
    return (
      <Box
        aria-busy="true"
        aria-live="polite"
        className="mx-auto max-w-xl text-center"
        role="status"
      >
        <Text as="h2" className="text-2xl font-medium">
          {preparing ? "Preparing your import" : "Importing your work"}
        </Text>
        <Text className="mt-2 leading-6" color="muted">
          {preparing
            ? "FortyOne is checking teams, workflows, members, and relationships before creating work."
            : "FortyOne is creating the reviewed structure in dependency order. Keep this window open until the result is ready."}
        </Text>
        <ProgressBar className="mt-6 h-2" progress={importProgress} />
        <Text className="mt-2 font-medium">
          {preparing ? "Preparing…" : `${importProgress}% complete`}
        </Text>
      </Box>
    );
  }

  if (outcome) {
    const createdStructure =
      outcome.createdTeams +
      outcome.createdStrategicPillars +
      outcome.createdObjectives +
      outcome.createdKeyResults +
      outcome.createdSprints +
      outcome.createdLabels +
      outcome.createdLinks +
      outcome.addedMemberships +
      outcome.appliedCollaborators +
      outcome.createdAssociations +
      outcome.alignedObjectives;
    const successful = outcome.created + outcome.replayed + createdStructure;
    const hasIssues =
      outcome.failed > 0 ||
      outcome.destinationConflicts > 0 ||
      outcome.unresolvedAssociations > 0 ||
      outcome.unresolvedLinks > 0 ||
      outcome.unresolvedPeople > 0;
    const allFailed = successful === 0 && hasIssues;
    const partial = successful > 0 && hasIssues;
    let outcomeTitle = "Your import is ready";
    if (allFailed) outcomeTitle = "Nothing was imported";
    else if (partial) outcomeTitle = "Import finished with issues";
    let outcomeLead = "Applied the reviewed import";
    if (allFailed) outcomeLead = "Created no work";
    else if (outcome.created) {
      outcomeLead = `Created ${formatEntityCount(outcome.created, storyTerm, storyTermPlural)}`;
    }

    return (
      <Box aria-live="polite" className="mx-auto max-w-xl text-center">
        <Box
          className={cn(
            "mx-auto flex h-16 w-16 items-center justify-center rounded-3xl",
            hasIssues
              ? "bg-warning/10 text-warning"
              : "bg-success/10 text-success",
          )}
        >
          {hasIssues ? (
            <WarningIcon className="h-8" />
          ) : (
            <CheckIcon className="h-8" strokeWidth={2.5} />
          )}
        </Box>
        <Text as="h2" className="mt-5 text-2xl font-medium">
          {outcomeTitle}
        </Text>
        <Text className="mt-2 leading-6" color="muted">
          {outcomeLead}
          {outcome.replayed
            ? `, recognized ${formatEntityCount(outcome.replayed, `previously imported ${storyTerm}`, `previously imported ${storyTermPlural}`)}`
            : ""}
          {outcome.createdTeams
            ? `, ${formatEntityCount(outcome.createdTeams, "team")}`
            : ""}
          {outcome.createdStrategicPillars
            ? `, ${formatEntityCount(outcome.createdStrategicPillars, "strategic pillar")}`
            : ""}
          {outcome.createdObjectives
            ? `, ${formatEntityCount(outcome.createdObjectives, objectiveTerm, objectiveTermPlural)}`
            : ""}
          {outcome.createdKeyResults
            ? `, ${formatEntityCount(outcome.createdKeyResults, keyResultTerm, keyResultTermPlural)}`
            : ""}
          {outcome.createdSprints
            ? `, ${formatEntityCount(outcome.createdSprints, sprintTerm, sprintTermPlural)}`
            : ""}
          {outcome.createdLabels
            ? `, ${formatEntityCount(outcome.createdLabels, "label")}`
            : ""}
          {outcome.createdLinks
            ? `, ${formatEntityCount(outcome.createdLinks, `${storyTerm} link`)}`
            : ""}
          {outcome.addedMemberships
            ? `, ${formatEntityCount(outcome.addedMemberships, "team membership")}`
            : ""}
          {outcome.appliedCollaborators
            ? `, ${formatEntityCount(outcome.appliedCollaborators, "collaborator assignment")}`
            : ""}
          {outcome.createdAssociations
            ? `, ${formatEntityCount(outcome.createdAssociations, `${storyTerm} relationship`)}`
            : ""}
          {outcome.alignedObjectives
            ? `, ${formatEntityCount(outcome.alignedObjectives, `${objectiveTerm} alignment`)}`
            : ""}
          {outcome.failed
            ? `, and found ${formatEntityCount(outcome.failed, storyTerm, storyTermPlural)} that ${outcome.failed === 1 ? "needs" : "need"} attention`
            : ""}
          .
        </Text>
        {outcome.unresolvedPeople ? (
          <Text className="mt-2 leading-6" color="muted">
            {outcome.unresolvedPeople} people could not be matched safely and
            were left unassigned. No invitations were sent.
          </Text>
        ) : null}
        {outcome.destinationConflicts ? (
          <Text className="text-warning mt-2 leading-6">
            {outcome.destinationConflicts} source objects or relationships had
            ambiguous or incompatible destination matches and were left
            unchanged. Review the mapping notes or adjust the source names
            before retrying.
          </Text>
        ) : null}
        {outcome.unresolvedAssociations ? (
          <Text className="text-warning mt-2 leading-6">
            {outcome.unresolvedAssociations}{" "}
            {outcome.unresolvedAssociations === 1
              ? `${storyTerm} relationship could`
              : `${storyTerm} relationships could`}{" "}
            not be preserved because a source or destination {storyTerm} was
            unavailable, unselected, or in another team.
          </Text>
        ) : null}
        {outcome.unresolvedLinks ? (
          <Text className="text-warning mt-2 leading-6">
            {outcome.unresolvedLinks}{" "}
            {outcome.unresolvedLinks === 1
              ? `${storyTerm} link could`
              : `${storyTerm} links could`}{" "}
            not be preserved because the destination {storyTerm} or link service
            was unavailable. Existing links were not duplicated.
          </Text>
        ) : null}
        {outcome.failed ? (
          <Box className="bg-warning/8 mt-5 max-h-52 overflow-y-auto rounded-2xl p-4 text-left">
            <Text className="font-medium">
              {storyTermPluralCapitalized} needing attention
            </Text>
            <Box as="ul" className="text-text-muted mt-2 space-y-1">
              {outcome.items
                .filter((item) => item.error !== null)
                .map((item) => (
                  <li key={item.sourceKey}>
                    {item.sourceKey}: {item.error?.message}
                  </li>
                ))}
            </Box>
          </Box>
        ) : null}
      </Box>
    );
  }

  return (
    <Box aria-live="assertive" className="mx-auto max-w-xl text-center">
      <Box className="bg-danger/10 text-danger mx-auto flex h-16 w-16 items-center justify-center rounded-3xl">
        <WarningIcon className="h-8" />
      </Box>
      <Text as="h2" className="mt-5 text-2xl font-medium">
        The import paused
      </Text>
      <Text className="mt-2 leading-6" color="muted">
        {runError || "The import could not finish."} Retrying reuses exact
        structure matches and stable {storyTerm} source IDs.
      </Text>
    </Box>
  );
};

export const ImportWizard = ({ onOpenChange, open }: ImportWizardProps) => {
  const { data: session } = useSession();
  const { getTermDisplay } = useTerminology();
  const storyTerm = getTermDisplay("storyTerm");
  const storyTermPlural = getTermDisplay("storyTerm", { variant: "plural" });
  const storyTermCapitalized = getTermDisplay("storyTerm", {
    capitalize: true,
  });
  const objectiveTerm = getTermDisplay("objectiveTerm");
  const objectiveTermPlural = getTermDisplay("objectiveTerm", {
    variant: "plural",
  });
  const objectiveTermPluralCapitalized = getTermDisplay("objectiveTerm", {
    capitalize: true,
    variant: "plural",
  });
  const keyResultTerm = getTermDisplay("keyResultTerm");
  const keyResultTermPlural = getTermDisplay("keyResultTerm", {
    variant: "plural",
  });
  const sprintTerm = getTermDisplay("sprintTerm");
  const sprintTermPlural = getTermDisplay("sprintTerm", { variant: "plural" });
  const { data: teams = [] } = useJoinedTeams();
  const {
    data: workspaceTeams = [],
    isError: workspaceTeamsError,
    isPending: workspaceTeamsPending,
  } = useTeams();
  const { withWorkspace, workspaceSlug } = useWorkspacePath();
  const queryClient = useQueryClient();
  const mappingEdited = useRef(false);
  const mappingOverrideFields = useRef(new Set<keyof ImportMapping>());
  const draftRef = useRef<ImportDraft | null>(null);
  const analysisGeneration = useRef(0);
  const analysisPollingSession = useRef<{
    responseId: string;
    startedAt: number;
  } | null>(null);
  const createdSourceTeamIds = useRef(new Map<string, string>());
  const sourceObjectiveCache = useRef(
    new Map<string, { id: string; teamId: string }>(),
  );
  const createdFallbackTeam = useRef<{
    id: string;
    signature: string;
  } | null>(null);
  const [step, setStep] = useState<number>(WIZARD_STEP.upload);
  const [fileName, setFileName] = useState("");
  const [draft, setDraft] = useState<ImportDraft | null>(null);
  const [responseId, setResponseId] = useState<string | null>(null);
  const [fileHash, setFileHash] = useState("");
  const [analysisPending, setAnalysisPending] = useState(false);
  const [analysisError, setAnalysisError] = useState("");
  const [analysisNotice, setAnalysisNotice] = useState("");
  const [uploadPending, setUploadPending] = useState(false);
  const [destination, setDestination] =
    useState<DestinationChoice>(initialNewTeam);
  const [structureMode, setStructureMode] =
    useState<ImportStructureMode>("single");
  const [selectedMemberIdsByIdentityKey, setSelectedMemberIdsByIdentityKey] =
    useState<Map<string, string | null>>(new Map());
  const [lockedMemberIdsByIdentityKey, setLockedMemberIdsByIdentityKey] =
    useState<Map<string, string | null> | null>(null);
  const [excludedRows, setExcludedRows] = useState<Set<number>>(new Set());
  const [includeArchivedTrelloCards, setIncludeArchivedTrelloCards] =
    useState(false);
  const [excludedObjectives, setExcludedObjectives] = useState<Set<string>>(
    new Set(),
  );
  const [excludedStrategicPillars, setExcludedStrategicPillars] = useState<
    Set<string>
  >(new Set());
  const [objectivesByTeamId, setObjectivesByTeamId] = useState<
    Map<string, Objective[]>
  >(new Map());
  const [objectivePreflightPending, setObjectivePreflightPending] =
    useState(false);
  const [objectivePreflightError, setObjectivePreflightError] = useState("");
  const [objectivePreflightRevision, setObjectivePreflightRevision] =
    useState(0);
  const [workspaceMembersForReview, setWorkspaceMembersForReview] = useState<
    Member[]
  >([]);
  const [peoplePreflightPending, setPeoplePreflightPending] = useState(false);
  const [peoplePreflightError, setPeoplePreflightError] = useState("");
  const [peoplePreflightRevision, setPeoplePreflightRevision] = useState(0);
  const [strategyMapForReview, setStrategyMapForReview] = useState<StrategyMap>(
    EMPTY_IMPORT_STRATEGY_MAP,
  );
  const [strategyPreflightPending, setStrategyPreflightPending] =
    useState(false);
  const [strategyPreflightError, setStrategyPreflightError] = useState("");
  const [strategyPreflightRevision, setStrategyPreflightRevision] = useState(0);
  const [createdFallbackTeamForReview, setCreatedFallbackTeamForReview] =
    useState<{ id: string; signature: string } | null>(null);
  const [sourceTeamMappingsForReview, setSourceTeamMappingsForReview] =
    useState<Map<string, string>>(new Map());
  const [
    sourceObjectiveMappingsForReview,
    setSourceObjectiveMappingsForReview,
  ] = useState<Map<string, { id: string; teamId: string }>>(new Map());
  const [importPending, setImportPending] = useState(false);
  const [importProgress, setImportProgress] = useState(0);
  const [hasAttemptedImport, setHasAttemptedImport] = useState(false);
  const [outcome, setOutcome] = useState<ImportRunResult | null>(null);
  const [runError, setRunError] = useState("");
  const [reviewPage, setReviewPage] = useState(0);
  const importIdentities = useMemo(
    () => collectImportIdentities(draft),
    [draft],
  );
  const peopleMappingPreview = useMemo(
    () =>
      importIdentities.map((identity) => {
        const resolution = identity.hasConflictingClaims
          ? undefined
          : resolveImportPerson(identity, workspaceMembersForReview);
        const suggestion = identity.hasConflictingClaims
          ? undefined
          : suggestImportPersonMember(identity, workspaceMembersForReview);
        return {
          identity,
          identityKey: identity.identityKey,
          suggestedMember: resolution?.member ?? suggestion?.member,
        };
      }),
    [importIdentities, workspaceMembersForReview],
  );
  const confirmedMemberIdsByIdentityKey = useMemo(() => {
    const eligibleMemberIds = new Set(
      workspaceMembersForReview
        .filter((member) => member.isActive && !member.isSystem)
        .map((member) => member.id),
    );
    const selected = new Map<string, string | null>();
    for (const preview of peopleMappingPreview) {
      const memberId = selectedMemberIdsByIdentityKey.has(preview.identityKey)
        ? selectedMemberIdsByIdentityKey.get(preview.identityKey)
        : preview.suggestedMember?.id;
      selected.set(
        preview.identityKey,
        memberId && eligibleMemberIds.has(memberId) ? memberId : null,
      );
    }
    return selected;
  }, [
    peopleMappingPreview,
    selectedMemberIdsByIdentityKey,
    workspaceMembersForReview,
  ]);
  const hasImportIdentities = importIdentities.length > 0;
  const reviewedMemberIdsByIdentityKey =
    lockedMemberIdsByIdentityKey ?? confirmedMemberIdsByIdentityKey;
  const draftTeams = draft?.teams;
  const strategyPreflightRequired = Boolean(
    draft?.strategicPillars.length ||
      draft?.objectives.some((objective) => objective.pillarSourceId),
  );

  const knownWorkspaceTeams = workspaceTeamsPending ? teams : workspaceTeams;
  const joinedTeamIds = useMemo(
    () => new Set(teams.map((team) => team.id)),
    [teams],
  );
  const fallbackTargetPlan = useMemo<ObjectiveTargetPlan>(() => {
    const newTeamSignature =
      destination.kind === "new"
        ? getNewTeamImportSignature(fileHash, destination)
        : null;
    const createdFallbackTeamId =
      createdFallbackTeamForReview?.signature === newTeamSignature
        ? createdFallbackTeamForReview.id
        : null;
    const fallbackTeam =
      destination.kind === "existing"
        ? knownWorkspaceTeams.find((team) => team.id === destination.teamId)
        : undefined;
    const teamId =
      destination.kind === "existing"
        ? destination.teamId
        : createdFallbackTeamId;
    return {
      teamConflict: false,
      teamId,
      teamKey: teamId ? `existing:${teamId}` : "new:fallback",
      teamLabel:
        destination.kind === "existing"
          ? fallbackTeam?.name ?? "destination team"
          : destination.name.trim() || "new destination team",
    };
  }, [
    createdFallbackTeamForReview,
    destination,
    fileHash,
    knownWorkspaceTeams,
  ]);
  const sourceTeamTargetPlans = useMemo(() => {
    const plans = new Map<string, ObjectiveTargetPlan>();
    if (!draftTeams || structureMode !== "preserve") return plans;

    for (const sourceTeam of draftTeams) {
      const cachedTeamId = sourceTeamMappingsForReview.get(sourceTeam.sourceId);
      if (cachedTeamId) {
        plans.set(sourceTeam.sourceId, {
          teamConflict: false,
          teamId: cachedTeamId,
          teamKey: `existing:${cachedTeamId}`,
          teamLabel: sourceTeam.name,
        });
        continue;
      }
      const resolution = resolveImportSourceTeam(
        sourceTeam,
        knownWorkspaceTeams,
      );
      if (resolution.kind === "unique") {
        plans.set(sourceTeam.sourceId, {
          teamConflict: false,
          teamId: resolution.team.id,
          teamKey: `existing:${resolution.team.id}`,
          teamLabel: resolution.team.name,
        });
        continue;
      }
      const matchesNewFallback =
        resolution.kind === "none" &&
        destination.kind === "new" &&
        sourceTeam.isPrivate === destination.isPrivate &&
        normalizeImportReviewName(
          getImportSourceTeamDestination(sourceTeam).name,
        ) === normalizeImportReviewName(destination.name);
      let teamKey = `new:source:${sourceTeam.sourceId}`;
      if (resolution.kind === "ambiguous") {
        teamKey = `conflict:source:${sourceTeam.sourceId}`;
      } else if (matchesNewFallback) {
        teamKey = "new:fallback";
      }
      plans.set(sourceTeam.sourceId, {
        teamConflict: resolution.kind === "ambiguous",
        teamId: null,
        teamKey,
        teamLabel: matchesNewFallback
          ? destination.name.trim() || sourceTeam.name
          : sourceTeam.name,
      });
    }

    const destinationGroups = new Map<string, string[]>();
    for (const sourceTeam of draftTeams) {
      const plan = plans.get(sourceTeam.sourceId);
      if (!plan || plan.teamConflict) continue;
      let collisionKey = `new-name:${sourceTeam.isPrivate}:${normalizeImportReviewName(
        getImportSourceTeamDestination(sourceTeam).name,
      )}`;
      if (plan.teamId) collisionKey = `existing:${plan.teamId}`;
      else if (plan.teamKey === "new:fallback") collisionKey = plan.teamKey;
      const sourceIds = destinationGroups.get(collisionKey) ?? [];
      sourceIds.push(sourceTeam.sourceId);
      destinationGroups.set(collisionKey, sourceIds);
    }
    for (const [collisionKey, sourceIds] of destinationGroups) {
      if (sourceIds.length < 2) continue;
      for (const sourceId of sourceIds) {
        const plan = plans.get(sourceId);
        if (!plan) continue;
        plans.set(sourceId, {
          ...plan,
          teamConflict: true,
          teamKey: `conflict:collision:${collisionKey}`,
        });
      }
    }
    return plans;
  }, [
    destination,
    draftTeams,
    knownWorkspaceTeams,
    sourceTeamMappingsForReview,
    structureMode,
  ]);
  const objectiveTargetPlans = useMemo(() => {
    const plans = new Map<string, ObjectiveTargetPlan>();
    if (!draft) return plans;

    for (const objective of draft.objectives) {
      if (structureMode === "single" || !objective.teamSourceId) {
        plans.set(objective.sourceId, fallbackTargetPlan);
        continue;
      }
      const sourceTeamPlan = sourceTeamTargetPlans.get(objective.teamSourceId);
      plans.set(objective.sourceId, sourceTeamPlan ?? fallbackTargetPlan);
    }
    return plans;
  }, [draft, fallbackTargetPlan, sourceTeamTargetPlans, structureMode]);
  const hasSourceTeamConflict = [...sourceTeamTargetPlans.values()].some(
    (plan) => plan.teamConflict,
  );
  const conflictedSourceTeamIds = useMemo(
    () =>
      new Set(
        [...sourceTeamTargetPlans].flatMap(([sourceId, plan]) =>
          plan.teamConflict ? [sourceId] : [],
        ),
      ),
    [sourceTeamTargetPlans],
  );
  const objectiveTargetTeamIds = useMemo(() => {
    const teamIds = [...objectiveTargetPlans.values()].flatMap((plan) =>
      plan.teamId ? [plan.teamId] : [],
    );
    return [...new Set(teamIds)].sort();
  }, [objectiveTargetPlans]);
  const objectiveDestinationMatches = useMemo(() => {
    const matches = new Map<string, ObjectiveImportDestinationPreview>();
    if (!draft) return matches;

    const sourceNameCounts = new Map<string, number>();
    for (const objective of draft.objectives) {
      if (excludedObjectives.has(objective.sourceId)) continue;
      const target = objectiveTargetPlans.get(objective.sourceId);
      if (!target) continue;
      const key = `${target.teamKey}\0${normalizeImportReviewName(objective.name)}`;
      sourceNameCounts.set(key, (sourceNameCounts.get(key) ?? 0) + 1);
    }

    const withPillarReview = (
      objective: ImportDraft["objectives"][number],
      preview: ObjectiveImportDestinationPreview,
      destinationObjectiveId?: string,
    ): ObjectiveImportDestinationPreview => {
      if (!objective.pillarSourceId) return preview;
      const sourcePillar = draft.strategicPillars.find(
        (pillar) => pillar.sourceId === objective.pillarSourceId,
      );
      if (
        !sourcePillar ||
        excludedStrategicPillars.has(objective.pillarSourceId)
      ) {
        return {
          ...preview,
          pillarLabel: `Referenced pillar ${sourcePillar?.name ?? objective.pillarSourceId} is not selected; this ${objectiveTerm} will import without that alignment.`,
        };
      }
      if (!destinationObjectiveId) return preview;

      const currentPillars = strategyMapForReview.pillars.filter((pillar) =>
        pillar.objectiveIds.includes(destinationObjectiveId),
      );
      if (currentPillars.length === 0) return preview;
      const destinationPillar = resolveImportEntityNameMatch(
        sourcePillar.name,
        strategyMapForReview.pillars,
      );
      if (
        currentPillars.length === 1 &&
        destinationPillar.kind === "unique" &&
        currentPillars[0].id === destinationPillar.entity.id
      ) {
        return preview;
      }
      return {
        ...preview,
        kind: "pillar_conflict",
        pillarLabel: `The reused ${objectiveTerm} is already aligned to ${currentPillars.map((pillar) => pillar.name).join(", ")}; the source requests ${sourcePillar.name}. Resolve that alignment before importing.`,
      };
    };

    for (const objective of draft.objectives) {
      const target = objectiveTargetPlans.get(objective.sourceId);
      if (!target) continue;
      if (target.teamConflict) {
        matches.set(objective.sourceId, {
          kind: "team_conflict",
          teamLabel: target.teamLabel,
        });
        continue;
      }
      const sourceNameKey = `${target.teamKey}\0${normalizeImportReviewName(objective.name)}`;
      if (
        !excludedObjectives.has(objective.sourceId) &&
        (sourceNameCounts.get(sourceNameKey) ?? 0) > 1
      ) {
        matches.set(objective.sourceId, {
          kind: "source_conflict",
          teamLabel: target.teamLabel,
        });
        continue;
      }
      const cachedObjective = sourceObjectiveMappingsForReview.get(
        objective.sourceId,
      );
      if (target.teamId && cachedObjective?.teamId === target.teamId) {
        const existingObjective = (
          objectivesByTeamId.get(target.teamId) ?? []
        ).find((candidate) => candidate.id === cachedObjective.id);
        matches.set(
          objective.sourceId,
          withPillarReview(
            objective,
            {
              kind: "unique",
              locked: true,
              objectiveName: existingObjective?.name ?? objective.name,
              teamLabel: target.teamLabel,
            },
            cachedObjective.id,
          ),
        );
        continue;
      }
      if (!target.teamId) {
        matches.set(
          objective.sourceId,
          withPillarReview(objective, {
            kind: "none",
            teamLabel: target.teamLabel,
          }),
        );
        continue;
      }
      const exactNameCandidates = (
        objectivesByTeamId.get(target.teamId) ?? []
      ).filter(
        (candidate) =>
          normalizeImportReviewName(candidate.name) ===
          normalizeImportReviewName(objective.name),
      );
      const candidates = exactNameCandidates.filter(
        (candidate) => candidate.isPrivate === objective.isPrivate,
      );
      const resolution = resolveImportEntityNameMatch(
        objective.name,
        candidates,
      );
      if (resolution.kind === "unique") {
        matches.set(
          objective.sourceId,
          withPillarReview(
            objective,
            {
              kind: "unique",
              objectiveName: resolution.entity.name,
              teamLabel: target.teamLabel,
            },
            resolution.entity.id,
          ),
        );
      } else if (resolution.kind === "ambiguous") {
        matches.set(objective.sourceId, {
          kind: "ambiguous",
          matchCount: resolution.entities.length,
          teamLabel: target.teamLabel,
        });
      } else if (exactNameCandidates.length > 0) {
        matches.set(objective.sourceId, {
          kind: "privacy_conflict",
          teamLabel: target.teamLabel,
        });
      } else {
        matches.set(
          objective.sourceId,
          withPillarReview(objective, {
            kind: "none",
            teamLabel: target.teamLabel,
          }),
        );
      }
    }
    return matches;
  }, [
    draft,
    excludedObjectives,
    excludedStrategicPillars,
    objectiveTerm,
    objectiveTargetPlans,
    objectivesByTeamId,
    sourceObjectiveMappingsForReview,
    strategyMapForReview,
  ]);
  const strategicPillarDestinationMatches = useMemo(() => {
    const matches = new Map<string, StrategicPillarImportDestinationPreview>();
    if (!draft) return matches;
    const sourceNameCounts = new Map<string, number>();
    for (const pillar of draft.strategicPillars) {
      if (excludedStrategicPillars.has(pillar.sourceId)) continue;
      const name = normalizeImportReviewName(pillar.name);
      sourceNameCounts.set(name, (sourceNameCounts.get(name) ?? 0) + 1);
    }
    for (const pillar of draft.strategicPillars) {
      const normalizedName = normalizeImportReviewName(pillar.name);
      if (
        !excludedStrategicPillars.has(pillar.sourceId) &&
        (sourceNameCounts.get(normalizedName) ?? 0) > 1
      ) {
        matches.set(pillar.sourceId, { kind: "source_conflict" });
        continue;
      }
      const resolution = resolveImportEntityNameMatch(
        pillar.name,
        strategyMapForReview.pillars,
      );
      if (resolution.kind === "unique") {
        matches.set(pillar.sourceId, {
          kind: "unique",
          pillarName: resolution.entity.name,
        });
      } else if (resolution.kind === "ambiguous") {
        matches.set(pillar.sourceId, {
          kind: "ambiguous",
          matchCount: resolution.entities.length,
        });
      } else {
        matches.set(pillar.sourceId, { kind: "none" });
      }
    }
    return matches;
  }, [draft, excludedStrategicPillars, strategyMapForReview.pillars]);

  const reset = () => {
    analysisGeneration.current += 1;
    analysisPollingSession.current = null;
    mappingEdited.current = false;
    mappingOverrideFields.current.clear();
    createdSourceTeamIds.current.clear();
    sourceObjectiveCache.current.clear();
    createdFallbackTeam.current = null;
    setCreatedFallbackTeamForReview(null);
    setSourceTeamMappingsForReview(new Map());
    setSourceObjectiveMappingsForReview(new Map());
    setStep(WIZARD_STEP.upload);
    setFileName("");
    setDraft(null);
    setResponseId(null);
    setFileHash("");
    setAnalysisPending(false);
    setAnalysisError("");
    setAnalysisNotice("");
    setUploadPending(false);
    setDestination(
      teams[0] ? { kind: "existing", teamId: teams[0].id } : initialNewTeam,
    );
    setStructureMode("single");
    setSelectedMemberIdsByIdentityKey(new Map());
    setLockedMemberIdsByIdentityKey(null);
    setExcludedRows(new Set());
    setIncludeArchivedTrelloCards(false);
    setExcludedObjectives(new Set());
    setExcludedStrategicPillars(new Set());
    setObjectivesByTeamId(new Map());
    setObjectivePreflightPending(false);
    setObjectivePreflightError("");
    setObjectivePreflightRevision(0);
    setWorkspaceMembersForReview([]);
    setPeoplePreflightPending(false);
    setPeoplePreflightError("");
    setPeoplePreflightRevision(0);
    setStrategyMapForReview(EMPTY_IMPORT_STRATEGY_MAP);
    setStrategyPreflightPending(false);
    setStrategyPreflightError("");
    setStrategyPreflightRevision(0);
    setImportPending(false);
    setImportProgress(0);
    setHasAttemptedImport(false);
    setOutcome(null);
    setRunError("");
    setReviewPage(0);
  };

  useEffect(() => {
    draftRef.current = draft;
  }, [draft]);

  useEffect(() => {
    if (!open || destination.kind !== "new" || destination.name) return;
    if (teams[0]) setDestination({ kind: "existing", teamId: teams[0].id });
  }, [destination, open, teams]);

  useEffect(() => {
    let cancelled = false;
    if (!session || objectiveTargetTeamIds.length === 0) {
      setObjectivesByTeamId(new Map());
      setObjectivePreflightPending(false);
      setObjectivePreflightError("");
      return;
    }

    const loadObjectives = () => {
      setObjectivePreflightPending(true);
      setObjectivePreflightError("");
      void Promise.all(
        objectiveTargetTeamIds.map(async (teamId) => {
          const objectives = await getImportTeamObjectives(teamId, {
            session,
            workspaceSlug,
          });
          return [teamId, objectives] as const;
        }),
      ).then(
        (entries) => {
          if (cancelled) return;
          setObjectivesByTeamId(new Map(entries));
          setObjectivePreflightPending(false);
        },
        () => {
          if (cancelled) return;
          setObjectivePreflightError(
            `Existing ${objectiveTermPlural} could not be checked safely.`,
          );
          setObjectivePreflightPending(false);
        },
      );
    };

    loadObjectives();
    return () => {
      cancelled = true;
    };
  }, [
    objectivePreflightRevision,
    objectiveTermPlural,
    objectiveTargetTeamIds,
    session,
    workspaceSlug,
  ]);

  useEffect(() => {
    let cancelled = false;
    if (!session || !hasImportIdentities) {
      setWorkspaceMembersForReview([]);
      setPeoplePreflightPending(false);
      setPeoplePreflightError("");
      return;
    }

    const loadWorkspaceMembers = () => {
      setPeoplePreflightPending(true);
      setPeoplePreflightError("");
      void getImportWorkspaceMembers({
        session,
        workspaceSlug,
      }).then(
        (members) => {
          if (cancelled) return;
          setWorkspaceMembersForReview(members);
          setPeoplePreflightPending(false);
        },
        () => {
          if (cancelled) return;
          setPeoplePreflightError(
            "Workspace members could not be checked safely.",
          );
          setPeoplePreflightPending(false);
        },
      );
    };

    loadWorkspaceMembers();
    return () => {
      cancelled = true;
    };
  }, [
    fileHash,
    hasImportIdentities,
    peoplePreflightRevision,
    session,
    workspaceSlug,
  ]);

  useEffect(() => {
    let cancelled = false;
    if (!session || !strategyPreflightRequired) {
      setStrategyMapForReview(EMPTY_IMPORT_STRATEGY_MAP);
      setStrategyPreflightPending(false);
      setStrategyPreflightError("");
      return;
    }

    const loadStrategyMap = () => {
      setStrategyPreflightPending(true);
      setStrategyPreflightError("");
      void getImportStrategyMap({
        session,
        workspaceSlug,
      }).then(
        (strategyMap) => {
          if (cancelled) return;
          setStrategyMapForReview(strategyMap);
          setStrategyPreflightPending(false);
        },
        () => {
          if (cancelled) return;
          setStrategyPreflightError(
            "Existing strategic pillars could not be checked safely.",
          );
          setStrategyPreflightPending(false);
        },
      );
    };

    loadStrategyMap();
    return () => {
      cancelled = true;
    };
  }, [
    fileHash,
    session,
    strategyPreflightRequired,
    strategyPreflightRevision,
    workspaceSlug,
  ]);

  // react-doctor-disable-next-line react-doctor/effect-needs-cleanup -- Cleanup clears the active timer, and the cancellation guard prevents an in-flight poll from scheduling another one after unmount.
  useEffect(() => {
    if (!responseId || !fileHash || !analysisPending) return;
    const generation = analysisGeneration.current;
    const pollingSession =
      analysisPollingSession.current?.responseId === responseId
        ? analysisPollingSession.current
        : { responseId, startedAt: Date.now() };
    analysisPollingSession.current = pollingSession;
    const deadline = pollingSession.startedAt + IMPORT_ANALYSIS_POLL_TIMEOUT_MS;
    let cancelled = false;
    let pollAttempts = 0;
    let consecutivePollErrors = 0;
    let pollTimer: ReturnType<typeof setTimeout> | undefined;

    const isCurrentPollingSession = () =>
      !cancelled &&
      generation === analysisGeneration.current &&
      analysisPollingSession.current?.responseId === responseId;
    const clearPollingSession = () => {
      if (analysisPollingSession.current?.responseId === responseId) {
        analysisPollingSession.current = null;
      }
      clearTimeout(deadlineTimer);
    };

    const handlePollFailure = (error: unknown) => {
      if (!isCurrentPollingSession()) return;
      const message =
        error instanceof Error ? error.message : "AI analysis failed";
      clearPollingSession();
      setAnalysisPending(false);
      setResponseId(null);
      setDraft((current) =>
        current
          ? { ...current, warnings: [...current.warnings, message] }
          : current,
      );
      if (draftRef.current) setAnalysisNotice(message);
      else setAnalysisError(message);
    };
    const handlePollTimeout = () => {
      handlePollFailure(
        new Error(
          "AI analysis is taking longer than expected. You can continue with the deterministic preview or upload the file again.",
        ),
      );
    };
    const schedulePoll = (delay: number) => {
      if (!isCurrentPollingSession()) return;
      const remainingTime = deadline - Date.now();
      if (remainingTime <= 0) {
        handlePollTimeout();
        return;
      }
      pollTimer = setTimeout(poll, Math.min(delay, remainingTime));
    };
    const poll = () => {
      if (!isCurrentPollingSession()) return;
      if (Date.now() >= deadline) {
        handlePollTimeout();
        return;
      }
      pollAttempts += 1;
      void pollImportAnalysis({
        fileHash,
        responseId,
        workspaceSlug,
      }).then(
        (response) => {
          if (!isCurrentPollingSession()) return;
          consecutivePollErrors = 0;
          if (response.status !== "completed") {
            schedulePoll(getImportAnalysisPollDelay(pollAttempts, 0));
            return;
          }
          const completedAnalysis = prepareCompletedAIImportAnalysis(
            response.analysis,
          );

          clearPollingSession();
          setDraft((current) => {
            const usesDeterministicRowMapping =
              current?.sourceType === "csv" ||
              current?.sourceType === "jira_csv";
            const preservesDeterministicTrelloGraph =
              isDeterministicTrelloDraft(current);
            const preservesDeterministicTaskSet =
              usesDeterministicRowMapping || preservesDeterministicTrelloGraph;
            const canMergeDeterministicAnalysis = Boolean(
              current &&
                preservesDeterministicTaskSet &&
                (current.rows.length > 0 || preservesDeterministicTrelloGraph),
            );
            if (current && canMergeDeterministicAnalysis) {
              let mapping = completedAnalysis.mapping;
              if (usesDeterministicRowMapping) {
                mapping =
                  !mappingEdited.current && completedAnalysis.mapping
                    ? sanitizeAIImportMapping(
                        completedAnalysis.mapping,
                        current.columns,
                      )
                    : current.mapping;
              }
              const mappedTasks =
                usesDeterministicRowMapping && mapping
                  ? mapRowsToImportTasks(current.rows, mapping)
                  : current.tasks;
              return {
                ...current,
                teams: preservesDeterministicTrelloGraph
                  ? mergeDeterministicImportEntities(
                      current.teams,
                      completedAnalysis.teams,
                    )
                  : completedAnalysis.teams,
                people: preservesDeterministicTrelloGraph
                  ? mergeDeterministicImportEntities(
                      current.people,
                      completedAnalysis.people,
                    )
                  : completedAnalysis.people,
                labels: preservesDeterministicTrelloGraph
                  ? mergeDeterministicImportEntities(
                      current.labels,
                      completedAnalysis.labels,
                    )
                  : completedAnalysis.labels,
                strategicPillars: completedAnalysis.strategicPillars,
                objectives: completedAnalysis.objectives,
                keyResults: completedAnalysis.keyResults,
                sprints: completedAnalysis.sprints,
                mapping,
                sourceNamespace:
                  current.sourceNamespace ?? completedAnalysis.sourceNamespace,
                summary: preservesDeterministicTrelloGraph
                  ? current.summary
                  : completedAnalysis.summary,
                tasks: mergeAnalyzedTaskGraph(
                  mappedTasks,
                  completedAnalysis.tasks,
                  {
                    authoritativeFields: mappingOverrideFields.current,
                    enrichmentOnly: preservesDeterministicTrelloGraph,
                  },
                ),
                warnings: preservesDeterministicTrelloGraph
                  ? [
                      ...new Set([
                        ...current.warnings,
                        ...completedAnalysis.warnings,
                      ]),
                    ].slice(0, 50)
                  : completedAnalysis.warnings,
              };
            }
            if (current) {
              return {
                ...current,
                ...completedAnalysis,
                sourceNamespace:
                  current.sourceNamespace ?? completedAnalysis.sourceNamespace,
              };
            }
            return {
              ...completedAnalysis,
              columns: [],
              fileHash,
              fileName,
              rows: [],
            };
          });
          setAnalysisPending(false);
          setAnalysisNotice("");
          setResponseId(null);
          setReviewPage(0);
          if (completedAnalysis.teams.length > 0) {
            setStructureMode("preserve");
            const sourceTeam = completedAnalysis.teams[0];
            if (teams.length === 0) {
              const sourceDestination =
                getImportSourceTeamDestination(sourceTeam);
              setDestination((current) =>
                current.kind === "new"
                  ? { kind: "new", ...sourceDestination }
                  : current,
              );
            }
          }
        },
        (error: unknown) => {
          if (!isCurrentPollingSession()) return;
          consecutivePollErrors += 1;
          if (consecutivePollErrors > IMPORT_ANALYSIS_POLL_ERROR_RETRIES) {
            handlePollFailure(
              error instanceof Error && error.message.trim()
                ? error
                : new Error(
                    "AI analysis could not be checked after several attempts. You can continue with the deterministic preview or upload the file again.",
                  ),
            );
            return;
          }
          schedulePoll(
            getImportAnalysisPollDelay(pollAttempts, consecutivePollErrors),
          );
        },
      );
    };

    const deadlineTimer = setTimeout(
      handlePollTimeout,
      Math.max(0, deadline - Date.now()),
    );
    schedulePoll(IMPORT_ANALYSIS_INITIAL_POLL_DELAY_MS);
    return () => {
      cancelled = true;
      clearTimeout(deadlineTimer);
      if (pollTimer) clearTimeout(pollTimer);
    };
  }, [
    analysisPending,
    fileHash,
    fileName,
    responseId,
    teams.length,
    workspaceSlug,
  ]);

  const handleFile = (file: File) => {
    const generation = analysisGeneration.current + 1;
    analysisGeneration.current = generation;
    analysisPollingSession.current = null;
    setUploadPending(true);
    setHasAttemptedImport(false);
    setAnalysisError("");
    setAnalysisNotice("");
    setDraft(null);
    setSelectedMemberIdsByIdentityKey(new Map());
    setLockedMemberIdsByIdentityKey(null);
    setFileName(file.name);
    setReviewPage(0);
    setExcludedRows(new Set());
    setIncludeArchivedTrelloCards(false);
    setExcludedObjectives(new Set());
    setExcludedStrategicPillars(new Set());
    createdSourceTeamIds.current.clear();
    sourceObjectiveCache.current.clear();
    createdFallbackTeam.current = null;
    setCreatedFallbackTeamForReview(null);
    setSourceTeamMappingsForReview(new Map());
    setSourceObjectiveMappingsForReview(new Map());
    setObjectivesByTeamId(new Map());
    setObjectivePreflightPending(false);
    setObjectivePreflightError("");
    setObjectivePreflightRevision(0);
    setWorkspaceMembersForReview([]);
    setPeoplePreflightPending(false);
    setPeoplePreflightError("");
    setPeoplePreflightRevision(0);
    setStrategyMapForReview(EMPTY_IMPORT_STRATEGY_MAP);
    setStrategyPreflightPending(false);
    setStrategyPreflightError("");
    setStrategyPreflightRevision(0);
    mappingEdited.current = false;
    mappingOverrideFields.current.clear();
    const completeUpload = () => {
      if (generation === analysisGeneration.current) {
        setUploadPending(false);
      }
    };
    void startImportAnalysis(file, workspaceSlug)
      .then((response) => {
        if (generation !== analysisGeneration.current) return;
        setFileHash(response.fileHash);
        setDraft(response.analysis);
        const archivedTaskSourceIds = getTrelloArchivedTaskSourceIds(
          response.analysis,
        );
        setExcludedRows(
          getTaskIndexesBySourceId(response.analysis, archivedTaskSourceIds),
        );
        setResponseId(response.responseId);
        setAnalysisPending(response.status === "queued");

        if (!response.analysis) return;
        setStructureMode(
          response.analysis.teams.length > 0 ? "preserve" : "single",
        );
        const teamName = getSuggestedTeamName(
          response.analysis.sourceType,
          file.name,
        );
        const sourceTeam = response.analysis.teams.at(0);
        setDestination((current) => {
          if (current.kind !== "new") return current;
          if (teams.length === 0 && sourceTeam) {
            return {
              kind: "new",
              ...getImportSourceTeamDestination(sourceTeam),
            };
          }
          return {
            ...current,
            name: teamName,
            code: formatTeamCode(teamName),
          };
        });
      })
      .catch((error: unknown) => {
        if (generation !== analysisGeneration.current) return;
        setAnalysisError(
          error instanceof Error
            ? error.message
            : "Unable to analyze this file",
        );
        setAnalysisPending(false);
      })
      .then(completeUpload, completeUpload);
  };

  const dropzone = useDropzone({
    accept: fileAccept,
    disabled: uploadPending || analysisPending,
    maxFiles: 1,
    maxSize: IMPORT_MAX_FILE_BYTES,
    multiple: false,
    onDropAccepted: ([file]) => {
      handleFile(file);
    },
    onDropRejected: (rejections) => {
      const tooLarge = rejections.some(({ errors }) =>
        errors.some((error) => error.code === "file-too-large"),
      );
      let message =
        "Choose a CSV, TSV, JSON, Excel, PDF, JPG, PNG, or WebP file.";
      if (tooLarge) message = "The import file must be 20 MB or smaller.";
      setAnalysisError(message);
    },
  });

  const selectedTasks = useMemo(
    () => draft?.tasks.filter((_, index) => !excludedRows.has(index)) ?? [],
    [draft?.tasks, excludedRows],
  );
  const archivedTrelloTaskSourceIds = useMemo(
    () => getTrelloArchivedTaskSourceIds(draft),
    [draft],
  );
  const archivedTrelloTaskIndexes = useMemo(
    () => getTaskIndexesBySourceId(draft, archivedTrelloTaskSourceIds),
    [archivedTrelloTaskSourceIds, draft],
  );
  const selectedStrategicPillars = useMemo(
    () =>
      draft?.strategicPillars.filter(
        (pillar) => !excludedStrategicPillars.has(pillar.sourceId),
      ) ?? [],
    [draft?.strategicPillars, excludedStrategicPillars],
  );
  const selectedObjectives = useMemo(
    () =>
      draft?.objectives.filter(
        (objective) => !excludedObjectives.has(objective.sourceId),
      ) ?? [],
    [draft?.objectives, excludedObjectives],
  );
  const importableKeyResults = useMemo(
    () =>
      draft?.keyResults.filter(
        (keyResult) =>
          keyResult.objectiveSourceId &&
          !excludedObjectives.has(keyResult.objectiveSourceId) &&
          keyResult.measurementType !== null &&
          keyResult.startValue !== null &&
          keyResult.currentValue !== null &&
          keyResult.targetValue !== null &&
          keyResult.startDate !== null &&
          keyResult.endDate !== null &&
          isValidImportDateRange(keyResult.startDate, keyResult.endDate),
      ) ?? [],
    [draft?.keyResults, excludedObjectives],
  );
  const importableSprints = useMemo(
    () =>
      draft?.sprints.filter(
        (sprint) =>
          sprint.startDate !== null &&
          sprint.endDate !== null &&
          isValidImportDateRange(sprint.startDate, sprint.endDate),
      ) ?? [],
    [draft?.sprints],
  );
  const hasSelectedTeamScopedImport = Boolean(
    selectedTasks.length > 0 ||
      selectedObjectives.length > 0 ||
      importableSprints.length > 0 ||
      Boolean(draft?.labels.some((label) => label.teamSourceId !== null)) ||
      (structureMode === "preserve" && (draft?.teams.length ?? 0) > 0),
  );
  const selectedSourceTeamIds = useMemo(() => {
    const sourceTeamIds = new Set<string>();
    const addSourceTeamId = (sourceTeamId: string | null) => {
      if (sourceTeamId) sourceTeamIds.add(sourceTeamId);
    };
    for (const task of selectedTasks) addSourceTeamId(task.teamSourceId);
    for (const objective of selectedObjectives) {
      addSourceTeamId(objective.teamSourceId);
    }
    for (const sprint of importableSprints) {
      addSourceTeamId(sprint.teamSourceId);
    }
    for (const label of draft?.labels ?? []) {
      addSourceTeamId(label.teamSourceId);
    }
    return sourceTeamIds;
  }, [draft?.labels, importableSprints, selectedObjectives, selectedTasks]);
  const privateSourceTeamCount =
    draft?.teams.filter(
      (team) => team.isPrivate && selectedSourceTeamIds.has(team.sourceId),
    ).length ?? 0;
  const destinationIsPrivate =
    destination.kind === "new"
      ? destination.isPrivate
      : knownWorkspaceTeams.find((team) => team.id === destination.teamId)
          ?.isPrivate;
  const hasPrivacyWideningRisk = Boolean(
    hasSelectedTeamScopedImport &&
      structureMode === "single" &&
      privateSourceTeamCount > 0 &&
      destinationIsPrivate === false,
  );
  const selectedTaskParentCycleCount = useMemo(() => {
    const getTargetTeamKey = (sourceTeamId: string | null) => {
      if (structureMode === "single" || !sourceTeamId) {
        return fallbackTargetPlan.teamKey;
      }
      const plan = sourceTeamTargetPlans.get(sourceTeamId);
      return plan?.teamConflict
        ? null
        : plan?.teamKey ?? fallbackTargetPlan.teamKey;
    };
    return getImportParentCycleSourceIds(
      selectedTasks,
      (task, parent) =>
        getTargetTeamKey(task.teamSourceId) ===
        getTargetTeamKey(parent.teamSourceId),
    ).size;
  }, [
    fallbackTargetPlan.teamKey,
    selectedTasks,
    sourceTeamTargetPlans,
    structureMode,
  ]);
  const selectedEntityCount =
    selectedTasks.length +
    selectedStrategicPillars.length +
    selectedObjectives.length +
    importableKeyResults.length +
    importableSprints.length +
    (draft?.labels.length ?? 0) +
    (structureMode === "preserve" ? draft?.teams.length ?? 0 : 0);
  const destinationValid =
    importDestinationSchema.safeParse(destination).success;
  let canContinue = false;
  if (step === WIZARD_STEP.upload) {
    canContinue =
      Boolean(
        draft &&
          (draft.tasks.length > 0 ||
            draft.strategicPillars.length > 0 ||
            draft.objectives.length > 0 ||
            draft.sprints.length > 0 ||
            draft.labels.length > 0 ||
            draft.teams.length > 0),
      ) &&
      !analysisPending &&
      !uploadPending;
  } else if (step === WIZARD_STEP.teams) {
    canContinue =
      (!hasSelectedTeamScopedImport || destinationValid) &&
      !hasPrivacyWideningRisk &&
      (!hasSelectedTeamScopedImport ||
        structureMode === "single" ||
        (!workspaceTeamsPending &&
          !workspaceTeamsError &&
          !hasSourceTeamConflict)) &&
      (!hasSelectedTeamScopedImport ||
        (!objectivePreflightPending && !objectivePreflightError)) &&
      !strategyPreflightPending &&
      !strategyPreflightError;
  } else if (step === WIZARD_STEP.members) {
    canContinue =
      !hasImportIdentities ||
      (!peoplePreflightPending && !peoplePreflightError);
  } else if (step === WIZARD_STEP.review) {
    canContinue =
      selectedEntityCount > 0 &&
      selectedTasks.every((task) => task.title.trim().length > 0) &&
      selectedTaskParentCycleCount === 0 &&
      !hasPrivacyWideningRisk &&
      (!hasSelectedTeamScopedImport ||
        (!objectivePreflightPending && !objectivePreflightError)) &&
      (!hasImportIdentities ||
        (!peoplePreflightPending && !peoplePreflightError)) &&
      !strategyPreflightPending &&
      !strategyPreflightError &&
      selectedStrategicPillars.every((pillar) => {
        if (!pillar.name.trim()) return false;
        const match = strategicPillarDestinationMatches.get(pillar.sourceId);
        return match?.kind === "none" || match?.kind === "unique";
      }) &&
      selectedObjectives.every((objective) => {
        if (!objective.name.trim()) return false;
        const match = objectiveDestinationMatches.get(objective.sourceId);
        return match?.kind === "none" || match?.kind === "unique";
      });
  }

  const updateMapping = (field: keyof ImportMapping, selectedValue: string) => {
    if (isImportMappingFieldLocked(field, hasAttemptedImport)) return;
    const value = selectedValue === DO_NOT_IMPORT_VALUE ? null : selectedValue;
    mappingEdited.current = true;
    mappingOverrideFields.current.add(field);
    setDraft((current) => {
      if (!current?.mapping) return current;
      const mapping = { ...current.mapping, [field]: value };
      return {
        ...current,
        mapping,
        tasks: mergeAnalyzedTaskGraph(
          mapRowsToImportTasks(current.rows, mapping),
          current.tasks,
          { authoritativeFields: mappingOverrideFields.current },
        ),
        warnings:
          field === "sourceId"
            ? [
                ...new Set([
                  `${storyTermCapitalized} parent and ${storyTerm}-association links were cleared because the source ID mapping changed; their previous references could no longer be verified safely.`,
                  ...current.warnings,
                ]),
              ]
            : current.warnings,
      };
    });
    setExcludedRows(new Set());
    setReviewPage(0);
  };

  const updateTaskTitle = (index: number, title: string) => {
    setDraft((current) => {
      if (!current) return current;
      return {
        ...current,
        tasks: current.tasks.map((task, taskIndex) =>
          taskIndex === index ? { ...task, title } : task,
        ),
      };
    });
  };

  const toggleTask = (index: number, checked: boolean) => {
    setExcludedRows((current) => {
      const next = new Set(current);
      if (checked) next.delete(index);
      else next.add(index);
      return next;
    });
  };

  const toggleArchivedTrelloCards = (included: boolean) => {
    setIncludeArchivedTrelloCards(included);
    setExcludedRows((current) => {
      const next = new Set(current);
      for (const index of archivedTrelloTaskIndexes) {
        if (included) next.delete(index);
        else next.add(index);
      }
      return next;
    });
    setReviewPage(0);
  };

  const toggleObjective = (sourceId: string, checked: boolean) => {
    setExcludedObjectives((current) => {
      const next = new Set(current);
      if (checked) next.delete(sourceId);
      else next.add(sourceId);
      return next;
    });
  };

  const toggleStrategicPillar = (sourceId: string, checked: boolean) => {
    setExcludedStrategicPillars((current) => {
      const next = new Set(current);
      if (checked) next.delete(sourceId);
      else next.add(sourceId);
      return next;
    });
  };

  const updateStrategicPillarName = (sourceId: string, name: string) => {
    setDraft((current) => {
      if (!current) return current;
      return {
        ...current,
        strategicPillars: current.strategicPillars.map((pillar) =>
          pillar.sourceId === sourceId ? { ...pillar, name } : pillar,
        ),
      };
    });
  };

  const updateObjectiveName = (sourceId: string, name: string) => {
    setDraft((current) => {
      if (!current) return current;
      return {
        ...current,
        objectives: current.objectives.map((objective) =>
          objective.sourceId === sourceId ? { ...objective, name } : objective,
        ),
      };
    });
  };

  const getDestinationTeam = async () => {
    if (destination.kind === "existing") {
      return { created: false, id: destination.teamId };
    }
    if (!session)
      throw new Error("Your session expired. Refresh and try again.");

    const signature = getNewTeamImportSignature(fileHash, destination);
    if (createdFallbackTeam.current?.signature === signature) {
      return { created: false, id: createdFallbackTeam.current.id };
    }

    const response = await createImportTeam(
      {
        code: destination.code.trim().toUpperCase(),
        color: destination.color,
        isPrivate: destination.isPrivate,
        name: destination.name.trim(),
      },
      { session, workspaceSlug },
    );
    if (response.error?.message || !response.data?.id) {
      throw new Error(response.error?.message || "Unable to create the team");
    }
    createdFallbackTeam.current = { id: response.data.id, signature };
    setCreatedFallbackTeamForReview({ id: response.data.id, signature });
    return { created: true, id: response.data.id };
  };

  const startImport = () => {
    if (!draft || !session || selectedEntityCount === 0) return;
    const memberIdsByIdentityKey =
      lockedMemberIdsByIdentityKey ??
      new Map<string, string | null>(confirmedMemberIdsByIdentityKey);
    if (!lockedMemberIdsByIdentityKey) {
      setLockedMemberIdsByIdentityKey(memberIdsByIdentityKey);
    }
    setHasAttemptedImport(true);
    setStep(WIZARD_STEP.import);
    setAnalysisPending(false);
    setResponseId(null);
    setImportPending(true);
    setImportProgress(0);
    setRunError("");
    setOutcome(null);

    const finishImport = () => {
      setSourceTeamMappingsForReview(new Map(createdSourceTeamIds.current));
      setSourceObjectiveMappingsForReview(
        new Map(sourceObjectiveCache.current),
      );
      void Promise.allSettled([
        queryClient.invalidateQueries({
          queryKey: storyKeys.all(workspaceSlug),
        }),
        queryClient.invalidateQueries({
          queryKey: teamKeys.all(workspaceSlug),
        }),
        queryClient.invalidateQueries({
          queryKey: statusKeys.all(workspaceSlug),
        }),
        queryClient.invalidateQueries({
          queryKey: memberKeys.all(workspaceSlug),
        }),
        queryClient.invalidateQueries({
          queryKey: objectiveKeys.all(workspaceSlug),
        }),
        queryClient.invalidateQueries({
          queryKey: sprintKeys.all(workspaceSlug),
        }),
        queryClient.invalidateQueries({
          queryKey: labelKeys.all(workspaceSlug),
        }),
        queryClient.invalidateQueries({
          queryKey: strategyKeys.map(workspaceSlug),
        }),
      ]).then(() => {
        setImportPending(false);
      });
    };
    const fallbackTeamPromise = hasSelectedTeamScopedImport
      ? getDestinationTeam()
      : Promise.resolve(null);
    void fallbackTeamPromise
      .then((fallbackTeam) => {
        const fallbackTeamId = fallbackTeam?.id ?? null;
        const existingFallbackTeam = fallbackTeamId
          ? knownWorkspaceTeams.find((team) => team.id === fallbackTeamId)
          : undefined;
        return runImport({
          actorUserId: session.user.id,
          confirmedMemberIdsByIdentityKey: memberIdsByIdentityKey,
          ctx: { session, workspaceSlug },
          draft,
          existingTeams: knownWorkspaceTeams,
          fallbackTeamCode:
            destination.kind === "new"
              ? destination.code
              : existingFallbackTeam?.code ?? "",
          fallbackTeamCreated: fallbackTeam?.created ?? false,
          fallbackTeamId,
          fallbackTeamIsPrivate:
            destination.kind === "new"
              ? destination.isPrivate
              : existingFallbackTeam?.isPrivate ?? false,
          fallbackTeamIsNew: Boolean(
            fallbackTeam && destination.kind === "new",
          ),
          fallbackTeamName:
            destination.kind === "new"
              ? destination.name
              : existingFallbackTeam?.name ?? "",
          forceCreateObjectiveSourceIds: EMPTY_IMPORT_SOURCE_IDS,
          joinedTeamIds,
          onProgress: setImportProgress,
          selectedObjectiveSourceIds: new Set(
            selectedObjectives.map((objective) => objective.sourceId),
          ),
          selectedStrategicPillarSourceIds: new Set(
            selectedStrategicPillars.map((pillar) => pillar.sourceId),
          ),
          selectedTaskIndexes: new Set(
            draft.tasks.flatMap((_, index) =>
              excludedRows.has(index) ? [] : [index],
            ),
          ),
          sourceTeamCache: createdSourceTeamIds.current,
          sourceObjectiveCache: sourceObjectiveCache.current,
          structureMode,
        });
      })
      .then(
        (result) => {
          setOutcome(result);
          finishImport();
        },
        (error: unknown) => {
          setRunError(
            error instanceof Error
              ? error.message
              : "The import could not finish",
          );
          finishImport();
        },
      );
  };
  const handleOpenChange = (nextOpen: boolean) => {
    if (!nextOpen && (analysisPending || importPending || uploadPending)) {
      toast.info("Keep this window open while the import is being prepared.");
      return;
    }
    onOpenChange(nextOpen);
    if (!nextOpen) reset();
  };

  const title = "Import work";
  const description =
    "Upload an export from your work tool. FortyOne maps its objects and relationships, then gives you a full review before anything is created.";
  const suggestedTeamName = getSuggestedTeamName(draft?.sourceType, fileName);

  let uploadLabel = "Drop a file here or choose one";
  if (uploadPending) uploadLabel = "Reading your file…";
  else if (dropzone.isDragActive) uploadLabel = "Drop it here";

  const reviewTasks =
    draft?.tasks.flatMap((task, taskIndex) =>
      !includeArchivedTrelloCards && archivedTrelloTaskIndexes.has(taskIndex)
        ? []
        : [{ task, taskIndex }],
    ) ?? [];
  const reviewPageCount = Math.max(
    1,
    Math.ceil(reviewTasks.length / REVIEW_PAGE_SIZE),
  );
  const reviewPageStart = reviewPage * REVIEW_PAGE_SIZE;
  const visibleReviewTasks = reviewTasks.slice(
    reviewPageStart,
    reviewPageStart + REVIEW_PAGE_SIZE,
  );
  const relationshipReview = useMemo(() => {
    if (!draft) return { crossTeam: 0, unresolved: 0 };
    const taskCounts = new Map<string, number>();
    for (const task of selectedTasks) {
      taskCounts.set(task.sourceId, (taskCounts.get(task.sourceId) ?? 0) + 1);
    }
    const tasksBySourceId = new Map(
      selectedTasks
        .filter((task) => taskCounts.get(task.sourceId) === 1)
        .map((task) => [task.sourceId, task]),
    );
    const objectivesBySourceId = new Map(
      selectedObjectives.map((objective) => [objective.sourceId, objective]),
    );
    const keyResultsBySourceId = new Map(
      importableKeyResults.map((keyResult) => [keyResult.sourceId, keyResult]),
    );
    const sprintsBySourceId = new Map(
      importableSprints.map((sprint) => [sprint.sourceId, sprint]),
    );
    const getTargetTeamKey = (sourceTeamId: string | null) => {
      if (structureMode === "single" || !sourceTeamId) {
        return fallbackTargetPlan.teamKey;
      }
      const plan = sourceTeamTargetPlans.get(sourceTeamId);
      return plan?.teamConflict
        ? null
        : plan?.teamKey ?? fallbackTargetPlan.teamKey;
    };
    let crossTeam = 0;
    let unresolved = 0;
    const seenAssociationKeys = new Set<string>();

    for (const task of selectedTasks) {
      const taskTeamKey = getTargetTeamKey(task.teamSourceId);
      if (task.parentSourceId) {
        const parent = tasksBySourceId.get(task.parentSourceId);
        if (!parent) unresolved += 1;
        else if (
          taskTeamKey &&
          getTargetTeamKey(parent.teamSourceId) !== taskTeamKey
        ) {
          crossTeam += 1;
        }
      }
      if (
        task.objectiveSourceId &&
        !objectivesBySourceId.has(task.objectiveSourceId)
      ) {
        unresolved += 1;
      }
      if (
        task.keyResultSourceId &&
        !keyResultsBySourceId.has(task.keyResultSourceId)
      ) {
        unresolved += 1;
      }
      if (task.sprintSourceId) {
        const sprint = sprintsBySourceId.get(task.sprintSourceId);
        if (!sprint) unresolved += 1;
        else if (
          taskTeamKey &&
          getTargetTeamKey(sprint.teamSourceId) !== taskTeamKey
        ) {
          crossTeam += 1;
        }
      }
      for (const association of task.associations) {
        const targetSourceId = association.targetSourceId.trim();
        const canonicalAssociation = getCanonicalImportAssociation(
          task.sourceId,
          targetSourceId,
          association.type,
        );
        const associationKey = getImportAssociationKey(canonicalAssociation);
        if (seenAssociationKeys.has(associationKey)) continue;
        seenAssociationKeys.add(associationKey);
        const target = tasksBySourceId.get(targetSourceId);
        if (!target || target.sourceId === task.sourceId) unresolved += 1;
        else if (
          taskTeamKey &&
          getTargetTeamKey(target.teamSourceId) !== taskTeamKey
        ) {
          crossTeam += 1;
        }
      }
    }
    for (const sprint of importableSprints) {
      if (!sprint.objectiveSourceId) continue;
      const objective = objectivesBySourceId.get(sprint.objectiveSourceId);
      if (!objective) unresolved += 1;
      else if (
        getTargetTeamKey(sprint.teamSourceId) !==
        getTargetTeamKey(objective.teamSourceId)
      ) {
        crossTeam += 1;
      }
    }
    return { crossTeam, unresolved };
  }, [
    draft,
    fallbackTargetPlan.teamKey,
    importableKeyResults,
    importableSprints,
    selectedObjectives,
    selectedTasks,
    sourceTeamTargetPlans,
    structureMode,
  ]);
  const reviewWarnings = useMemo(() => {
    if (!draft) return [];
    const importableKeyResultSourceIds = new Set(
      importableKeyResults.map((keyResult) => keyResult.sourceId),
    );
    const incompleteKeyResults = draft.keyResults.filter(
      (keyResult) =>
        keyResult.objectiveSourceId &&
        !excludedObjectives.has(keyResult.objectiveSourceId) &&
        !importableKeyResultSourceIds.has(keyResult.sourceId),
    ).length;
    const incompleteSprints = draft.sprints.length - importableSprints.length;
    const invalidStoryDates = selectedTasks.filter(
      (task) => !isValidImportDateRange(task.startDate, task.endDate),
    ).length;
    const invalidObjectiveDates = selectedObjectives.filter(
      (objective) =>
        !isValidImportDateRange(objective.startDate, objective.endDate),
    ).length;
    const omittedPillarAlignments = selectedObjectives.filter(
      (objective) =>
        objective.pillarSourceId &&
        excludedStrategicPillars.has(objective.pillarSourceId),
    ).length;
    return [
      ...(incompleteKeyResults
        ? [
            `${incompleteKeyResults} ${incompleteKeyResults === 1 ? `${keyResultTerm} is` : `${keyResultTermPlural} are`} missing an explicit measure or valid date range and will be skipped.`,
          ]
        : []),
      ...(incompleteSprints
        ? [
            `${incompleteSprints} ${incompleteSprints === 1 ? `${sprintTerm} is` : `${sprintTermPlural} are`} missing a valid date range and will be skipped.`,
          ]
        : []),
      ...(invalidStoryDates
        ? [
            `${invalidStoryDates} ${invalidStoryDates === 1 ? `${storyTerm} has` : `${storyTermPlural} have`} an invalid date range; the work will import without those dates.`,
          ]
        : []),
      ...(invalidObjectiveDates
        ? [
            `${invalidObjectiveDates} ${invalidObjectiveDates === 1 ? `${objectiveTerm} has` : `${objectiveTermPlural} have`} an invalid date range; the ${objectiveTerm} will import without those dates.`,
          ]
        : []),
      ...(omittedPillarAlignments
        ? [
            `${omittedPillarAlignments} selected ${omittedPillarAlignments === 1 ? `${objectiveTerm} references` : `${objectiveTermPlural} reference`} a pillar that is not selected and will import without that alignment.`,
          ]
        : []),
      ...(selectedTaskParentCycleCount
        ? [
            `${selectedTaskParentCycleCount} selected ${selectedTaskParentCycleCount === 1 ? `${storyTerm} forms` : `${storyTermPlural} form`} a circular parent chain. Deselect at least one ${storyTerm} in each cycle before importing.`,
          ]
        : []),
      ...(relationshipReview.crossTeam
        ? [
            `${relationshipReview.crossTeam} ${relationshipReview.crossTeam === 1 ? "relationship spans" : "relationships span"} destination teams. Parent, ${sprintTerm}, and ${storyTerm}-association links must stay within one team, so these links will be skipped and reported.`,
          ]
        : []),
      ...(relationshipReview.unresolved
        ? [
            `${relationshipReview.unresolved} ${relationshipReview.unresolved === 1 ? "relationship points" : "relationships point"} to an object that is missing or not selected and will not be linked.`,
          ]
        : []),
      ...draft.warnings,
    ];
  }, [
    draft,
    excludedObjectives,
    excludedStrategicPillars,
    importableKeyResults,
    importableSprints.length,
    keyResultTerm,
    keyResultTermPlural,
    objectiveTerm,
    objectiveTermPlural,
    relationshipReview,
    selectedObjectives,
    selectedTasks,
    selectedTaskParentCycleCount,
    sprintTerm,
    sprintTermPlural,
    storyTerm,
    storyTermPlural,
  ]);

  let sourceTeamReview: ReactNode = null;
  if (structureMode === "preserve" && draft) {
    if (workspaceTeamsPending) {
      sourceTeamReview = (
        <Text className="mt-5" color="muted">
          Checking workspace teams for safe reuse…
        </Text>
      );
    } else if (workspaceTeamsError) {
      sourceTeamReview = (
        <Box className="bg-danger/8 mt-5 rounded-xl p-4">
          <Text className="text-danger font-medium">
            Workspace teams could not be checked
          </Text>
          <Text className="mt-1" color="muted">
            Try again before preserving source teams, or combine this import
            into one destination.
          </Text>
        </Box>
      );
    } else {
      sourceTeamReview = (
        <>
          <SourceTeamPreview
            conflictedSourceIds={conflictedSourceTeamIds}
            draft={draft}
            existingTeams={knownWorkspaceTeams}
            joinedTeamIds={joinedTeamIds}
          />
          {hasSourceTeamConflict ? (
            <Box className="bg-warning/8 mt-3 rounded-xl p-4">
              <Text className="font-medium">Team mapping needs review</Text>
              <Text className="mt-1" color="muted">
                Combine this import into one destination, or make the workspace
                team names and codes unambiguous before preserving source teams.
              </Text>
            </Box>
          ) : null}
        </>
      );
    }
  }

  let footerBack: ReactNode = <span />;
  if (step > WIZARD_STEP.upload && step < WIZARD_STEP.import) {
    footerBack = (
      <Button
        color="tertiary"
        leftIcon={<ArrowLeft2Icon />}
        onClick={() => {
          setStep((current) => Math.max(WIZARD_STEP.upload, current - 1));
        }}
        variant="outline"
      >
        Back
      </Button>
    );
  } else if (step === WIZARD_STEP.import && !importPending && !outcome) {
    footerBack = (
      <Button
        color="tertiary"
        leftIcon={<ArrowLeft2Icon />}
        onClick={() => {
          setObjectivePreflightRevision((current) => current + 1);
          setPeoplePreflightRevision((current) => current + 1);
          setStrategyPreflightRevision((current) => current + 1);
          setStep(WIZARD_STEP.review);
        }}
        variant="outline"
      >
        Review again
      </Button>
    );
  } else if (
    step === WIZARD_STEP.import &&
    (outcome?.failed ||
      outcome?.destinationConflicts ||
      outcome?.unresolvedAssociations ||
      outcome?.unresolvedLinks ||
      outcome?.unresolvedPeople)
  ) {
    footerBack = (
      <Button
        color="tertiary"
        leftIcon={<ArrowLeft2Icon />}
        onClick={() => {
          setOutcome(null);
          setObjectivePreflightRevision((current) => current + 1);
          setPeoplePreflightRevision((current) => current + 1);
          setStrategyPreflightRevision((current) => current + 1);
          setStep(WIZARD_STEP.review);
        }}
        variant="outline"
      >
        Review import
      </Button>
    );
  }

  const outcomeStrategyChanges = outcome
    ? outcome.createdStrategicPillars + outcome.alignedObjectives
    : 0;
  const outcomeNonStrategyChanges = outcome
    ? outcome.created +
      outcome.replayed +
      outcome.createdTeams +
      outcome.createdObjectives +
      outcome.createdKeyResults +
      outcome.createdSprints +
      outcome.createdLabels +
      outcome.createdLinks +
      outcome.addedMemberships +
      outcome.appliedCollaborators +
      outcome.createdAssociations
    : 0;
  const outcomeHasIssues = Boolean(
    outcome &&
      (outcome.failed ||
        outcome.destinationConflicts ||
        outcome.unresolvedAssociations ||
        outcome.unresolvedLinks ||
        outcome.unresolvedPeople),
  );
  const canViewOutcome = Boolean(
    outcome &&
      (outcomeStrategyChanges + outcomeNonStrategyChanges > 0 ||
        !outcomeHasIssues),
  );
  const viewOutcomeInStrategy = Boolean(
    outcome && outcomeStrategyChanges > 0 && outcomeNonStrategyChanges === 0,
  );

  let footerAction: ReactNode = null;
  if (step < WIZARD_STEP.review) {
    footerAction = (
      <Button
        color="invert"
        disabled={!canContinue}
        onClick={() => {
          setStep((current) => current + 1);
        }}
      >
        Continue
      </Button>
    );
  } else if (step === WIZARD_STEP.review) {
    footerAction = (
      <Button color="invert" disabled={!canContinue} onClick={startImport}>
        Import {selectedEntityCount}{" "}
        {selectedEntityCount === 1 ? "item" : "items"}
      </Button>
    );
  } else if (outcome && canViewOutcome) {
    footerAction = (
      <Button
        color="invert"
        href={withWorkspace(viewOutcomeInStrategy ? "/strategy" : "/summary")}
        onClick={() => {
          handleOpenChange(false);
        }}
        rightIcon={<ArrowRight2Icon />}
      >
        {viewOutcomeInStrategy ? "View strategy" : "View workspace summary"}
      </Button>
    );
  } else if (outcome) {
    footerAction = (
      <Button color="invert" onClick={startImport}>
        Retry safely
      </Button>
    );
  } else if (!importPending) {
    footerAction = (
      <Button color="invert" onClick={startImport}>
        Retry safely
      </Button>
    );
  }

  return (
    <Dialog onOpenChange={handleOpenChange} open={open}>
      <Dialog.Content
        className="mt-0 flex max-h-[calc(100dvh-2rem)] max-w-3xl flex-col md:mt-0"
        onEscapeKeyDown={(event) => {
          event.preventDefault();
        }}
        onInteractOutside={(event) => {
          event.preventDefault();
        }}
        overlayClassName="items-center py-4"
        size="lg"
      >
        <Dialog.Header className="shrink-0 px-6 pt-5 pb-2">
          <Dialog.Title className="text-xl">{title}</Dialog.Title>
          <Dialog.Description className="mt-2 px-0 text-base leading-6">
            {description}
          </Dialog.Description>
          <WizardProgress step={step} />
        </Dialog.Header>

        <Dialog.Body className="min-h-0 flex-1 overflow-y-auto overscroll-contain px-6 pt-4 pb-6">
          {step === WIZARD_STEP.upload ? (
            <Box>
              <Text as="h2" className="text-xl font-medium">
                Add your export
              </Text>
              <Text className="mt-1 leading-6" color="muted">
                Export projects, teams, tasks, or issues from Jira, Trello,
                ClickUp, monday.com, Asana, or another tool, then upload the
                file here. Mapping starts automatically, and you can review
                every suggestion before the import runs.
              </Text>

              {fileName ? (
                <DropZone>
                  <DropZone.Input inputProps={dropzone.getInputProps()} />
                  <ImportAnalysisBanner
                    analysisError={analysisError}
                    analysisNotice={analysisNotice}
                    analysisPending={analysisPending}
                    fileName={fileName}
                    hasAttemptedImport={hasAttemptedImport}
                    key={fileName}
                    onReplace={dropzone.open}
                    uploadPending={uploadPending}
                  />
                </DropZone>
              ) : (
                <DropZone>
                  <DropZone.Root
                    className="bg-surface-muted/35 mt-5 h-44"
                    isDragActive={dropzone.isDragActive}
                    rootProps={dropzone.getRootProps()}
                  >
                    <DropZone.Input inputProps={dropzone.getInputProps()} />
                    <Flex align="center" direction="column" gap={2}>
                      <Box className="bg-surface flex h-12 w-12 items-center justify-center rounded-xl">
                        <DownloadIcon className="h-6" />
                      </Box>
                      <Text className="font-medium">{uploadLabel}</Text>
                      <Text color="muted">
                        CSV, JSON, Excel, PDF, JPG, PNG, or WebP, up to 20 MB
                      </Text>
                    </Flex>
                  </DropZone.Root>
                </DropZone>
              )}

              {analysisError && !fileName ? (
                <Box className="bg-danger/8 mt-4 rounded-2xl p-4">
                  <Text className="text-danger font-medium">
                    Unable to continue
                  </Text>
                  <Text className="mt-1" color="muted">
                    {analysisError}
                  </Text>
                </Box>
              ) : null}
            </Box>
          ) : null}

          {step === WIZARD_STEP.teams ? (
            <Box>
              <Text as="h2" className="text-xl font-medium">
                {hasSelectedTeamScopedImport
                  ? "Where should this work live?"
                  : "Review workspace-level objects"}
              </Text>
              <Text className="mt-1 leading-6" color="muted">
                {hasSelectedTeamScopedImport
                  ? "Keep useful source structure or combine everything into one destination. No team will be created before the final review."
                  : "Strategic pillars and global labels live at workspace level. This import does not need or create a destination team."}
              </Text>

              {draft ? <ImportPlanSummary draft={draft} /> : null}

              {hasAttemptedImport ? (
                <Box
                  className="bg-warning/8 mt-5 rounded-xl p-4"
                  id="import-destination-lock-note"
                >
                  <Text className="font-medium">
                    Import setup is locked for safe retries
                  </Text>
                  <Text className="mt-1 leading-6" color="muted">
                    The destination, team structure, and privacy stay fixed
                    after the first attempt. Upload a new file to change them.
                  </Text>
                </Box>
              ) : null}

              {strategyPreflightPending ? (
                <Text className="mt-5" color="muted">
                  Checking existing strategic pillars for safe reuse…
                </Text>
              ) : null}
              {strategyPreflightError ? (
                <Box className="bg-danger/8 mt-5 rounded-xl p-4">
                  <Text className="text-danger font-medium">
                    Strategic pillars could not be checked
                  </Text>
                  <Text className="mt-1" color="muted">
                    Check again before continuing so the import cannot create a
                    duplicate pillar or reuse an ambiguous match.
                  </Text>
                  <Button
                    className="mt-3"
                    color="tertiary"
                    onClick={() => {
                      setStrategyPreflightRevision((current) => current + 1);
                    }}
                    size="sm"
                    variant="outline"
                  >
                    Check again
                  </Button>
                </Box>
              ) : null}

              {hasSelectedTeamScopedImport ? (
                <>
                  {draft?.teams.length ? (
                    <>
                      <Box className="mt-5 grid gap-3 md:grid-cols-2">
                        <SelectionCard
                          description="Reuse or create source teams, joining matched teams when needed."
                          disabled={hasAttemptedImport}
                          icon={<TeamIcon />}
                          label="Preserve source teams"
                          onClick={() => {
                            setStructureMode("preserve");
                          }}
                          selected={structureMode === "preserve"}
                        />
                        <SelectionCard
                          description={`Put teams, ${objectiveTermPlural}, ${sprintTermPlural}, and ${storyTermPlural} into one team you choose.`}
                          disabled={hasAttemptedImport}
                          icon={<PlusIcon />}
                          label="Combine into one team"
                          onClick={() => {
                            setStructureMode("single");
                          }}
                          selected={structureMode === "single"}
                        />
                      </Box>
                      {sourceTeamReview}
                    </>
                  ) : null}

                  <Text className="mt-5 font-medium">
                    {structureMode === "preserve"
                      ? "Fallback destination"
                      : "Destination team"}
                  </Text>
                  <Text className="mt-1 leading-6" color="muted">
                    {structureMode === "preserve"
                      ? "Work without a reliable source-team relationship will go here."
                      : "Choose a team you belong to or create a focused destination."}
                  </Text>

                  <Box className="mt-3 grid gap-3 md:grid-cols-2">
                    <SelectionCard
                      description="Use a team you already belong to for imported work."
                      disabled={hasAttemptedImport || teams.length === 0}
                      icon={<TeamIcon />}
                      label="Existing team"
                      onClick={() => {
                        setDestination({
                          kind: "existing",
                          teamId: teams[0]?.id ?? "",
                        });
                      }}
                      selected={destination.kind === "existing"}
                    />
                    <SelectionCard
                      description={`Create a team with its own workflow and ${storyTerm} sequence.`}
                      disabled={hasAttemptedImport}
                      icon={<PlusIcon />}
                      label="Create a new team"
                      onClick={() => {
                        setDestination({
                          ...initialNewTeam,
                          code: formatTeamCode(suggestedTeamName),
                          name: suggestedTeamName,
                        });
                      }}
                      selected={destination.kind === "new"}
                    />
                  </Box>

                  {destination.kind === "existing" ? (
                    <DestinationTeamPicker
                      disabled={hasAttemptedImport}
                      onChange={(teamId) => {
                        setDestination({ kind: "existing", teamId });
                      }}
                      teams={teams}
                      value={destination.teamId}
                    />
                  ) : (
                    <Box className="border-border bg-surface mt-5 rounded-xl border-[0.5px] p-5">
                      <Box className="grid gap-4 md:grid-cols-[minmax(0,1fr)_8rem_7rem]">
                        <Input
                          className="text-base"
                          disabled={hasAttemptedImport}
                          label="Team name"
                          maxLength={24}
                          minLength={3}
                          onChange={(event) => {
                            const name = event.target.value;
                            setDestination((current) =>
                              current.kind === "new"
                                ? {
                                    ...current,
                                    name,
                                    code: current.code
                                      ? current.code
                                      : formatTeamCode(name),
                                  }
                                : current,
                            );
                          }}
                          placeholder="Product migration"
                          required
                          value={destination.name}
                        />
                        <Input
                          className="text-base uppercase"
                          disabled={hasAttemptedImport}
                          label="Team code"
                          maxLength={3}
                          minLength={2}
                          onChange={(event) => {
                            setDestination((current) =>
                              current.kind === "new"
                                ? {
                                    ...current,
                                    code: formatTeamCode(event.target.value),
                                  }
                                : current,
                            );
                          }}
                          placeholder="PM"
                          required
                          value={destination.code}
                        />
                        <Box>
                          <Text className="mb-2 font-medium">Team color</Text>
                          <ColorPicker
                            ariaLabel="Choose team color"
                            className="h-12 w-12 rounded-xl"
                            disabled={hasAttemptedImport}
                            onChange={(color) => {
                              setDestination((current) =>
                                current.kind === "new"
                                  ? { ...current, color }
                                  : current,
                              );
                            }}
                            size="lg"
                            value={destination.color}
                          />
                        </Box>
                      </Box>
                      <Flex
                        align="start"
                        className="border-border mt-4 border-t-[0.5px] pt-4"
                        gap={3}
                      >
                        <Checkbox
                          checked={destination.isPrivate}
                          className="mt-1"
                          disabled={hasAttemptedImport}
                          id="import-private-team"
                          onCheckedChange={(checked) => {
                            setDestination((current) =>
                              current.kind === "new"
                                ? { ...current, isPrivate: checked === true }
                                : current,
                            );
                          }}
                        />
                        <label
                          className="cursor-pointer"
                          htmlFor="import-private-team"
                        >
                          <Text className="font-medium">Private team</Text>
                          <Text className="mt-0.5" color="muted">
                            Only invited members can see its imported work.
                          </Text>
                        </label>
                      </Flex>
                    </Box>
                  )}

                  {hasPrivacyWideningRisk ? (
                    <Box className="bg-danger/8 mt-5 rounded-xl p-4">
                      <Text className="text-danger font-medium">
                        Private source work needs a private destination
                      </Text>
                      <Text className="mt-1 leading-6" color="muted">
                        {privateSourceTeamCount}{" "}
                        {privateSourceTeamCount === 1
                          ? "private source team is"
                          : "private source teams are"}{" "}
                        included. Choose a private destination team or preserve
                        the source teams before continuing, so this work is not
                        exposed more broadly.
                      </Text>
                    </Box>
                  ) : null}

                  {objectiveTargetTeamIds.length > 0 &&
                  objectivePreflightPending ? (
                    <Text className="mt-5" color="muted">
                      Checking existing {objectiveTermPlural} for safe reuse…
                    </Text>
                  ) : null}
                  {objectivePreflightError ? (
                    <Box className="bg-danger/8 mt-5 rounded-xl p-4">
                      <Text className="text-danger font-medium">
                        {objectiveTermPluralCapitalized} could not be checked
                      </Text>
                      <Text className="mt-1" color="muted">
                        Check again before continuing so the import cannot
                        create a duplicate or reuse an incompatible{" "}
                        {objectiveTerm}.
                      </Text>
                      <Button
                        className="mt-3"
                        color="tertiary"
                        onClick={() => {
                          setObjectivePreflightRevision(
                            (current) => current + 1,
                          );
                        }}
                        size="sm"
                        variant="outline"
                      >
                        Check again
                      </Button>
                    </Box>
                  ) : null}
                </>
              ) : (
                <Box className="border-border bg-surface mt-5 rounded-xl border-[0.5px] p-4">
                  <Text className="font-medium">No team changes</Text>
                  <Text className="mt-1 leading-6" color="muted">
                    The selected pillars will be created or safely reused in the
                    workspace strategy map. Existing teams and memberships will
                    stay unchanged.
                  </Text>
                </Box>
              )}
            </Box>
          ) : null}

          {step === WIZARD_STEP.members ? (
            <Box>
              <Text as="h2" className="text-xl font-medium">
                Map members
              </Text>
              <Text className="mt-1 leading-6" color="muted">
                Review who imported assignments belong to. Likely matches are
                selected automatically, and anyone can remain unassigned.
              </Text>

              {hasAttemptedImport ? (
                <Text className="mt-4" color="muted">
                  Member mappings are locked for safe retries.
                </Text>
              ) : null}

              {peoplePreflightPending ? (
                <Text className="mt-5" color="muted">
                  Checking source identities against workspace members…
                </Text>
              ) : null}

              {peoplePreflightError ? (
                <Box className="bg-danger/8 mt-5 rounded-xl p-4">
                  <Text className="text-danger font-medium">
                    Workspace members could not be checked
                  </Text>
                  <Text className="mt-1" color="muted">
                    Check again before continuing so assignments and team
                    memberships remain reviewable.
                  </Text>
                  <Button
                    className="mt-3"
                    color="tertiary"
                    onClick={() => {
                      setPeoplePreflightRevision((current) => current + 1);
                    }}
                    size="sm"
                    variant="outline"
                  >
                    Check again
                  </Button>
                </Box>
              ) : null}

              {!hasImportIdentities ? (
                <Text className="mt-4" color="muted">
                  No source members need mapping for this import.
                </Text>
              ) : null}

              {peopleMappingPreview.length > 0 &&
              !peoplePreflightPending &&
              !peoplePreflightError ? (
                <Box className="border-border/70 bg-surface/60 mt-4 overflow-hidden rounded-xl border-[0.5px]">
                  <Flex
                    align="center"
                    className="border-border border-b-[0.5px] px-4 py-2.5"
                    justify="between"
                  >
                    <Text className="font-medium">Source members</Text>
                    <Text color="muted">
                      {peopleMappingPreview.length} detected
                    </Text>
                  </Flex>
                  <Box className="divide-border max-h-72 divide-y overflow-y-auto">
                    {peopleMappingPreview.map(
                      ({ identity, identityKey, suggestedMember }) => {
                        const sourceLabel =
                          identity.name ||
                          identity.email ||
                          `Source identity ${identity.sourceId ?? "unknown"}`;
                        const selectedMemberId =
                          reviewedMemberIdsByIdentityKey.get(identityKey);
                        const selectedMember = workspaceMembersForReview.find(
                          (member) => member.id === selectedMemberId,
                        );
                        return (
                          <Flex
                            align="center"
                            className="flex-col items-stretch px-4 py-2.5 md:flex-row md:items-center"
                            gap={3}
                            justify="between"
                            key={identityKey}
                          >
                            <Text className="min-w-0 truncate font-medium">
                              {sourceLabel}
                            </Text>
                            <ImportMemberPicker
                              disabled={hasAttemptedImport}
                              members={workspaceMembersForReview}
                              onChange={(memberId) => {
                                setSelectedMemberIdsByIdentityKey((current) => {
                                  const next = new Map(current);
                                  next.set(identityKey, memberId);
                                  return next;
                                });
                              }}
                              suggestedMemberId={suggestedMember?.id}
                              value={selectedMember?.id ?? null}
                            />
                          </Flex>
                        );
                      },
                    )}
                  </Box>
                </Box>
              ) : null}
            </Box>
          ) : null}

          {step === WIZARD_STEP.review && draft ? (
            <Box>
              <Text as="h2" className="text-xl font-medium">
                Review your import
              </Text>
              <Text className="mt-1 leading-6" color="muted">
                Confirm the proposed structure, field mapping, and work that
                FortyOne will create.
              </Text>

              <ImportPlanSummary draft={draft} />

              {reviewWarnings.length ? (
                <Box className="bg-warning/8 mt-5 rounded-xl p-4">
                  <Text className="text-foreground/90 font-medium dark:text-white/90">
                    Mapping notes
                  </Text>
                  <Box
                    as="ul"
                    className="text-foreground/75 mt-2 max-h-24 overflow-y-auto pr-2 pl-5 dark:text-white/70"
                  >
                    {reviewWarnings.map((warning) => (
                      <li className="list-disc leading-6" key={warning}>
                        {warning}
                      </li>
                    ))}
                  </Box>
                </Box>
              ) : null}

              <StrategicPillarImportReview
                destinationMatches={strategicPillarDestinationMatches}
                draft={draft}
                excludedSourceIds={excludedStrategicPillars}
                onCheckedChange={toggleStrategicPillar}
                onNameChange={updateStrategicPillarName}
              />

              <ObjectiveImportReview
                destinationMatches={objectiveDestinationMatches}
                draft={draft}
                excludedSourceIds={excludedObjectives}
                onCheckedChange={toggleObjective}
                onNameChange={updateObjectiveName}
              />

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
                        Retries use stable source IDs to recognize{" "}
                        {storyTermPlural}
                        already created and complete only unfinished work.
                        Previously created {storyTermPlural} are reused as-is;
                        {storyTerm}-field edits apply only to {storyTermPlural}
                        that have not been created yet. Upload a new file to
                        change the Source ID mapping.
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
                            {required ? (
                              <span className="text-danger"> *</span>
                            ) : null}
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
                                mappingFieldLocked &&
                                  "cursor-not-allowed opacity-60",
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

              {draft.tasks.length ? (
                <Box className="border-border bg-surface mt-5 overflow-hidden rounded-xl border-[0.5px]">
                  <Flex
                    align="center"
                    className="border-border border-b-[0.5px] px-4 py-3"
                    justify="between"
                  >
                    <Text className="font-medium">
                      {storyTermCapitalized} review
                    </Text>
                    <Flex
                      align="center"
                      className="flex-wrap justify-end"
                      gap={4}
                    >
                      {archivedTrelloTaskIndexes.size > 0 ? (
                        <Flex align="center" gap={2}>
                          <Checkbox
                            aria-label={`Include ${archivedTrelloTaskIndexes.size} archived cards`}
                            checked={includeArchivedTrelloCards}
                            id="include-archived-trello-cards"
                            onCheckedChange={(checked) => {
                              toggleArchivedTrelloCards(checked === true);
                            }}
                          />
                          <label
                            className="cursor-pointer whitespace-nowrap"
                            htmlFor="include-archived-trello-cards"
                          >
                            Include archived ({archivedTrelloTaskIndexes.size})
                          </label>
                        </Flex>
                      ) : null}
                      <Text color="muted">{selectedTasks.length} selected</Text>
                    </Flex>
                  </Flex>
                  <Box className="divide-border max-h-96 divide-y overflow-y-auto">
                    {visibleReviewTasks.map(({ task, taskIndex }) => {
                      const isExcluded = excludedRows.has(taskIndex);
                      const titleMissing =
                        !isExcluded && task.title.trim().length === 0;
                      return (
                        <Flex
                          align="center"
                          className={cn(
                            "gap-2.5 px-4 py-2",
                            isExcluded && "opacity-55",
                          )}
                          key={`${task.sourceId}-${taskIndex}`}
                        >
                          <Checkbox
                            aria-label={`Import ${task.title}`}
                            checked={!isExcluded}
                            onCheckedChange={(checked) => {
                              toggleTask(taskIndex, checked === true);
                            }}
                          />
                          <Box className="min-w-0 flex-1">
                            <Input
                              aria-invalid={titleMissing}
                              aria-label={`Title for ${task.sourceId}`}
                              className={cn(
                                "h-9 bg-transparent px-2 text-base font-medium dark:bg-transparent",
                                titleMissing
                                  ? "border-danger"
                                  : "hover:border-input border-transparent",
                              )}
                              disabled={isExcluded}
                              hasError={titleMissing}
                              maxLength={255}
                              onChange={(event) => {
                                updateTaskTitle(taskIndex, event.target.value);
                              }}
                              placeholder="Add a title"
                              value={task.title}
                            />
                          </Box>
                        </Flex>
                      );
                    })}
                  </Box>
                  <Flex
                    className="border-border flex-col items-stretch border-t-[0.5px] px-4 py-3 sm:flex-row sm:items-center"
                    gap={3}
                    justify="between"
                  >
                    <Text color="muted">
                      {reviewTasks.length > 0
                        ? `Showing ${reviewPageStart + 1}–${Math.min(
                            reviewPageStart + REVIEW_PAGE_SIZE,
                            reviewTasks.length,
                          )} of ${reviewTasks.length}`
                        : "No active cards to review"}
                    </Text>
                    {reviewPageCount > 1 ? (
                      <Flex className="justify-end" gap={2}>
                        <Button
                          color="tertiary"
                          disabled={reviewPage === 0}
                          onClick={() => {
                            setReviewPage((current) =>
                              Math.max(0, current - 1),
                            );
                          }}
                          variant="outline"
                        >
                          Previous
                        </Button>
                        <Button
                          color="tertiary"
                          disabled={reviewPage >= reviewPageCount - 1}
                          onClick={() => {
                            setReviewPage((current) =>
                              Math.min(reviewPageCount - 1, current + 1),
                            );
                          }}
                          variant="outline"
                        >
                          Next
                        </Button>
                      </Flex>
                    ) : null}
                  </Flex>
                </Box>
              ) : null}
            </Box>
          ) : null}

          {step === WIZARD_STEP.import ? (
            <Box className="py-6">
              <ImportRunStep
                importPending={importPending}
                importProgress={importProgress}
                outcome={outcome}
                runError={runError}
              />
            </Box>
          ) : null}
        </Dialog.Body>

        <Dialog.Footer
          className="bg-surface-muted/35 shrink-0 gap-3 px-6 py-4"
          justify="between"
          variant="bordered"
        >
          {footerBack}
          {footerAction}
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
