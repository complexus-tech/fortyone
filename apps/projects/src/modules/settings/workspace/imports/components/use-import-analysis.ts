"use client";
import { useEffect, useRef, useState } from "react";
import type { ImportAnalysis, ImportDraft, ImportMapping } from "../schema";
import { pollImportAnalysis, startImportAnalysis } from "../api";
import { mapRowsToImportTasks } from "../csv";
import { mergeAnalyzedTaskGraph } from "../task-graph";
import {
  isImportMappingFieldLocked,
  prepareCompletedAIImportAnalysis,
} from "./import-wizard-safety";
import {
  DO_NOT_IMPORT_VALUE,
  mergeCompletedImportAnalysis,
} from "./import-draft-model";
import { useImportTerms } from "./use-import-terms";

const IMPORT_ANALYSIS_POLL_TIMEOUT_MS = 7 * 60 * 1000;
const IMPORT_ANALYSIS_INITIAL_POLL_DELAY_MS = 700;
const IMPORT_ANALYSIS_MIN_POLL_DELAY_MS = 1_500;
const IMPORT_ANALYSIS_MAX_POLL_DELAY_MS = 10_000;
const IMPORT_ANALYSIS_POLL_ERROR_RETRIES = 3;
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

type ImportAnalysisOptions = {
  workspaceSlug: string;
  onNewFile: () => void;
  onUploaded: (draft: ImportDraft | null, fileName: string) => void;
  onCompleted: (analysis: ImportAnalysis) => void;
};
export const useImportAnalysis = ({
  workspaceSlug,
  onNewFile,
  onUploaded,
  onCompleted,
}: ImportAnalysisOptions) => {
  const { storyTerm, storyTermCapitalized } = useImportTerms();
  const mappingEdited = useRef(false);
  const mappingOverrideFields = useRef(new Set<keyof ImportMapping>());
  const draftRef = useRef<ImportDraft | null>(null);
  const analysisGeneration = useRef(0);
  const analysisPollingSession = useRef<{
    responseId: string;
    startedAt: number;
  } | null>(null);
  const [fileName, setFileName] = useState("");
  const [draft, setDraft] = useState<ImportDraft | null>(null);
  const [responseId, setResponseId] = useState<string | null>(null);
  const [fileHash, setFileHash] = useState("");
  const [analysisPending, setAnalysisPending] = useState(false);
  const [analysisError, setAnalysisError] = useState("");
  const [analysisNotice, setAnalysisNotice] = useState("");
  const [uploadPending, setUploadPending] = useState(false);
  const onAnalysisReady = useRef(onCompleted);
  useEffect(() => {
    draftRef.current = draft;
    onAnalysisReady.current = onCompleted;
  }, [draft, onCompleted]);
  const reset = () => {
    analysisGeneration.current += 1;
    analysisPollingSession.current = null;
    mappingEdited.current = false;
    mappingOverrideFields.current.clear();
    setFileName("");
    setDraft(null);
    setResponseId(null);
    setFileHash("");
    setAnalysisPending(false);
    setAnalysisError("");
    setAnalysisNotice("");
    setUploadPending(false);
  };
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
          setDraft((current) =>
            mergeCompletedImportAnalysis(current, completedAnalysis, {
              fileHash,
              fileName,
              mappingEdited: mappingEdited.current,
              mappingOverrideFields: mappingOverrideFields.current,
            }),
          );
          setAnalysisPending(false);
          setAnalysisNotice("");
          setResponseId(null);
          onAnalysisReady.current(completedAnalysis);
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
  }, [analysisPending, fileHash, fileName, responseId, workspaceSlug]);

  const handleFile = (file: File) => {
    reset();
    onNewFile();
    const generation = analysisGeneration.current;
    setUploadPending(true);
    setFileName(file.name);
    const completeUpload = () => {
      if (generation === analysisGeneration.current) setUploadPending(false);
    };
    void startImportAnalysis(file, workspaceSlug)
      .then((response) => {
        if (generation !== analysisGeneration.current) return;
        setFileHash(response.fileHash);
        setDraft(response.analysis);
        setResponseId(response.responseId);
        setAnalysisPending(response.status === "queued");
        onUploaded(response.analysis, file.name);
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
  const stop = () => {
    setAnalysisPending(false);
    setResponseId(null);
  };
  const updateMapping = (
    field: keyof ImportMapping,
    selectedValue: string,
    hasAttemptedImport: boolean,
  ) => {
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
    return true;
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

  return {
    draft,
    fileHash,
    fileName,
    analysisPending,
    analysisError,
    analysisNotice,
    uploadPending,
    handleFile,
    reset,
    stop,
    setAnalysisError,
    updateMapping,
    updateTaskTitle,
    updateObjectiveName,
    updateStrategicPillarName,
  };
};
