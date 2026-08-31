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
import { useWorkspacePath } from "@/hooks";
import { TeamColor } from "@/components/ui/team-color";
import { storyKeys } from "@/modules/stories/constants";
import { statusKeys, teamKeys } from "@/constants/keys";
import { useJoinedTeams } from "@/modules/teams/hooks/teams";
import type { Team } from "@/modules/teams/types";
import type { ImportDraft, ImportMapping } from "../schema";
import {
  IMPORT_MAX_FILE_BYTES,
  JIRA_ISSUE_KEY_PATTERN,
  importDestinationSchema,
} from "../schema";
import { mapRowsToImportTasks } from "../csv";
import { getBoundedImportSourceKey, toImportStoryPayload } from "../execution";
import {
  createImportTeam,
  buildImportStoryRequests,
  getImportTeamMembers,
  getImportTeamStatuses,
  importStoriesBatch,
  pollImportAnalysis,
  startImportAnalysis,
  type ImportStoryResult,
} from "../api";

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

type ImportOutcome = {
  created: number;
  replayed: number;
  failed: number;
  items: ImportStoryResult[];
  teamId: string;
};

const STEPS = ["Upload", "Organize", "Review", "Import"] as const;
const REVIEW_PAGE_SIZE = 50;
const MAX_ANALYSIS_POLLS = 80;
const DO_NOT_IMPORT_VALUE = "__do_not_import__";

const IMPORT_ANALYSIS_PHASES = [
  {
    detail: "Securely uploading your export.",
    label: "Upload",
    title: "Uploading your file",
  },
  {
    detail: "Finding rows, headers, and source fields.",
    label: "Read",
    title: "Reading your export",
  },
  {
    detail: "Matching source fields to FortyOne stories.",
    label: "Map",
    title: "Mapping your work",
  },
  {
    detail: "Your mapping is ready to organize.",
    label: "Prepare",
    title: "Ready to organize",
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

const SelectionCard = ({
  description,
  icon,
  label,
  onClick,
  selected,
}: {
  description: string;
  icon: ReactNode;
  label: string;
  onClick: () => void;
  selected: boolean;
}) => (
  <button
    aria-pressed={selected}
    className={cn(
      "border-border bg-surface hover:border-border-strong focus-visible:ring-ring relative min-h-24 rounded-xl border-[0.5px] p-4 text-left transition-[border-color,background-color,box-shadow] focus-visible:ring-2 focus-visible:outline-none",
      selected && "border-primary bg-primary/5 ring-primary/15 ring-1",
    )}
    onClick={onClick}
    type="button"
  >
    <Flex align="start" gap={3}>
      <Box
        className={cn(
          "bg-surface-muted text-text-muted flex h-10 w-10 shrink-0 items-center justify-center rounded-lg transition-colors",
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
  onReplace,
  uploadPending,
}: {
  analysisError: string;
  analysisNotice: string;
  analysisPending: boolean;
  fileName: string;
  onReplace: () => void;
  uploadPending: boolean;
}) => {
  const [uploadPhase, setUploadPhase] = useState(0);
  const busy = uploadPending || analysisPending;

  useEffect(() => {
    setUploadPhase(0);
    if (!uploadPending) return;

    const timer = window.setTimeout(() => {
      setUploadPhase(1);
    }, 700);

    return () => {
      window.clearTimeout(timer);
    };
  }, [fileName, uploadPending]);

  let visiblePhase = 3;
  if (uploadPending) visiblePhase = uploadPhase;
  else if (analysisPending) visiblePhase = 2;
  else if (analysisError) visiblePhase = -1;

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
        "border-border bg-surface mt-5 overflow-hidden rounded-xl border-[0.5px] p-5",
        busy &&
          "border-transparent bg-[linear-gradient(100deg,rgba(102,121,248,0.07)_0%,rgba(243,90,168,0.07)_100%)]",
        analysisError && "border-danger/30 bg-danger/5",
      )}
    >
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
              "font-semibold tracking-[0.08em] uppercase",
              analysisError
                ? "text-danger"
                : "bg-[linear-gradient(100deg,#6679F8_0%,#F35AA8_100%)] bg-clip-text text-transparent",
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
            const complete = visiblePhase >= 0 && index <= visiblePhase;
            return (
              <Box key={item.label}>
                <Box className="bg-border h-1 overflow-hidden rounded-full">
                  <Box
                    className={cn(
                      "h-full origin-left scale-x-0 rounded-full bg-[linear-gradient(100deg,#6679F8_0%,#F35AA8_100%)] transition-transform duration-200 motion-reduce:transition-none",
                      complete && "scale-x-100",
                      active && "animate-pulse motion-reduce:animate-none",
                    )}
                  />
                </Box>
                <Text
                  className={cn(
                    "text-text-muted mt-1.5",
                    active && "text-foreground font-medium",
                  )}
                >
                  {item.label}
                </Text>
              </Box>
            );
          })}
        </Box>
      )}
      <Text className="mt-3 truncate" color="muted">
        {fileName} · Nothing has been imported yet
      </Text>
    </Box>
  );
};

const DestinationTeamPicker = ({
  onChange,
  teams,
  value,
}: {
  onChange: (teamId: string) => void;
  teams: Team[];
  value: string;
}) => {
  const [open, setOpen] = useState(false);
  const selectedTeam = teams.find((team) => team.id === value);

  if (teams.length === 0) {
    return (
      <Box className="border-border bg-surface mt-5 rounded-xl border-[0.5px] p-4">
        <Text className="font-medium">No existing teams</Text>
        <Text className="mt-1 leading-6" color="muted">
          Choose “Create a new team” to continue.
        </Text>
      </Box>
    );
  }

  return (
    <Box className="border-border bg-surface mt-5 rounded-xl border-[0.5px] p-4">
      <Text className="mb-2 font-medium">Destination team</Text>
      <Popover onOpenChange={setOpen} open={open}>
        <Popover.Trigger asChild>
          <Button
            align="between"
            className="h-11 w-full"
            color="tertiary"
            rightIcon={<ArrowDown2Icon className="h-4 shrink-0" />}
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
          className="z-[60] w-[var(--radix-popover-trigger-width)] p-2"
        >
          <Command>
            <Command.Input autoFocus placeholder="Search teams..." />
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
    </Box>
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
        className="bg-foreground h-full rounded-full transition-all duration-300"
        style={{ width: `${(step / STEPS.length) * 100}%` }}
      />
    </Box>
  </Box>
);

const ImportRunStep = ({
  importPending,
  importProgress,
  outcome,
  runError,
}: {
  importPending: boolean;
  importProgress: number;
  outcome: ImportOutcome | null;
  runError: string;
}) => {
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
            ? "FortyOne is checking the destination workflow and members before creating stories."
            : "FortyOne is creating stories in small, idempotent batches. Keep this window open until the result is ready."}
        </Text>
        <ProgressBar className="mt-6 h-2" progress={importProgress} />
        <Text className="mt-2 font-medium">
          {preparing ? "Preparing…" : `${importProgress}% complete`}
        </Text>
      </Box>
    );
  }

  if (outcome) {
    const successful = outcome.created + outcome.replayed;
    const allFailed = successful === 0 && outcome.failed > 0;
    const partial = successful > 0 && outcome.failed > 0;
    let outcomeTitle = "Your import is ready";
    if (allFailed) outcomeTitle = "Nothing was imported";
    else if (partial) outcomeTitle = "Import finished with issues";

    return (
      <Box aria-live="polite" className="mx-auto max-w-xl text-center">
        <Box
          className={cn(
            "mx-auto flex h-16 w-16 items-center justify-center rounded-3xl",
            outcome.failed
              ? "bg-warning/10 text-warning"
              : "bg-success/10 text-success",
          )}
        >
          {outcome.failed ? (
            <WarningIcon className="h-8" />
          ) : (
            <CheckIcon className="h-8" strokeWidth={2.5} />
          )}
        </Box>
        <Text as="h2" className="mt-5 text-2xl font-medium">
          {outcomeTitle}
        </Text>
        <Text className="mt-2 leading-6" color="muted">
          {allFailed
            ? "Created no stories"
            : `Created ${outcome.created} stories`}
          {outcome.replayed
            ? `, recognized ${outcome.replayed} previously imported stories`
            : ""}
          {outcome.failed
            ? `, and found ${outcome.failed} rows that need attention`
            : ""}
          .
        </Text>
        {outcome.failed ? (
          <Box className="bg-warning/8 mt-5 max-h-52 overflow-y-auto rounded-2xl p-4 text-left">
            <Text className="font-medium">Rows needing attention</Text>
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
        {runError || "The import could not finish."} Retrying is safe: completed
        rows use stable source IDs and will not be duplicated.
      </Text>
    </Box>
  );
};

export const ImportWizard = ({ onOpenChange, open }: ImportWizardProps) => {
  const { data: session } = useSession();
  const { data: teams = [] } = useJoinedTeams();
  const { withWorkspace, workspaceSlug } = useWorkspacePath();
  const queryClient = useQueryClient();
  const mappingEdited = useRef(false);
  const draftRef = useRef<ImportDraft | null>(null);
  const analysisGeneration = useRef(0);
  const [step, setStep] = useState(1);
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
  const [excludedRows, setExcludedRows] = useState<Set<number>>(new Set());
  const [importPending, setImportPending] = useState(false);
  const [importProgress, setImportProgress] = useState(0);
  const [outcome, setOutcome] = useState<ImportOutcome | null>(null);
  const [runError, setRunError] = useState("");
  const [createdTeamId, setCreatedTeamId] = useState<string | null>(null);
  const [reviewPage, setReviewPage] = useState(0);

  const reset = () => {
    analysisGeneration.current += 1;
    mappingEdited.current = false;
    setStep(1);
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
    setExcludedRows(new Set());
    setImportPending(false);
    setImportProgress(0);
    setOutcome(null);
    setRunError("");
    setCreatedTeamId(null);
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
    if (!responseId || !fileHash || !analysisPending) return;
    let cancelled = false;
    let pollAttempts = 0;
    let timer: ReturnType<typeof setTimeout> | undefined;

    const poll = async () => {
      try {
        pollAttempts += 1;
        const response = await pollImportAnalysis({
          fileHash,
          responseId,
          workspaceSlug,
        });
        if (cancelled) return;
        if (response.status !== "completed") {
          if (pollAttempts >= MAX_ANALYSIS_POLLS) {
            throw new Error(
              "AI analysis is taking longer than expected. You can continue with the deterministic preview or upload the file again.",
            );
          }
          timer = setTimeout(() => void poll(), 1_500);
          return;
        }

        setDraft((current) => {
          if (current?.rows.length) {
            const mapping =
              !mappingEdited.current && response.analysis.mapping
                ? response.analysis.mapping
                : current.mapping;
            return {
              ...current,
              mapping,
              summary: response.analysis.summary,
              tasks: mapping
                ? mapRowsToImportTasks(current.rows, mapping)
                : current.tasks,
              warnings: response.analysis.warnings,
            };
          }
          return {
            ...response.analysis,
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
      } catch (error) {
        if (cancelled) return;
        const message =
          error instanceof Error ? error.message : "AI analysis failed";
        setAnalysisPending(false);
        setResponseId(null);
        setDraft((current) =>
          current
            ? { ...current, warnings: [...current.warnings, message] }
            : current,
        );
        if (draftRef.current) setAnalysisNotice(message);
        else setAnalysisError(message);
      }
    };

    timer = setTimeout(() => void poll(), 700);
    return () => {
      cancelled = true;
      if (timer) clearTimeout(timer);
    };
  }, [analysisPending, fileHash, fileName, responseId, workspaceSlug]);

  const handleFile = async (file: File) => {
    const generation = analysisGeneration.current + 1;
    analysisGeneration.current = generation;
    setUploadPending(true);
    setAnalysisError("");
    setAnalysisNotice("");
    setDraft(null);
    setFileName(file.name);
    setReviewPage(0);
    mappingEdited.current = false;
    try {
      const response = await startImportAnalysis(file, workspaceSlug);
      if (generation !== analysisGeneration.current) return;
      setFileHash(response.fileHash);
      setDraft(response.analysis);
      setResponseId(response.responseId);
      setAnalysisPending(response.status === "queued");

      if (response.analysis) {
        const teamName =
          response.analysis.sourceType === "jira_csv"
            ? "Jira Import"
            : fileNameToTeamName(file.name);
        setDestination((current) =>
          current.kind === "new"
            ? {
                ...current,
                name: teamName,
                code: formatTeamCode(teamName),
              }
            : current,
        );
      }
    } catch (error) {
      if (generation !== analysisGeneration.current) return;
      setAnalysisError(
        error instanceof Error ? error.message : "Unable to analyze this file",
      );
      setAnalysisPending(false);
    } finally {
      if (generation === analysisGeneration.current) {
        setUploadPending(false);
      }
    }
  };

  const dropzone = useDropzone({
    accept: fileAccept,
    disabled: uploadPending || analysisPending,
    maxFiles: 1,
    maxSize: IMPORT_MAX_FILE_BYTES,
    multiple: false,
    onDropAccepted: ([file]) => {
      void handleFile(file);
    },
    onDropRejected: (rejections) => {
      const tooLarge = rejections.some(({ errors }) =>
        errors.some((error) => error.code === "file-too-large"),
      );
      let message = "Choose a CSV, TSV, Excel, PDF, JPG, PNG, or WebP file.";
      if (tooLarge) message = "The import file must be 20 MB or smaller.";
      setAnalysisError(message);
    },
  });

  const selectedTasks = useMemo(
    () => draft?.tasks.filter((_, index) => !excludedRows.has(index)) ?? [],
    [draft?.tasks, excludedRows],
  );
  const destinationValid =
    importDestinationSchema.safeParse(destination).success;
  let canContinue = false;
  if (step === 1) {
    canContinue =
      Boolean(draft?.tasks.length) && !analysisPending && !uploadPending;
  } else if (step === 2) {
    canContinue = destinationValid;
  } else if (step === 3) {
    canContinue =
      selectedTasks.length > 0 &&
      selectedTasks.every((task) => task.title.trim().length > 0);
  }

  const updateMapping = (field: keyof ImportMapping, selectedValue: string) => {
    const value = selectedValue === DO_NOT_IMPORT_VALUE ? null : selectedValue;
    mappingEdited.current = true;
    setDraft((current) => {
      if (!current?.mapping) return current;
      const mapping = { ...current.mapping, [field]: value };
      return {
        ...current,
        mapping,
        tasks: mapRowsToImportTasks(current.rows, mapping),
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

  const getDestinationTeam = async () => {
    if (createdTeamId) return createdTeamId;
    if (destination.kind === "existing") return destination.teamId;
    if (!session)
      throw new Error("Your session expired. Refresh and try again.");

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
    setCreatedTeamId(response.data.id);
    return response.data.id;
  };

  const startImport = async () => {
    if (!draft || !session || selectedTasks.length === 0) return;
    setStep(4);
    setAnalysisPending(false);
    setResponseId(null);
    setImportPending(true);
    setImportProgress(0);
    setRunError("");
    setOutcome(null);

    try {
      const teamId = await getDestinationTeam();
      const ctx = { session, workspaceSlug };
      const [statuses, members] = await Promise.all([
        getImportTeamStatuses(teamId, ctx),
        getImportTeamMembers(teamId, ctx),
      ]);
      const allResults: ImportStoryResult[] = [];
      let processed = 0;
      const sourceIdCounts = draft.tasks.reduce<Map<string, number>>(
        (counts, task) => {
          const sourceId = task.sourceId.trim();
          counts.set(sourceId, (counts.get(sourceId) ?? 0) + 1);
          return counts;
        },
        new Map(),
      );
      const selectedSourceIds = selectedTasks.map((task) =>
        task.sourceId.trim().toUpperCase(),
      );
      const provider =
        draft.sourceType === "jira_csv" &&
        selectedSourceIds.every((sourceId) =>
          JIRA_ISSUE_KEY_PATTERN.test(sourceId),
        ) &&
        new Set(selectedSourceIds).size === selectedSourceIds.length
          ? "jira_csv"
          : "file";
      const keyedTasks = await Promise.all(
        draft.tasks
          .map((task, taskIndex) => ({ task, taskIndex }))
          .filter(({ taskIndex }) => !excludedRows.has(taskIndex))
          .map(async ({ task, taskIndex }) => {
            const sourceId =
              provider === "jira_csv"
                ? task.sourceId.trim().toUpperCase()
                : task.sourceId.trim();
            const sourceKey = await getBoundedImportSourceKey(
              provider === "jira_csv" || sourceIdCounts.get(sourceId) === 1
                ? sourceId
                : `${sourceId}#row-${taskIndex + 1}`,
            );
            return {
              sourceKey,
              story: toImportStoryPayload({
                members,
                statuses,
                task,
                teamId,
              }),
            };
          }),
      );
      const importRequests = buildImportStoryRequests({
        items: keyedTasks,
        provider,
        sourceDigest: draft.fileHash,
      });

      for (const request of importRequests) {
        // eslint-disable-next-line no-await-in-loop -- Sequential batches cap load and make progress truthful.
        const response = await importStoriesBatch(request, ctx);
        if (response.error?.message || !response.data) {
          throw new Error(
            response.error?.message || "A batch could not be imported",
          );
        }
        allResults.push(...response.data.items);
        processed += request.items.length;
        setImportProgress(Math.round((processed / keyedTasks.length) * 100));
      }

      const created = allResults.filter((item) => item.created).length;
      const failed = allResults.filter((item) => item.error !== null).length;
      const replayed = allResults.length - created - failed;
      setOutcome({ created, failed, items: allResults, replayed, teamId });
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: storyKeys.all(workspaceSlug),
        }),
        queryClient.invalidateQueries({
          queryKey: teamKeys.all(workspaceSlug),
        }),
        queryClient.invalidateQueries({
          queryKey: statusKeys.all(workspaceSlug),
        }),
      ]);
    } catch (error) {
      setRunError(
        error instanceof Error ? error.message : "The import could not finish",
      );
    } finally {
      setImportPending(false);
    }
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
    "Upload an export from your work tool. FortyOne uses its configured OpenAI service to map every file automatically, then gives you a full review before any story is created.";
  const suggestedTeamName =
    draft?.sourceType === "jira_csv"
      ? "Jira Import"
      : fileNameToTeamName(fileName);

  let uploadLabel = "Drop a file here or choose one";
  if (uploadPending) uploadLabel = "Reading your file…";
  else if (dropzone.isDragActive) uploadLabel = "Drop it here";

  const reviewPageCount = Math.max(
    1,
    Math.ceil((draft?.tasks.length ?? 0) / REVIEW_PAGE_SIZE),
  );
  const reviewPageStart = reviewPage * REVIEW_PAGE_SIZE;
  const visibleReviewTasks =
    draft?.tasks.slice(reviewPageStart, reviewPageStart + REVIEW_PAGE_SIZE) ??
    [];

  let footerBack: ReactNode = <span />;
  if (step > 1 && step < 4) {
    footerBack = (
      <Button
        color="tertiary"
        leftIcon={<ArrowLeft2Icon />}
        onClick={() => {
          setStep((current) => Math.max(1, current - 1));
        }}
        variant="outline"
      >
        Back
      </Button>
    );
  } else if (step === 4 && !importPending && !outcome) {
    footerBack = (
      <Button
        color="tertiary"
        leftIcon={<ArrowLeft2Icon />}
        onClick={() => {
          setStep(3);
        }}
        variant="outline"
      >
        Review again
      </Button>
    );
  } else if (step === 4 && outcome?.failed) {
    footerBack = (
      <Button
        color="tertiary"
        leftIcon={<ArrowLeft2Icon />}
        onClick={() => {
          setOutcome(null);
          setStep(3);
        }}
        variant="outline"
      >
        Review rows
      </Button>
    );
  }

  let footerAction: ReactNode = null;
  if (step < 3) {
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
  } else if (step === 3) {
    footerAction = (
      <Button
        color="invert"
        disabled={!canContinue}
        onClick={() => void startImport()}
      >
        Import {selectedTasks.length} stories
      </Button>
    );
  } else if (outcome && outcome.created + outcome.replayed > 0) {
    footerAction = (
      <Button
        color="invert"
        href={withWorkspace(`/teams/${outcome.teamId}/stories`)}
        onClick={() => {
          handleOpenChange(false);
        }}
        rightIcon={<ArrowRight2Icon />}
      >
        View imported work
      </Button>
    );
  } else if (outcome) {
    footerAction = (
      <Button color="invert" onClick={() => void startImport()}>
        Retry safely
      </Button>
    );
  } else if (!importPending) {
    footerAction = (
      <Button color="invert" onClick={() => void startImport()}>
        Retry safely
      </Button>
    );
  }

  return (
    <Dialog onOpenChange={handleOpenChange} open={open}>
      <Dialog.Content
        className="mt-0 flex max-h-[calc(100dvh-2rem)] max-w-3xl flex-col md:mt-0"
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
          {step === 1 ? (
            <Box>
              <Text as="h2" className="text-xl font-medium">
                Add your export
              </Text>
              <Text className="mt-1 leading-6" color="muted">
                Export tasks or issues from Jira, ClickUp, monday.com, Asana, or
                another tool, then upload the file here. AI mapping starts
                automatically, and you can correct every suggestion before the
                import runs.
              </Text>

              {fileName ? (
                <DropZone>
                  <DropZone.Input inputProps={dropzone.getInputProps()} />
                  <ImportAnalysisBanner
                    analysisError={analysisError}
                    analysisNotice={analysisNotice}
                    analysisPending={analysisPending}
                    fileName={fileName}
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
                        CSV, Excel, PDF, JPG, PNG, or WebP, up to 20 MB
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

          {step === 2 ? (
            <Box>
              <Text as="h2" className="text-xl font-medium">
                Where should this work live?
              </Text>
              <Text className="mt-1 leading-6" color="muted">
                Import into a current team or create a focused destination
                first.
              </Text>

              <Box className="mt-5 grid gap-3 md:grid-cols-2">
                <SelectionCard
                  description="Add the imported stories to a team you already belong to."
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
                  description="Create a new team with its own workflow and story sequence."
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
            </Box>
          ) : null}

          {step === 3 && draft ? (
            <Box>
              <Text as="h2" className="text-xl font-medium">
                Review your import
              </Text>
              <Text className="mt-1 leading-6" color="muted">
                Confirm the field mapping and the stories you want to create.
              </Text>

              {draft.mapping && draft.columns.length > 0 ? (
                <Box className="border-border bg-surface mt-5 rounded-xl border-[0.5px] p-5">
                  <Text className="font-medium">Field mapping</Text>
                  <Box className="mt-4 grid gap-x-5 gap-y-3 md:grid-cols-2">
                    {MAPPING_FIELDS.map(({ key, label, required }) => (
                      <label key={key}>
                        <span className="mb-1.5 block font-medium">
                          {label}
                          {required ? (
                            <span className="text-danger"> *</span>
                          ) : null}
                        </span>
                        <Select
                          onValueChange={(value) => {
                            updateMapping(key, value);
                          }}
                          value={draft.mapping?.[key] ?? DO_NOT_IMPORT_VALUE}
                        >
                          <Select.Trigger className="bg-surface h-11 rounded-xl text-base">
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
                    ))}
                  </Box>
                </Box>
              ) : null}

              <Box className="border-border bg-surface mt-5 overflow-hidden rounded-xl border-[0.5px]">
                <Flex
                  align="center"
                  className="border-border border-b-[0.5px] px-4 py-3"
                  justify="between"
                >
                  <Text className="font-medium">Task review</Text>
                  <Text color="muted">{selectedTasks.length} selected</Text>
                </Flex>
                <Box className="divide-border max-h-96 divide-y overflow-y-auto">
                  {visibleReviewTasks.map((task, index) => {
                    const taskIndex = reviewPageStart + index;
                    const isExcluded = excludedRows.has(taskIndex);
                    return (
                      <Flex
                        align="start"
                        className={cn(
                          "gap-3 px-4 py-3",
                          isExcluded && "opacity-55",
                        )}
                        key={`${task.sourceId}-${taskIndex}`}
                      >
                        <Checkbox
                          aria-label={`Import ${task.title}`}
                          checked={!isExcluded}
                          className="mt-3"
                          onCheckedChange={(checked) => {
                            toggleTask(taskIndex, checked === true);
                          }}
                        />
                        <Box className="min-w-0 flex-1">
                          <Input
                            aria-label={`Title for ${task.sourceId}`}
                            className="h-10 text-base font-medium"
                            disabled={isExcluded}
                            maxLength={255}
                            onChange={(event) => {
                              updateTaskTitle(taskIndex, event.target.value);
                            }}
                            value={task.title}
                          />
                          {!task.title.trim() && !isExcluded ? (
                            <Text className="text-danger mt-1.5 font-medium">
                              Add a title or leave this row out.
                            </Text>
                          ) : null}
                          {task.description ? (
                            <Text
                              className="mt-2 line-clamp-1 leading-6"
                              color="muted"
                            >
                              {task.description}
                            </Text>
                          ) : null}
                          <Flex
                            className="text-text-muted mt-1.5 flex-wrap"
                            gap={3}
                          >
                            <span>{task.sourceId}</span>
                            {task.status ? <span>{task.status}</span> : null}
                            {task.priority !== "No Priority" ? (
                              <span>{task.priority}</span>
                            ) : null}
                          </Flex>
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
                    Showing {reviewPageStart + 1}–
                    {Math.min(
                      reviewPageStart + REVIEW_PAGE_SIZE,
                      draft.tasks.length,
                    )}{" "}
                    of {draft.tasks.length}
                  </Text>
                  {reviewPageCount > 1 ? (
                    <Flex className="justify-end" gap={2}>
                      <Button
                        color="tertiary"
                        disabled={reviewPage === 0}
                        onClick={() => {
                          setReviewPage((current) => Math.max(0, current - 1));
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
            </Box>
          ) : null}

          {step === 4 ? (
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
