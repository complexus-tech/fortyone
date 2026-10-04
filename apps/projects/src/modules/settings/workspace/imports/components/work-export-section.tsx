"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import { Box, Button, Flex, Select, Text } from "ui";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useUserRole } from "@/hooks/role";
import { SectionHeader } from "@/components/ui/section-header";
import { useJoinedTeams } from "@/modules/teams/public/client";
import { downloadWorkExport, getWorkExport } from "../work-export";

export const WorkExportSection = () => {
  const [teamId, setTeamId] = useState("");
  const [format, setFormat] = useState<"json" | "csv">("json");
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const { userRole } = useUserRole();
  const {
    data: teams = [],
    isPending: teamsPending,
    isError: teamsError,
  } = useJoinedTeams();
  const exportWork = useMutation({
    mutationFn: () => getWorkExport(teamId || null, { session, workspaceSlug }),
    onSuccess: (backup) => {
      downloadWorkExport(backup, format, workspaceSlug);
      toast.success("Export ready", {
        description: `${backup.analysis.tasks.length.toLocaleString()} items included.`,
      });
    },
  });
  if (userRole !== "admin") return null;
  return (
    <Box
      aria-label="Export your work"
      as="section"
      className="border-border bg-surface mt-6 overflow-hidden rounded-2xl border-[0.5px]"
    >
      <SectionHeader
        description="Back up your work as JSON or export a CSV for analysis."
        title="Export your work"
      />
      <Box className="divide-border divide-y-[0.5px]">
        <Flex
          align="center"
          className="flex-col items-start gap-4 px-6 py-4 md:flex-row md:items-center"
          justify="between"
        >
          <Box>
            <Text className="font-medium" id="export-team-label">
              Teams
            </Text>
            <Text color="muted">
              Choose a joined team or include all joined teams.
            </Text>
          </Box>
          <Select
            disabled={teamsPending || teamsError || exportWork.isPending}
            onValueChange={(value) => {
              setTeamId(value === "all" ? "" : value);
            }}
            value={teamId || "all"}
          >
            <Select.Trigger
              aria-labelledby="export-team-label"
              className="max-w-64 shrink-0 text-base"
            >
              <Select.Input />
            </Select.Trigger>
            <Select.Content>
              <Select.Option className="text-base" value="all">
                All joined teams
              </Select.Option>
              {teams.map((team) => (
                <Select.Option
                  className="text-base"
                  key={team.id}
                  value={team.id}
                >
                  {team.name}
                </Select.Option>
              ))}
            </Select.Content>
          </Select>
        </Flex>
        <Flex
          align="center"
          className="flex-col items-start gap-4 px-6 py-4 md:flex-row md:items-center"
          justify="between"
        >
          <Box>
            <Text className="font-medium" id="export-format-label">
              Format
            </Text>
            <Text className="max-w-md" color="muted" id="export-format-help">
              {format === "json"
                ? "Includes tasks, comments, relationships, and custom fields. Uploaded files are excluded."
                : "Includes task properties and custom fields for spreadsheet analysis."}
            </Text>
          </Box>
          <Select
            disabled={exportWork.isPending}
            onValueChange={(value) => {
              setFormat(value === "csv" ? "csv" : "json");
            }}
            value={format}
          >
            <Select.Trigger
              aria-describedby="export-format-help"
              aria-labelledby="export-format-label"
              className="shrink-0 text-base"
            >
              <Select.Input />
            </Select.Trigger>
            <Select.Content>
              <Select.Option className="text-base" value="json">
                JSON backup
              </Select.Option>
              <Select.Option className="text-base" value="csv">
                CSV spreadsheet
              </Select.Option>
            </Select.Content>
          </Select>
        </Flex>
        <Box className="px-6 py-4">
          <Text color="muted">
            Up to 10,000 items · 20 MB · Joined teams only
          </Text>
          {teamsError ? (
            <Text className="mt-4" color="danger" role="alert">
              Teams could not be loaded. Refresh this page to try again.
            </Text>
          ) : null}
          {exportWork.error ? (
            <Text className="mt-4" color="danger" role="alert">
              {exportWork.error.message}
            </Text>
          ) : null}
          <Button
            className="mt-4"
            disabled={
              !session || teamsPending || teamsError || exportWork.isPending
            }
            loading={exportWork.isPending}
            onClick={() => {
              exportWork.mutate();
            }}
          >
            {exportWork.isPending ? "Preparing export…" : "Download export"}
          </Button>
        </Box>
      </Box>
    </Box>
  );
};
