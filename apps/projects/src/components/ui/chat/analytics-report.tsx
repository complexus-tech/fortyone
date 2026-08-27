"use client";

import type { ReactNode } from "react";
import { format } from "date-fns";
import { cn } from "lib";
import { Box, Button, Flex, Text } from "ui";
import { useTheme } from "next-themes";
import type { TooltipProps } from "recharts";
import {
  Bar,
  BarChart,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip as ChartTooltip,
  XAxis,
  YAxis,
} from "recharts";
import { useTerminology } from "@/hooks";
import { BurndownChart } from "@/modules/sprints/stories/burndown";
import type { SprintAnalytics as SingleSprintAnalytics } from "@/modules/sprints/types";

type ChartRow = Record<string, string | number | null | undefined>;

type Metric = {
  label: string;
  value: string | number;
};

const COLORS = {
  primary: "#6366F1",
  success: "#22C55E",
  warning: "#F59E0B",
  muted: "#94A3B8",
};

const COMPACT_CHART_HEIGHT = 190;

const isRecord = (value: unknown): value is Record<string, unknown> =>
  Boolean(value) && typeof value === "object" && !Array.isArray(value);

const asRows = (value: unknown): ChartRow[] =>
  Array.isArray(value) ? (value as ChartRow[]) : [];

const asRecord = (value: unknown): Record<string, unknown> =>
  isRecord(value) ? value : {};

const MetricGrid = ({ metrics }: { metrics: Metric[] }) => {
  const rowCount = Math.ceil(metrics.length / 2);

  return (
    <Box className="border-border/40 grid grid-cols-2 overflow-hidden rounded-lg border dark:border-white/[0.07]">
      {metrics.map((metric, index) => {
        const rowIndex = Math.floor(index / 2);

        return (
          <Flex
            align="center"
            className={cn(
              "border-border/40 bg-surface-muted/10 min-h-10 min-w-0 gap-2 px-2.5 py-2 dark:border-white/[0.07] dark:bg-white/[0.01]",
              index % 2 === 1 && "border-l",
              rowIndex < rowCount - 1 && "border-b",
            )}
            justify="between"
            key={metric.label}
          >
            <Text className="text-foreground/55 min-w-0 truncate text-[0.95rem] font-medium dark:text-white/45">
              {metric.label}
            </Text>
            <Text className="text-foreground shrink-0 text-base font-semibold dark:text-white">
              {metric.value}
            </Text>
          </Flex>
        );
      })}
    </Box>
  );
};

const EmptyChart = () => (
  <Box className="text-foreground/60 bg-surface-muted/30 flex h-24 items-center justify-center rounded-lg text-base dark:bg-white/[0.02] dark:text-white/55">
    No chart data available
  </Box>
);

const ChartSection = ({
  title,
  children,
  description,
}: {
  title: string;
  children: ReactNode;
  description?: string;
}) => (
  <Box className="space-y-2">
    <Box>
      <Text className="text-foreground font-semibold dark:text-white">
        {title}
      </Text>
      {description ? (
        <Text className="mt-1 text-base leading-6" color="muted">
          {description}
        </Text>
      ) : null}
    </Box>
    {children}
  </Box>
);

const KeyValueList = ({
  rows,
}: {
  rows: {
    key?: string;
    label: string;
    value: string | number | null | undefined;
  }[];
}) => (
  <Box className="border-border/40 divide-border/50 divide-y overflow-hidden rounded-lg border dark:divide-white/[0.07] dark:border-white/[0.07]">
    {rows.map((row) => (
      <Flex
        align="center"
        className="bg-surface-muted/10 gap-3 px-3 py-2 text-base dark:bg-white/[0.01]"
        justify="between"
        key={row.key ?? row.label}
      >
        <Text className="text-foreground/55 font-medium dark:text-white/45">
          {row.label}
        </Text>
        <Text className="text-foreground text-right font-medium dark:text-white">
          {row.value === null || row.value === undefined || row.value === ""
            ? "Not set"
            : row.value}
        </Text>
      </Flex>
    ))}
  </Box>
);

const PillList = ({
  emptyText,
  items,
}: {
  emptyText: string;
  items: string[];
}) => {
  if (!items.length) {
    return (
      <Text className="text-foreground/60 bg-surface-muted/30 rounded-lg px-3 py-2 text-base dark:bg-white/[0.02] dark:text-white/55">
        {emptyText}
      </Text>
    );
  }

  return (
    <Flex className="flex-wrap gap-1.5">
      {items.map((item) => (
        <Text
          className="border-border/40 bg-surface-muted/20 text-foreground/90 rounded-md border px-2 py-1 text-[0.95rem] font-medium dark:border-white/[0.07] dark:bg-white/[0.02] dark:text-white/85"
          key={item}
        >
          {item}
        </Text>
      ))}
    </Flex>
  );
};

const CompactBarChart = ({
  bars,
  data,
  maxLabelLength = 15,
  xKey,
}: {
  bars: { key: string; color: string; name?: string }[];
  data: ChartRow[];
  maxLabelLength?: number;
  xKey: string;
}) => {
  const { resolvedTheme } = useTheme();
  if (!data.length) return <EmptyChart />;

  const isDark = resolvedTheme === "dark";
  const axisColor = isDark
    ? "rgba(255, 255, 255, 0.09)"
    : "rgba(15, 23, 42, 0.12)";
  const tickColor = isDark ? "#A1A1AA" : "#64748B";

  return (
    <ResponsiveContainer height={COMPACT_CHART_HEIGHT} width="100%">
      <BarChart data={data} margin={{ top: 6, right: 4, left: -16, bottom: 0 }}>
        <XAxis
          axisLine={{ stroke: axisColor }}
          dataKey={xKey}
          minTickGap={12}
          tick={{ fill: tickColor, fontSize: 13 }}
          tickFormatter={(value) => formatCategoryTick(value, maxLabelLength)}
          tickLine={{ stroke: axisColor }}
        />
        <YAxis
          axisLine={false}
          tick={{ fill: tickColor, fontSize: 13 }}
          tickLine={false}
          width={36}
        />
        <ChartTooltip
          allowEscapeViewBox={{ x: false, y: true }}
          content={<AnalyticsChartTooltip />}
          cursor={{
            fill: isDark
              ? "rgba(255, 255, 255, 0.035)"
              : "rgba(15, 23, 42, 0.035)",
          }}
          wrapperStyle={{ outline: "none" }}
        />
        {bars.map((bar) => (
          <Bar
            activeBar={{
              fillOpacity: 0.82,
              stroke: isDark
                ? "rgba(255, 255, 255, 0.3)"
                : "rgba(15, 23, 42, 0.18)",
              strokeWidth: 1,
            }}
            dataKey={bar.key}
            fill={bar.color}
            key={bar.key}
            maxBarSize={36}
            name={bar.name}
            radius={[4, 4, 1, 1]}
          />
        ))}
      </BarChart>
    </ResponsiveContainer>
  );
};

const CompactLineChart = ({
  data,
  xKey,
  lines,
}: {
  data: ChartRow[];
  xKey: string;
  lines: { key: string; color: string; name?: string }[];
}) => {
  const { resolvedTheme } = useTheme();
  if (!data.length) return <EmptyChart />;

  const isDark = resolvedTheme === "dark";
  const axisColor = isDark
    ? "rgba(255, 255, 255, 0.09)"
    : "rgba(15, 23, 42, 0.12)";
  const tickColor = isDark ? "#A1A1AA" : "#64748B";
  const isDateAxis = xKey.toLowerCase().includes("date");

  return (
    <ResponsiveContainer height={COMPACT_CHART_HEIGHT} width="100%">
      <LineChart
        data={data}
        margin={{ top: 6, right: 4, left: -16, bottom: 0 }}
      >
        <XAxis
          axisLine={{ stroke: axisColor }}
          dataKey={xKey}
          minTickGap={16}
          tick={{ fill: tickColor, fontSize: 13 }}
          tickFormatter={isDateAxis ? formatChartDate : formatCategoryTick}
          tickLine={{ stroke: axisColor }}
        />
        <YAxis
          axisLine={false}
          tick={{ fill: tickColor, fontSize: 13 }}
          tickLine={false}
          width={36}
        />
        <ChartTooltip
          allowEscapeViewBox={{ x: false, y: true }}
          content={
            <AnalyticsChartTooltip
              formatLabel={isDateAxis ? formatChartDate : undefined}
            />
          }
          cursor={{
            stroke: isDark
              ? "rgba(255, 255, 255, 0.16)"
              : "rgba(15, 23, 42, 0.18)",
            strokeDasharray: "3 4",
            strokeWidth: 1,
          }}
          wrapperStyle={{ outline: "none" }}
        />
        {lines.map((line) => (
          <Line
            activeDot={{
              fill: line.color,
              r: 6,
              stroke: isDark ? "#18181B" : "#FFFFFF",
              strokeWidth: 3,
            }}
            dataKey={line.key}
            dot={false}
            key={line.key}
            name={line.name}
            stroke={line.color}
            strokeWidth={2}
            type="monotone"
          />
        ))}
      </LineChart>
    </ResponsiveContainer>
  );
};

const formatChartDate = (value: unknown) => {
  const date = new Date(String(value));
  return Number.isNaN(date.getTime()) ? String(value) : format(date, "MMM d");
};

const formatCategoryTick = (value: unknown, maxLength = 15) => {
  const label = String(value ?? "");
  return label.length > maxLength ? `${label.slice(0, maxLength - 1)}…` : label;
};

const AnalyticsChartTooltip = ({
  active,
  formatLabel,
  label,
  payload,
}: TooltipProps<number, string> & {
  formatLabel?: (value: unknown) => string;
}) => {
  if (!active || !payload?.length) return null;

  return (
    <Box className="border-border/40 bg-surface-elevated/85 text-foreground z-50 min-w-40 rounded-2xl border-[0.5px] p-4 shadow-lg shadow-black/5 backdrop-blur-xl dark:border-white/[0.08] dark:shadow-black/20">
      <Text className="text-base font-semibold">
        {formatLabel ? formatLabel(label) : String(label ?? "")}
      </Text>
      <Box className="mt-2 space-y-1.5">
        {payload.map((entry) => (
          <Flex
            align="center"
            className="gap-2 text-[0.95rem] font-medium"
            key={String(entry.dataKey)}
          >
            <span
              aria-hidden
              className="size-2.5 shrink-0 rounded-full"
              style={{ backgroundColor: entry.color }}
            />
            <Text className="text-foreground/75 dark:text-white/70">
              {entry.name}:{" "}
              <span className="text-foreground dark:text-white">
                {entry.value}
              </span>
            </Text>
          </Flex>
        ))}
      </Box>
    </Box>
  );
};

const asSingleSprintBurndown = (
  value: unknown,
): SingleSprintAnalytics["burndown"] =>
  asRows(value).flatMap((row) => {
    const date = typeof row.date === "string" ? row.date : "";
    const ideal = Number(row.ideal);
    const remaining = Number(row.remaining);
    const parsedDate = new Date(date);

    if (
      !date ||
      Number.isNaN(parsedDate.getTime()) ||
      !Number.isFinite(ideal) ||
      !Number.isFinite(remaining)
    ) {
      return [];
    }

    return [{ date, ideal, remaining }];
  });

const asWorkingDays = (
  value: unknown,
): SingleSprintAnalytics["workingDays"] | undefined => {
  if (!Array.isArray(value)) return undefined;

  const workingDays = value.filter(
    (day): day is number =>
      typeof day === "number" && Number.isInteger(day) && day >= 1 && day <= 7,
  );

  return workingDays.length ? workingDays : undefined;
};

const completionRate = (completed: unknown, total: unknown) => {
  const completedNumber = Number(completed ?? 0);
  const totalNumber = Number(total ?? 0);
  if (!totalNumber) return "0%";
  return `${Math.round((completedNumber / totalNumber) * 100)}%`;
};

const ratioPercent = (value: unknown) =>
  `${Math.round(Number(value ?? 0) * 100)}%`;

const hasPositiveMetric = (record: Record<string, unknown>, keys: string[]) =>
  keys.some((key) => Number(record[key] ?? 0) > 0);

const humanizeLabel = (value: unknown) =>
  String(value ?? "")
    .replaceAll(/[_-]+/g, " ")
    .replaceAll(/\b\w/g, (character) => character.toUpperCase());

const progressSummary = (completed: unknown, total: unknown) =>
  `${completionRate(completed, total)} · ${Number(completed ?? 0)} of ${Number(
    total ?? 0,
  )}`;

const HealthAndProgressSection = ({
  description,
  healthRows,
  progressRows,
  title,
}: {
  description: string;
  healthRows: {
    key?: string;
    label: string;
    value: string | number | null | undefined;
  }[];
  progressRows: {
    key?: string;
    label: string;
    value: string | number | null | undefined;
  }[];
  title: string;
}) => (
  <ChartSection description={description} title={title}>
    <Box className="space-y-3">
      <KeyValueList rows={healthRows} />
      {progressRows.length ? (
        <Box className="space-y-2">
          <Text className="text-foreground/55 text-base font-medium dark:text-white/45">
            Progress
          </Text>
          <KeyValueList rows={progressRows} />
        </Box>
      ) : null}
    </Box>
  </ChartSection>
);

const RiskList = ({ risks }: { risks: ChartRow[] }) => {
  if (!risks.length) {
    return (
      <Text className="text-text-muted border-border/50 border-y py-2 text-base dark:border-white/[0.07]">
        No active risks found.
      </Text>
    );
  }

  return (
    <Box className="border-border/50 divide-border/50 divide-y border-y dark:divide-white/[0.07] dark:border-white/[0.07]">
      {risks.slice(0, 5).map((risk) => {
        const severity = String(risk.severity ?? "low");
        let tone = "bg-info";
        if (severity === "high") tone = "bg-danger";
        if (severity === "medium") tone = "bg-warning";

        return (
          <Flex
            align="center"
            className="min-w-0 gap-3 py-2 text-base"
            justify="between"
            key={String(
              risk.kind ??
                `${String(risk.title ?? "Risk")}-${String(risk.severity ?? "low")}`,
            )}
          >
            <span className={`size-2.5 shrink-0 rounded-[2px] ${tone}`} />
            <Text className="min-w-0 flex-1 truncate">
              {String(risk.title ?? "Risk")}
            </Text>
            <Text className="text-text-muted shrink-0">
              {Number(risk.count ?? 0)}
            </Text>
          </Flex>
        );
      })}
    </Box>
  );
};

export const AnalyticsReport = ({
  output,
}: {
  output: Record<string, unknown>;
}) => {
  const { getTermDisplay } = useTerminology();
  const storyTerm = getTermDisplay("storyTerm");
  const storyTermPlural = getTermDisplay("storyTerm", {
    capitalize: true,
    variant: "plural",
  });
  const kind = output.kind;
  const title = String(
    output.title ??
      (kind === "workload-analysis-report"
        ? "Workload analysis"
        : "Performance report"),
  );

  if (kind === "github-integration-report") {
    const summary = asRecord(output.summary);
    const settings = asRecord(output.settings);
    const repositories = asRows(output.repositories);
    const issueSyncLinks = asRows(output.issueSyncLinks);
    const installations = asRows(output.installations);
    const connected = Boolean(summary.connected);

    return (
      <Box className="mt-3 space-y-3">
        <Flex align="center" className="gap-3" justify="between">
          <Box>
            <Text className="text-foreground text-xl font-semibold dark:text-white">
              {title}
            </Text>
            <Text className="text-foreground/60 text-base dark:text-white/55">
              {connected
                ? "GitHub is connected to this workspace."
                : "GitHub is not connected to this workspace."}
            </Text>
          </Box>
          <Flex align="center" className="shrink-0 gap-1.5">
            <span
              aria-hidden
              className={cn(
                "size-2.5 shrink-0 rounded-[2px]",
                connected ? "bg-success" : "bg-warning",
              )}
            />
            <Text className="text-base font-semibold">
              {connected ? "Connected" : "Setup needed"}
            </Text>
          </Flex>
        </Flex>

        <MetricGrid
          metrics={[
            {
              label: "Repositories",
              value: Number(summary.repositories ?? 0),
            },
            {
              label: "Active repos",
              value: Number(summary.activeRepositories ?? 0),
            },
            {
              label: "Sync links",
              value: Number(summary.issueSyncLinks ?? 0),
            },
            {
              label: "Installations",
              value: Number(summary.installations ?? 0),
            },
          ]}
        />

        <ChartSection title="Workspace settings">
          <KeyValueList
            rows={[
              {
                label: "Branch format",
                value: String(settings.branchFormat ?? ""),
              },
              {
                label: "Magic word linking",
                value: settings.linkCommitsByMagicWords ? "On" : "Off",
              },
              {
                label: "Assignee sync",
                value: settings.syncAssignees ? "On" : "Off",
              },
              {
                label: "Label sync",
                value: settings.syncLabels ? "On" : "Off",
              },
            ]}
          />
        </ChartSection>

        <ChartSection title="Repositories">
          <PillList
            emptyText="No repositories have been synced yet."
            items={repositories
              .slice(0, 8)
              .map((repository) => String(repository.fullName ?? ""))}
          />
        </ChartSection>

        <ChartSection title="Issue sync">
          <PillList
            emptyText="No issue sync links are configured yet."
            items={issueSyncLinks
              .slice(0, 8)
              .map(
                (link) =>
                  `${String(link.repositoryName ?? "Repository")} -> ${String(
                    link.teamName ?? "Team",
                  )}`,
              )}
          />
        </ChartSection>

        {!installations.length ? (
          <Text className="text-foreground/60 text-base dark:text-white/55">
            Ask Maya to connect GitHub to generate an installation link.
          </Text>
        ) : null}
      </Box>
    );
  }

  if (kind === "github-team-automation-report") {
    const team = asRecord(output.team);
    const rules = asRows(output.rules);

    return (
      <Box className="mt-3 space-y-3">
        <Box>
          <Text className="text-foreground text-xl font-semibold dark:text-white">
            {title}
          </Text>
          <Text className="text-foreground/60 text-base dark:text-white/55">
            Automation rules for {String(team.name ?? "this team")}.
          </Text>
        </Box>
        <KeyValueList
          rows={[
            { label: "Team", value: String(team.name ?? "") },
            { label: "Code", value: String(team.code ?? "") },
            { label: "Rules", value: rules.length },
            {
              label: "Active rules",
              value: rules.filter((rule) => Boolean(rule.isActive)).length,
            },
          ]}
        />
        <ChartSection title="Rules">
          <PillList
            emptyText="No GitHub automation rules are configured for this team."
            items={rules.map(
              (rule) =>
                `${String(rule.eventKey ?? "GitHub event")} -> ${
                  rule.targetStatusId ? "status mapped" : "no target status"
                }`,
            )}
          />
        </ChartSection>
      </Box>
    );
  }

  if (kind === "github-story-report") {
    const story = asRecord(output.story);
    const links = asRows(output.links);

    return (
      <Box className="mt-3 space-y-3">
        <Box>
          <Text className="text-foreground text-xl font-semibold dark:text-white">
            {title}
          </Text>
          <Text className="text-foreground/60 text-base dark:text-white/55">
            GitHub links attached to {String(story.ref ?? `this ${storyTerm}`)}.
          </Text>
        </Box>

        {!links.length ? (
          <Text className="text-foreground/60 bg-surface-muted/30 rounded-lg px-3 py-2 text-base dark:bg-white/[0.02] dark:text-white/55">
            No GitHub links are attached to this {storyTerm}.
          </Text>
        ) : (
          <Box className="border-border/70 divide-border/70 divide-y overflow-hidden rounded-lg border dark:divide-white/12 dark:border-white/12">
            {links.map((link) => (
              <Flex
                align="center"
                className="bg-surface-muted/20 min-w-0 gap-3 px-3 py-2 dark:bg-white/[0.02]"
                justify="between"
                key={String(link.id)}
              >
                <Box className="min-w-0">
                  <Text className="text-foreground block truncate font-semibold dark:text-white">
                    {String(link.title ?? link.refName ?? "GitHub link")}
                  </Text>
                  <Text className="text-foreground/60 block truncate text-[0.95rem] dark:text-white/55">
                    {String(link.repositoryFullName ?? "")}
                    {link.number ? ` #${String(link.number)}` : ""}
                  </Text>
                </Box>
                {typeof link.url === "string" && link.url ? (
                  <Button
                    color="black"
                    href={link.url}
                    size="sm"
                    target="_blank"
                  >
                    Open
                  </Button>
                ) : null}
              </Flex>
            ))}
          </Box>
        )}
      </Box>
    );
  }

  if (kind === "workspace-performance-report") {
    const overview = asRecord(output.overview);
    const metrics = asRecord(overview.metrics);

    return (
      <Box className="mt-3 space-y-3">
        <Text className="text-xl font-semibold">{title}</Text>
        <MetricGrid
          metrics={[
            {
              label: storyTermPlural,
              value: Number(metrics.totalStories ?? 0),
            },
            {
              label: "Completed",
              value: Number(metrics.completedStories ?? 0),
            },
            {
              label: "Completion",
              value: completionRate(
                metrics.completedStories,
                metrics.totalStories,
              ),
            },
            {
              label: "Objectives",
              value: Number(metrics.activeObjectives ?? 0),
            },
            { label: "Sprints", value: Number(metrics.activeSprints ?? 0) },
            { label: "Members", value: Number(metrics.totalTeamMembers ?? 0) },
          ]}
        />
        <ChartSection title="Completion trend">
          <CompactLineChart
            data={asRows(overview.completionTrend)}
            lines={[
              { key: "completed", color: COLORS.success, name: "Completed" },
              { key: "total", color: COLORS.primary, name: "Total" },
            ]}
            xKey="date"
          />
        </ChartSection>
      </Box>
    );
  }

  if (kind === "workspace-command-center-report") {
    const report = asRecord(output.report);
    const overview = asRecord(report.overview);
    const metrics = asRecord(overview.metrics);
    const pulse = asRecord(report.pulse);
    const pulseSummary = asRecord(pulse.summary);
    const storyHealth = asRecord(pulse.stories);
    const sprintHealth = asRecord(pulse.sprints);
    const objectiveHealth = asRecord(pulse.objectives);
    const workload = asRecord(report.workload);
    const workloadSummary = asRecord(workload.summary);
    const sprints = asRecord(report.sprints);
    const objectives = asRecord(report.objectives);
    const requests = asRecord(report.requests);
    const engagement = asRecord(report.engagement);
    const members = [...asRows(workload.members)]
      .sort(
        (firstMember, secondMember) =>
          Number(secondMember.openStories ?? 0) -
          Number(firstMember.openStories ?? 0),
      )
      .slice(0, 8)
      .map((member) => ({
        ...member,
        memberName: String(member.fullName ?? member.username ?? "Team member"),
      }));
    const providers = asRows(requests.providers);
    const risks = asRows(pulse.risks);
    const sprintProgress = asRows(sprints.sprintProgress).filter(
      (sprint) => typeof sprint.sprintName === "string" && sprint.sprintName,
    );
    const objectiveProgress = asRows(objectives.keyResultsProgress).filter(
      (objective) =>
        typeof objective.objectiveName === "string" && objective.objectiveName,
    );
    const engagementEvents = asRows(engagement.eventsByName);
    const unavailableSections = asRows(report.sectionErrors).flatMap(
      (sectionError) => {
        const section = humanizeLabel(sectionError.section);
        return section ? [section] : [];
      },
    );
    const hasWorkload =
      members.length > 0 ||
      hasPositiveMetric(workloadSummary, [
        "totalOpenStories",
        "totalEstimate",
        "unassignedStories",
        "unestimatedStories",
      ]);
    const hasSprintData =
      sprintProgress.length > 0 ||
      hasPositiveMetric(sprintHealth, [
        "activeSprints",
        "upcomingSprints",
        "completedSprints",
        "atRiskSprints",
        "overdueSprints",
      ]);
    const hasObjectiveData =
      objectiveProgress.length > 0 ||
      hasPositiveMetric(objectiveHealth, [
        "activeObjectives",
        "atRiskObjectives",
        "offTrackObjectives",
        "overdueObjectives",
        "objectivesDueSoon",
      ]);
    const hasRequestData =
      providers.length > 0 ||
      hasPositiveMetric(requests, [
        "totalRequests",
        "pendingRequests",
        "acceptedRequests",
        "declinedRequests",
      ]);
    const hasEngagementData =
      engagementEvents.length > 0 ||
      hasPositiveMetric(engagement, ["totalEvents", "uniqueUsers"]);

    return (
      <Box className="mt-3 space-y-3">
        <Text className="text-xl font-semibold">{title}</Text>
        {unavailableSections.length ? (
          <Text className="border-warning/25 bg-warning/5 text-foreground/70 dark:bg-warning/10 rounded-lg border px-3 py-2 text-base dark:text-white/65">
            Some analytics are temporarily unavailable:{" "}
            {unavailableSections.join(", ")}.
          </Text>
        ) : null}
        <MetricGrid
          metrics={[
            {
              label: "Open work",
              value: Number(workloadSummary.totalOpenStories ?? 0),
            },
            {
              label: "Completion",
              value: completionRate(
                metrics.completedStories,
                metrics.totalStories,
              ),
            },
            {
              label: "Overdue",
              value: Number(pulseSummary.overdueStories ?? 0),
            },
            {
              label: "Blocked",
              value: Number(pulseSummary.blockedStories ?? 0),
            },
            {
              label: "Active risks",
              value: risks.length,
            },
            {
              label: "Overloaded",
              value: Number(pulseSummary.overloadedMembers ?? 0),
            },
            {
              label: "At-risk sprints",
              value: Number(pulseSummary.atRiskSprints ?? 0),
            },
            {
              label: "At-risk objectives",
              value: Number(pulseSummary.atRiskObjectives ?? 0),
            },
          ]}
        />

        <ChartSection
          description="Completion and work that is currently moving or waiting."
          title="Delivery health"
        >
          <KeyValueList
            rows={[
              {
                label: "Delivered",
                value: `${Number(metrics.completedStories ?? 0)} of ${Number(
                  metrics.totalStories ?? 0,
                )} (${completionRate(
                  metrics.completedStories,
                  metrics.totalStories,
                )})`,
              },
              {
                label: "In progress",
                value: Number(storyHealth.startedStories ?? 0),
              },
              {
                label: "Paused",
                value: Number(storyHealth.pausedStories ?? 0),
              },
              {
                label: "Unassigned",
                value: Number(storyHealth.unassignedStories ?? 0),
              },
            ]}
          />
        </ChartSection>

        {hasWorkload ? (
          <ChartSection
            description="Ownership, complexity, and the people carrying the most open work."
            title="Workload"
          >
            <Box className="space-y-3">
              <KeyValueList
                rows={[
                  {
                    label: "Total complexity",
                    value: Number(workloadSummary.totalEstimate ?? 0),
                  },
                  {
                    label: "Unassigned work",
                    value: Number(workloadSummary.unassignedStories ?? 0),
                  },
                  {
                    label: "Without complexity",
                    value: Number(workloadSummary.unestimatedStories ?? 0),
                  },
                ]}
              />
              {members.length ? (
                <CompactBarChart
                  bars={[
                    {
                      key: "openStories",
                      color: COLORS.primary,
                      name: "Open",
                    },
                    {
                      key: "overdueStories",
                      color: "#EF4444",
                      name: "Overdue",
                    },
                  ]}
                  data={members}
                  maxLabelLength={12}
                  xKey="memberName"
                />
              ) : null}
            </Box>
          </ChartSection>
        ) : null}

        {hasSprintData || hasObjectiveData ? (
          <Box className="grid gap-4 md:grid-cols-2">
            {hasSprintData ? (
              <HealthAndProgressSection
                description="Current sprint exposure and delivery progress."
                healthRows={[
                  {
                    label: "Active",
                    value: Number(sprintHealth.activeSprints ?? 0),
                  },
                  {
                    label: "At risk",
                    value: Number(sprintHealth.atRiskSprints ?? 0),
                  },
                  {
                    label: "Overdue",
                    value: Number(sprintHealth.overdueSprints ?? 0),
                  },
                ]}
                progressRows={sprintProgress.slice(0, 5).map((sprint) => ({
                  key: String(sprint.sprintId ?? sprint.sprintName),
                  label: String(sprint.sprintName),
                  value: `${progressSummary(
                    sprint.completed,
                    sprint.total,
                  )} · ${humanizeLabel(sprint.status)}`,
                }))}
                title="Sprint health and progress"
              />
            ) : null}
            {hasObjectiveData ? (
              <HealthAndProgressSection
                description="Objective exposure and average key-result progress."
                healthRows={[
                  {
                    label: "Active",
                    value: Number(objectiveHealth.activeObjectives ?? 0),
                  },
                  {
                    label: "At risk",
                    value: Number(objectiveHealth.atRiskObjectives ?? 0),
                  },
                  {
                    label: "Off track",
                    value: Number(objectiveHealth.offTrackObjectives ?? 0),
                  },
                  {
                    label: "Overdue",
                    value: Number(objectiveHealth.overdueObjectives ?? 0),
                  },
                ]}
                progressRows={objectiveProgress
                  .slice(0, 5)
                  .map((objective) => ({
                    key: String(
                      objective.objectiveId ?? objective.objectiveName,
                    ),
                    label: String(objective.objectiveName),
                    value: `${Math.round(
                      Number(objective.avgProgress ?? 0),
                    )}% average · ${Number(
                      objective.completed ?? 0,
                    )} of ${Number(objective.total ?? 0)} key results`,
                  }))}
                title="Objective health and progress"
              />
            ) : null}
          </Box>
        ) : null}

        {risks.length ? (
          <ChartSection
            description="The highest-signal delivery issues that need attention."
            title="Active risks"
          >
            <RiskList risks={risks} />
          </ChartSection>
        ) : null}

        {hasRequestData || hasEngagementData ? (
          <Box className="grid gap-4 md:grid-cols-2">
            {hasRequestData ? (
              <ChartSection title="Request sources">
                <KeyValueList
                  rows={[
                    {
                      key: "all-requests",
                      label: "All requests",
                      value: Number(requests.totalRequests ?? 0),
                    },
                    {
                      key: "pending-requests",
                      label: "Pending",
                      value: Number(requests.pendingRequests ?? 0),
                    },
                    ...providers.slice(0, 4).map((provider) => ({
                      key: `provider-${String(provider.provider ?? "source")}`,
                      label: humanizeLabel(provider.provider || "Source"),
                      value: `${Number(
                        provider.totalRequests ?? 0,
                      )} requests · ${ratioPercent(
                        provider.acceptanceRate,
                      )} accepted`,
                    })),
                  ]}
                />
              </ChartSection>
            ) : null}

            {hasEngagementData ? (
              <ChartSection title="Engagement">
                <KeyValueList
                  rows={[
                    {
                      key: "tracked-events",
                      label: "Tracked events",
                      value: Number(engagement.totalEvents ?? 0),
                    },
                    {
                      key: "active-people",
                      label: "Active people",
                      value: Number(engagement.uniqueUsers ?? 0),
                    },
                    ...engagementEvents.slice(0, 4).map((event) => ({
                      key: `event-${String(event.name ?? "event")}`,
                      label: humanizeLabel(event.name || "Event"),
                      value: Number(event.count ?? 0),
                    })),
                  ]}
                />
              </ChartSection>
            ) : null}
          </Box>
        ) : null}
      </Box>
    );
  }

  if (kind === "pulse-report") {
    const report = asRecord(output.report);
    const summary = asRecord(report.summary);
    const workload = asRecord(report.workload);
    const members = [...asRows(workload.members)]
      .sort(
        (firstMember, secondMember) =>
          Number(secondMember.openStories ?? 0) -
          Number(firstMember.openStories ?? 0),
      )
      .slice(0, 5);

    return (
      <Box className="mt-3 space-y-3">
        <Text className="text-xl font-semibold">{title}</Text>
        <MetricGrid
          metrics={[
            {
              label: "Open work",
              value: Number(summary.openStories ?? 0),
            },
            {
              label: "Overdue",
              value: Number(summary.overdueStories ?? 0),
            },
            {
              label: "Blocked",
              value: Number(summary.blockedStories ?? 0),
            },
            {
              label: "At-risk sprints",
              value: Number(summary.atRiskSprints ?? 0),
            },
            {
              label: "At-risk objectives",
              value: Number(summary.atRiskObjectives ?? 0),
            },
            {
              label: "Pending requests",
              value: Number(summary.pendingRequests ?? 0),
            },
          ]}
        />
        <ChartSection
          description="Shows the five members carrying the most open work."
          title="Highest workload"
        >
          <CompactBarChart
            bars={[
              { key: "openStories", color: COLORS.primary, name: "Open" },
              { key: "overdueStories", color: "#EF4444", name: "Overdue" },
            ]}
            data={members}
            maxLabelLength={12}
            xKey="username"
          />
        </ChartSection>
        <ChartSection
          description="The most important delivery issues that need attention."
          title="Active risks"
        >
          <RiskList risks={asRows(report.risks)} />
        </ChartSection>
      </Box>
    );
  }

  if (kind === "workload-analysis-report") {
    const analysis = asRecord(output.analysis);
    const summary = asRecord(analysis.summary);
    const risks = asRecord(analysis.risks);
    const members = [...asRows(analysis.members)]
      .sort(
        (firstMember, secondMember) =>
          Number(secondMember.openStories ?? 0) -
          Number(firstMember.openStories ?? 0),
      )
      .slice(0, 5);

    return (
      <Box className="mt-3 space-y-3">
        <Text className="text-xl font-semibold">{title}</Text>
        <MetricGrid
          metrics={[
            {
              label: "Open work",
              value: Number(summary.totalOpenStories ?? 0),
            },
            {
              label: "Complexity",
              value: Number(summary.totalEstimate ?? 0),
            },
            {
              label: "Overdue",
              value: Number(summary.overdueStories ?? 0),
            },
            {
              label: "Urgent",
              value: Number(summary.urgentStories ?? 0),
            },
            {
              label: "Unassigned",
              value: Number(summary.unassignedStories ?? 0),
            },
            {
              label: "No complexity",
              value: Number(summary.unestimatedStories ?? 0),
            },
          ]}
        />
        <ChartSection
          description="Shows the five members carrying the most open work."
          title="Workload by member"
        >
          <CompactBarChart
            bars={[
              { key: "openStories", color: COLORS.primary, name: "Open" },
              { key: "overdueStories", color: "#EF4444", name: "Overdue" },
            ]}
            data={members}
            maxLabelLength={12}
            xKey="username"
          />
        </ChartSection>
        <ChartSection title="Workload risks">
          <KeyValueList
            rows={[
              {
                label: "Overloaded members",
                value: asRows(risks.overloadedMembers).length,
              },
              {
                label: "Members with overdue work",
                value: asRows(risks.overdueMembers).length,
              },
              {
                label: "High-priority work",
                value: Number(risks.highPriorityStories ?? 0),
              },
            ]}
          />
        </ChartSection>
      </Box>
    );
  }

  if (kind === "story-performance-report") {
    const analytics = asRecord(output.analytics);

    return (
      <Box className="mt-3 space-y-3">
        <Text className="text-xl font-semibold">{title}</Text>
        <ChartSection title="Status breakdown">
          <CompactBarChart
            bars={[{ key: "count", color: COLORS.primary }]}
            data={asRows(analytics.statusBreakdown)}
            xKey="statusName"
          />
        </ChartSection>
        <ChartSection title="Completion by team">
          <CompactBarChart
            bars={[
              { key: "completed", color: COLORS.success, name: "Completed" },
              { key: "total", color: COLORS.muted, name: "Total" },
            ]}
            data={asRows(analytics.completionByTeam)}
            xKey="teamName"
          />
        </ChartSection>
      </Box>
    );
  }

  if (kind === "objective-progress-report") {
    const progress = asRecord(output.progress);

    return (
      <Box className="mt-3 space-y-3">
        <Text className="text-xl font-semibold">{title}</Text>
        <ChartSection title="Key-result progress">
          <CompactBarChart
            bars={[
              {
                key: "avgProgress",
                color: COLORS.primary,
                name: "Avg progress",
              },
            ]}
            data={asRows(progress.keyResultsProgress)}
            xKey="objectiveName"
          />
        </ChartSection>
        <ChartSection title="Progress by team">
          <CompactBarChart
            bars={[
              { key: "completed", color: COLORS.success, name: "Completed" },
              { key: "objectives", color: COLORS.muted, name: "Objectives" },
            ]}
            data={asRows(progress.progressByTeam)}
            xKey="teamName"
          />
        </ChartSection>
      </Box>
    );
  }

  if (kind === "team-performance-report") {
    const performance = asRecord(output.performance);
    const focusMember = asRecord(output.focusMember);

    return (
      <Box className="mt-3 space-y-3">
        <Flex align="center" justify="between">
          <Text className="text-xl font-semibold">{title}</Text>
          {focusMember.userId ? (
            <Text className="text-muted text-base">
              {completionRate(focusMember.completed, focusMember.assigned)}{" "}
              complete
            </Text>
          ) : null}
        </Flex>
        <ChartSection title="Team workload">
          <CompactBarChart
            bars={[
              { key: "completed", color: COLORS.success, name: "Completed" },
              { key: "assigned", color: COLORS.primary, name: "Assigned" },
              { key: "capacity", color: COLORS.warning, name: "Capacity" },
            ]}
            data={asRows(performance.teamWorkload)}
            xKey="teamName"
          />
        </ChartSection>
        <ChartSection title="Member contributions">
          <CompactBarChart
            bars={[
              { key: "completed", color: COLORS.success, name: "Completed" },
              { key: "assigned", color: COLORS.primary, name: "Assigned" },
            ]}
            data={asRows(performance.memberContributions).slice(0, 5)}
            maxLabelLength={12}
            xKey="username"
          />
        </ChartSection>
      </Box>
    );
  }

  if (kind === "sprint-performance-report") {
    const analytics = asRecord(output.analytics);

    return (
      <Box className="mt-3 space-y-3">
        <Text className="text-xl font-semibold">{title}</Text>
        <ChartSection title="Sprint progress">
          <CompactBarChart
            bars={[
              { key: "completed", color: COLORS.success, name: "Completed" },
              { key: "total", color: COLORS.muted, name: "Total" },
            ]}
            data={asRows(analytics.sprintProgress)}
            xKey="sprintName"
          />
        </ChartSection>
        <ChartSection title="Combined burndown">
          <CompactLineChart
            data={asRows(analytics.combinedBurndown)}
            lines={[
              { key: "planned", color: COLORS.muted, name: "Planned" },
              { key: "actual", color: COLORS.primary, name: "Actual" },
            ]}
            xKey="date"
          />
        </ChartSection>
      </Box>
    );
  }

  if (kind === "single-sprint-analytics-report") {
    const analytics = asRecord(output.analytics ?? output.analyticsReport);
    const overview = asRecord(analytics.overview);
    const storyBreakdown = asRecord(analytics.storyBreakdown);
    const burndown = asSingleSprintBurndown(analytics.burndown);
    const teamAllocation = [...asRows(analytics.teamAllocation)]
      .filter(
        (member) =>
          Number(member.assigned ?? 0) > 0 || Number(member.completed ?? 0) > 0,
      )
      .sort((firstMember, secondMember) => {
        const assignedDifference =
          Number(secondMember.assigned ?? 0) -
          Number(firstMember.assigned ?? 0);
        if (assignedDifference) return assignedDifference;

        return (
          Number(secondMember.completed ?? 0) -
          Number(firstMember.completed ?? 0)
        );
      })
      .slice(0, 5);

    return (
      <Box className="mt-3 space-y-3">
        <Text className="text-xl font-semibold">{title}</Text>
        <MetricGrid
          metrics={[
            {
              label: "Completion",
              value: `${Math.round(Number(overview.completionPercentage ?? 0))}%`,
            },
            {
              label: "Remaining",
              value: Number(overview.daysRemaining ?? 0),
            },
            { label: "Total", value: Number(storyBreakdown.total ?? 0) },
            {
              label: "Completed",
              value: Number(storyBreakdown.completed ?? 0),
            },
            {
              label: "In progress",
              value: Number(storyBreakdown.inProgress ?? 0),
            },
            { label: "Blocked", value: Number(storyBreakdown.blocked ?? 0) },
          ]}
        />
        <ChartSection
          description="Tracks remaining work against the ideal sprint pace."
          title="Burndown"
        >
          {burndown.length ? (
            <BurndownChart
              burndownData={burndown}
              className="h-52"
              workingDays={asWorkingDays(analytics.workingDays)}
            />
          ) : (
            <EmptyChart />
          )}
        </ChartSection>
        <ChartSection
          description="Shows completed and assigned work for each team member."
          title="Team allocation"
        >
          <CompactBarChart
            bars={[
              { key: "completed", color: COLORS.success, name: "Completed" },
              { key: "assigned", color: COLORS.primary, name: "Assigned" },
            ]}
            data={teamAllocation}
            maxLabelLength={12}
            xKey="username"
          />
        </ChartSection>
      </Box>
    );
  }

  if (kind === "timeline-trends-report") {
    const trends = asRecord(output.trends);

    return (
      <Box className="mt-3 space-y-3">
        <Text className="text-xl font-semibold">{title}</Text>
        <ChartSection
          title={`${getTermDisplay("storyTerm", { capitalize: true })} completion`}
        >
          <CompactLineChart
            data={asRows(trends.storyCompletion)}
            lines={[
              { key: "completed", color: COLORS.success, name: "Completed" },
              { key: "created", color: COLORS.primary, name: "Created" },
            ]}
            xKey="date"
          />
        </ChartSection>
        <ChartSection title="Key metrics">
          <CompactLineChart
            data={asRows(trends.keyMetricsTrend)}
            lines={[
              {
                key: "activeUsers",
                color: COLORS.primary,
                name: "Active users",
              },
              {
                key: "storiesPerDay",
                color: COLORS.success,
                name: "Stories/day",
              },
            ]}
            xKey="date"
          />
        </ChartSection>
      </Box>
    );
  }

  return null;
};
