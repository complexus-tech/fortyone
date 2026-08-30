import { cn } from "lib";
import { Box, Button, Flex, Text } from "ui";
import type { AnalyticsReportOutput } from "./model";
import { asRecord, asRows } from "./model";
import { ChartSection, KeyValueList, MetricGrid, PillList } from "./primitives";

type ReportProps = {
  output: AnalyticsReportOutput;
  title: string;
};

export const GitHubIntegrationReport = ({ output, title }: ReportProps) => {
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
};

export const GitHubTeamAutomationReport = ({ output, title }: ReportProps) => {
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
};

export const GitHubStoryReport = ({
  output,
  storyTerm,
  title,
}: ReportProps & { storyTerm: string }) => {
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
                <Button color="black" href={link.url} size="sm" target="_blank">
                  Open
                </Button>
              ) : null}
            </Flex>
          ))}
        </Box>
      )}
    </Box>
  );
};
