"use client";

import type { ReactNode } from "react";
import {
  ObjectiveIcon,
  OKRIcon,
  SprintsIcon,
  StrategyIcon,
  StoryIcon,
  TagsIcon,
  TeamIcon,
  UserMultiple02Icon,
} from "icons";
import { Box, Checkbox, Flex, Input, Text } from "ui";
import { useTerminology } from "@/hooks/use-terminology-display";
import type { Team } from "@/modules/teams/public/types";
import type { ImportDraft } from "../schema";
import { resolveImportSourceTeam } from "../import-team-model";
import type {
  ObjectiveImportDestinationPreview,
  StrategicPillarImportDestinationPreview,
} from "./import-review-types";

const EntityCount = ({
  count,
  icon,
  label,
}: {
  count: number;
  icon: ReactNode;
  label: string;
}) => (
  <Flex
    align="center"
    className="border-border bg-surface min-w-0 rounded-xl border-[0.5px] px-3 py-2.5"
    gap={3}
  >
    <Box className="bg-surface-muted text-text-muted flex h-8 w-8 shrink-0 items-center justify-center rounded-lg">
      {icon}
    </Box>
    <Box className="min-w-0">
      <Text className="font-medium">{count}</Text>
      <Text className="truncate" color="muted">
        {label}
      </Text>
    </Box>
  </Flex>
);

export const ImportPlanSummary = ({ draft }: { draft: ImportDraft }) => {
  const { getTermDisplay } = useTerminology();
  const counts = [
    { count: draft.teams.length, icon: <TeamIcon />, label: "Teams" },
    {
      count: draft.strategicPillars.length,
      icon: <StrategyIcon />,
      label: "Strategic pillars",
    },
    {
      count: draft.objectives.length,
      icon: <ObjectiveIcon />,
      label: getTermDisplay("objectiveTerm", {
        capitalize: true,
        variant: "plural",
      }),
    },
    {
      count: draft.keyResults.length,
      icon: <OKRIcon />,
      label: getTermDisplay("keyResultTerm", {
        capitalize: true,
        variant: "plural",
      }),
    },
    {
      count: draft.sprints.length,
      icon: <SprintsIcon />,
      label: getTermDisplay("sprintTerm", {
        capitalize: true,
        variant: "plural",
      }),
    },
    { count: draft.labels.length, icon: <TagsIcon />, label: "Labels" },
    {
      count: draft.tasks.length,
      icon: <StoryIcon />,
      label: getTermDisplay("storyTerm", {
        capitalize: true,
        variant: "plural",
      }),
    },
    {
      count: draft.people.length,
      icon: <UserMultiple02Icon />,
      label: "People",
    },
  ].filter(({ count }) => count > 0);

  if (counts.length === 0) return null;

  return (
    <Box className="mt-5">
      <Text className="font-medium">Detected import plan</Text>
      <Text className="mt-1 leading-6" color="muted">
        FortyOne found these source objects and preserved their relationships
        for review.
      </Text>
      <Box className="mt-3 grid grid-cols-2 gap-2 md:grid-cols-4">
        {counts.map((count) => (
          <EntityCount key={count.label} {...count} />
        ))}
      </Box>
    </Box>
  );
};

export const SourceTeamPreview = ({
  conflictedSourceIds,
  draft,
  existingTeams,
  joinedTeamIds,
}: {
  conflictedSourceIds: ReadonlySet<string>;
  draft: ImportDraft;
  existingTeams: Team[];
  joinedTeamIds: ReadonlySet<string>;
}) => {
  if (draft.teams.length === 0) return null;

  return (
    <Box className="border-border bg-surface mt-5 overflow-hidden rounded-xl border-[0.5px]">
      <Flex
        align="center"
        className="border-border border-b-[0.5px] px-4 py-3"
        justify="between"
      >
        <Text className="font-medium">Source teams</Text>
        <Text color="muted">{draft.teams.length} detected</Text>
      </Flex>
      <Box className="divide-border max-h-64 divide-y overflow-y-auto">
        {draft.teams.map((sourceTeam) => {
          const resolution = resolveImportSourceTeam(sourceTeam, existingTeams);
          const needsReview =
            conflictedSourceIds.has(sourceTeam.sourceId) ||
            resolution.kind === "ambiguous";
          const existing =
            resolution.kind === "unique" && !needsReview
              ? resolution.team
              : undefined;
          let destinationLabel = "Create team";
          if (needsReview) {
            destinationLabel = "Needs review";
          } else if (existing) {
            destinationLabel = joinedTeamIds.has(existing.id)
              ? `Reuse ${existing.code}`
              : `Reuse + join ${existing.code}`;
          }
          let destinationClassName = "shrink-0";
          if (needsReview) {
            destinationClassName = "text-warning shrink-0";
          } else if (existing) {
            destinationClassName = "text-success shrink-0";
          }
          return (
            <Flex
              align="center"
              className="px-4 py-3"
              gap={3}
              justify="between"
              key={sourceTeam.sourceId}
            >
              <Box className="min-w-0">
                <Text className="truncate font-medium">{sourceTeam.name}</Text>
                {sourceTeam.description ? (
                  <Text className="mt-0.5 line-clamp-1" color="muted">
                    {sourceTeam.description}
                  </Text>
                ) : null}
              </Box>
              <Text className={destinationClassName} color="muted">
                {destinationLabel}
              </Text>
            </Flex>
          );
        })}
      </Box>
    </Box>
  );
};

export const ObjectiveImportReview = ({
  destinationMatches,
  draft,
  excludedSourceIds,
  onCheckedChange,
  onNameChange,
}: {
  destinationMatches: ReadonlyMap<string, ObjectiveImportDestinationPreview>;
  draft: ImportDraft;
  excludedSourceIds: Set<string>;
  onCheckedChange: (sourceId: string, checked: boolean) => void;
  onNameChange: (sourceId: string, name: string) => void;
}) => {
  const { getTermDisplay } = useTerminology();
  const objectiveTerm = getTermDisplay("objectiveTerm");
  const objectiveTermCapitalized = getTermDisplay("objectiveTerm", {
    capitalize: true,
  });
  const keyResultTerm = getTermDisplay("keyResultTerm");
  const keyResultTermPlural = getTermDisplay("keyResultTerm", {
    variant: "plural",
  });
  if (draft.objectives.length === 0) return null;
  const teamsBySourceId = new Map(
    draft.teams.map((team) => [team.sourceId, team.name]),
  );
  const pillarsBySourceId = new Map(
    draft.strategicPillars.map((pillar) => [pillar.sourceId, pillar.name]),
  );
  const keyResultCounts = draft.keyResults.reduce<Map<string, number>>(
    (counts, keyResult) => {
      if (!keyResult.objectiveSourceId) return counts;
      counts.set(
        keyResult.objectiveSourceId,
        (counts.get(keyResult.objectiveSourceId) ?? 0) + 1,
      );
      return counts;
    },
    new Map(),
  );

  return (
    <Box className="border-border bg-surface mt-5 overflow-hidden rounded-xl border-[0.5px]">
      <Flex
        align="center"
        className="border-border border-b-[0.5px] px-4 py-3"
        justify="between"
      >
        <Text className="font-medium">{objectiveTermCapitalized} review</Text>
        <Text color="muted">
          {draft.objectives.length - excludedSourceIds.size} selected
        </Text>
      </Flex>
      <Box className="divide-border max-h-72 divide-y overflow-y-auto">
        {draft.objectives.map((objective) => {
          const checked = !excludedSourceIds.has(objective.sourceId);
          const teamName = objective.teamSourceId
            ? teamsBySourceId.get(objective.teamSourceId)
            : null;
          const keyResultCount = keyResultCounts.get(objective.sourceId) ?? 0;
          const pillarName = objective.pillarSourceId
            ? pillarsBySourceId.get(objective.pillarSourceId)
            : null;
          const destinationMatch = destinationMatches.get(objective.sourceId);
          let destinationLabel = destinationMatch?.teamLabel
            ? `Create in ${destinationMatch.teamLabel}`
            : `Create ${objectiveTerm}`;
          if (destinationMatch?.kind === "unique") {
            destinationLabel = destinationMatch.locked
              ? `Reuse previously imported ${destinationMatch.objectiveName ?? objective.name} in ${destinationMatch.teamLabel}`
              : `Reuse ${destinationMatch.objectiveName ?? objective.name} in ${destinationMatch.teamLabel}`;
          } else if (destinationMatch?.kind === "ambiguous") {
            destinationLabel = `${destinationMatch.matchCount ?? "Multiple"} exact matches in ${destinationMatch.teamLabel}; rename this ${objectiveTerm} or resolve the duplicates`;
          } else if (destinationMatch?.kind === "privacy_conflict") {
            destinationLabel = `A same-name ${objectiveTerm} in ${destinationMatch.teamLabel} has different privacy; rename this ${objectiveTerm} to preserve privacy`;
          } else if (destinationMatch?.kind === "source_conflict") {
            destinationLabel = `Another source ${objectiveTerm} has this name in ${destinationMatch.teamLabel}; rename one before importing`;
          } else if (destinationMatch?.kind === "team_conflict") {
            destinationLabel = `Team match for ${destinationMatch.teamLabel} needs review; this ${objectiveTerm} will be skipped`;
          } else if (destinationMatch?.kind === "pillar_conflict") {
            destinationLabel =
              destinationMatch.pillarLabel ??
              `This reused ${objectiveTerm} is aligned to a different strategic pillar`;
          }
          return (
            <Flex
              align="start"
              className={checked ? "px-4 py-3" : "px-4 py-3 opacity-55"}
              gap={3}
              key={objective.sourceId}
            >
              <Checkbox
                aria-label={`Import ${objectiveTerm} ${objective.name}`}
                checked={checked}
                className="mt-1"
                onCheckedChange={(value) => {
                  onCheckedChange(objective.sourceId, value === true);
                }}
              />
              <Box className="min-w-0 flex-1">
                <Input
                  aria-label={`Name for ${objectiveTerm} ${objective.sourceId}`}
                  className="h-10 text-base font-medium"
                  disabled={!checked || destinationMatch?.locked}
                  maxLength={255}
                  onChange={(event) => {
                    onNameChange(objective.sourceId, event.target.value);
                  }}
                  value={objective.name}
                />
                {!objective.name.trim() && checked ? (
                  <Text className="text-danger mt-1.5 font-medium">
                    Add a name or leave this {objectiveTerm} out.
                  </Text>
                ) : null}
                {objective.description ? (
                  <Text className="mt-1 line-clamp-2 leading-6" color="muted">
                    {objective.description}
                  </Text>
                ) : null}
                <Flex className="mt-1.5 flex-wrap" gap={3}>
                  {teamName ? <Text color="muted">{teamName}</Text> : null}
                  {objective.status ? (
                    <Text color="muted">{objective.status}</Text>
                  ) : null}
                  {objective.isPrivate ? (
                    <Text color="muted">Private</Text>
                  ) : null}
                  {pillarName ? (
                    <Text color="muted">Pillar: {pillarName}</Text>
                  ) : null}
                  {keyResultCount ? (
                    <Text color="muted">
                      {keyResultCount}{" "}
                      {keyResultCount === 1
                        ? keyResultTerm
                        : keyResultTermPlural}
                    </Text>
                  ) : null}
                </Flex>
                <Text
                  className={
                    destinationMatch?.kind === "ambiguous" ||
                    destinationMatch?.kind === "pillar_conflict" ||
                    destinationMatch?.kind === "privacy_conflict" ||
                    destinationMatch?.kind === "source_conflict" ||
                    destinationMatch?.kind === "team_conflict"
                      ? "text-warning mt-2"
                      : "text-success mt-2"
                  }
                  color="muted"
                >
                  {destinationLabel}
                </Text>
                {destinationMatch?.pillarLabel &&
                destinationMatch.kind !== "pillar_conflict" ? (
                  <Text className="text-warning mt-1" color="muted">
                    {destinationMatch.pillarLabel}
                  </Text>
                ) : null}
              </Box>
            </Flex>
          );
        })}
      </Box>
    </Box>
  );
};

export const StrategicPillarImportReview = ({
  destinationMatches,
  draft,
  excludedSourceIds,
  onCheckedChange,
  onNameChange,
}: {
  destinationMatches: ReadonlyMap<
    string,
    StrategicPillarImportDestinationPreview
  >;
  draft: ImportDraft;
  excludedSourceIds: Set<string>;
  onCheckedChange: (sourceId: string, checked: boolean) => void;
  onNameChange: (sourceId: string, name: string) => void;
}) => {
  const { getTermDisplay } = useTerminology();
  const objectiveTerm = getTermDisplay("objectiveTerm");
  const objectiveTermPlural = getTermDisplay("objectiveTerm", {
    variant: "plural",
  });
  if (draft.strategicPillars.length === 0) return null;
  const objectiveCounts = draft.objectives.reduce<Map<string, number>>(
    (counts, objective) => {
      if (!objective.pillarSourceId) return counts;
      counts.set(
        objective.pillarSourceId,
        (counts.get(objective.pillarSourceId) ?? 0) + 1,
      );
      return counts;
    },
    new Map(),
  );

  return (
    <Box className="border-border bg-surface mt-5 overflow-hidden rounded-xl border-[0.5px]">
      <Flex
        align="center"
        className="border-border border-b-[0.5px] px-4 py-3"
        justify="between"
      >
        <Text className="font-medium">Strategic pillar review</Text>
        <Text color="muted">
          {draft.strategicPillars.length - excludedSourceIds.size} selected
        </Text>
      </Flex>
      <Box className="divide-border max-h-72 divide-y overflow-y-auto">
        {draft.strategicPillars.map((pillar) => {
          const checked = !excludedSourceIds.has(pillar.sourceId);
          const destinationMatch = destinationMatches.get(pillar.sourceId);
          const objectiveCount = objectiveCounts.get(pillar.sourceId) ?? 0;
          let destinationLabel = "Create strategic pillar";
          if (destinationMatch?.kind === "unique") {
            destinationLabel = `Reuse ${destinationMatch.pillarName ?? pillar.name}`;
          } else if (destinationMatch?.kind === "ambiguous") {
            destinationLabel = `${destinationMatch.matchCount ?? "Multiple"} exact workspace matches; rename this pillar or resolve the duplicates`;
          } else if (destinationMatch?.kind === "source_conflict") {
            destinationLabel =
              "Another selected source pillar has this name; rename one";
          }
          const hasConflict =
            destinationMatch?.kind === "ambiguous" ||
            destinationMatch?.kind === "source_conflict";
          return (
            <Flex
              align="start"
              className={checked ? "px-4 py-3" : "px-4 py-3 opacity-55"}
              gap={3}
              key={pillar.sourceId}
            >
              <Checkbox
                aria-label={`Import strategic pillar ${pillar.name}`}
                checked={checked}
                className="mt-1"
                onCheckedChange={(value) => {
                  onCheckedChange(pillar.sourceId, value === true);
                }}
              />
              <Box className="min-w-0 flex-1">
                <Input
                  aria-label={`Name for strategic pillar ${pillar.sourceId}`}
                  className="h-10 text-base font-medium"
                  disabled={!checked}
                  maxLength={255}
                  onChange={(event) => {
                    onNameChange(pillar.sourceId, event.target.value);
                  }}
                  value={pillar.name}
                />
                {!pillar.name.trim() && checked ? (
                  <Text className="text-danger mt-1.5 font-medium">
                    Add a name or leave this pillar out.
                  </Text>
                ) : null}
                {pillar.description ? (
                  <Text className="mt-1 line-clamp-2 leading-6" color="muted">
                    {pillar.description}
                  </Text>
                ) : null}
                <Flex className="mt-1.5 flex-wrap" gap={3}>
                  <Text color="muted">Order {pillar.orderIndex + 1}</Text>
                  {objectiveCount ? (
                    <Text color="muted">
                      {objectiveCount} aligned{" "}
                      {objectiveCount === 1
                        ? objectiveTerm
                        : objectiveTermPlural}
                    </Text>
                  ) : null}
                </Flex>
                <Text
                  className={
                    hasConflict ? "text-warning mt-2" : "text-success mt-2"
                  }
                  color="muted"
                >
                  {destinationLabel}
                </Text>
              </Box>
            </Flex>
          );
        })}
      </Box>
    </Box>
  );
};
