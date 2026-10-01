"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import { Box, Button, Flex, Select, Text } from "ui";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useUserRole } from "@/hooks/role";
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
      aria-labelledby="work-export-heading"
      as="section"
      className="border-border bg-surface mt-6 rounded-2xl border-[0.5px] p-6"
    >
      <Text as="h2" className="text-xl font-medium" id="work-export-heading">
        Export your work
      </Text>
      <Text className="mt-2 max-w-3xl leading-6" color="muted">
        Back up your work as JSON or export a CSV for analysis.
      </Text>
      <Flex className="mt-5 gap-4" wrap>
        <Box className="min-w-0 flex-1 basis-60">
          <Text className="mb-2" id="export-team-label">
            Teams
          </Text>
          <Select
            disabled={teamsPending || teamsError || exportWork.isPending}
            onValueChange={(value) => {
              setTeamId(value === "all" ? "" : value);
            }}
            value={teamId || "all"}
          >
            <Select.Trigger
              aria-labelledby="export-team-label"
              className="h-11 w-full text-base"
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
        </Box>
        <Box className="min-w-0 flex-1 basis-60">
          <Text className="mb-2" id="export-format-label">
            Format
          </Text>
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
              className="h-11 w-full text-base"
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
        </Box>
      </Flex>
      <Text className="mt-4 leading-6" color="muted" id="export-format-help">
        {format === "json"
          ? "Includes tasks, comments, relationships, and custom fields. Uploaded files are excluded."
          : "Includes task properties and custom fields for spreadsheet analysis."}
      </Text>
      <Text className="mt-2 leading-6" color="muted">
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
        className="mt-5"
        color="invert"
        disabled={
          !session || teamsPending || teamsError || exportWork.isPending
        }
        loading={exportWork.isPending}
        onClick={() => {
          exportWork.mutate();
        }}
        size="lg"
      >
        {exportWork.isPending ? "Preparing export…" : "Download export"}
      </Button>
    </Box>
  );
};
