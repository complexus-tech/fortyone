"use client";
import type { ReactNode } from "react";
import { useEffect, useState } from "react";
import { CheckIcon } from "icons";
import { cn } from "lib";
import { Box, Button, Flex, Text } from "ui";
import { Thinking } from "@/components/ui/chat/thinking";

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

export const ImportAnalysisBanner = ({
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
