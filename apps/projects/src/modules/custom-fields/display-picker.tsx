"use client";

import { useState } from "react";
import { useParams } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import { Box, Button, Checkbox, Select, Text } from "ui";
import { useJoinedTeams } from "@/modules/teams/public/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import type { StoryPropertyDisplayPickerProps } from "@/shared/story/board-property-slots";
import { customFieldKeys, useTeamCustomFields } from "./hooks";
import type { CustomField } from "./types";
import { CustomFieldIcon } from "./icons";

export const CustomFieldDisplayPicker = ({
  selectedIds,
  onChange,
}: StoryPropertyDisplayPickerProps) => {
  const params = useParams<{ teamId?: string }>();
  const { workspaceSlug } = useWorkspacePath();
  const [selectedTeam, setSelectedTeam] = useState("");
  const { data: teams = [] } = useJoinedTeams();
  const teamId = params.teamId || selectedTeam || teams.at(0)?.id || "";
  const {
    data: fields = [],
    isPending,
    isError,
    refetch,
  } = useTeamCustomFields(teamId);
  const queryClient = useQueryClient();
  const knownFields = new Map(
    queryClient
      .getQueriesData<CustomField[]>({
        queryKey: customFieldKeys.all(workspaceSlug),
        predicate: (query) => query.queryKey[2] === "team",
      })
      .flatMap(([, data]) => data ?? [])
      .map((field) => [field.id, field]),
  );
  return (
    <Box className="border-border mt-5 border-t pt-4">
      <Text className="mb-2" fontWeight="medium">
        Custom fields
      </Text>
      <Text className="mb-3 leading-6" color="muted">
        Show up to three details on cards and rows.
      </Text>
      {!params.teamId && teams.length ? (
        <Select onValueChange={setSelectedTeam} value={teamId}>
          <Select.Trigger
            aria-label="Custom fields team"
            className="mb-3 h-11 w-full text-base"
          >
            <Select.Input />
          </Select.Trigger>
          <Select.Content>
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
      ) : null}
      {isPending && teamId ? <Text color="muted">Loading fields…</Text> : null}
      {isError ? (
        <Box>
          <Text color="danger" role="alert">
            Fields could not be loaded.
          </Text>
          <Button
            className="mt-2"
            color="tertiary"
            onClick={() => void refetch()}
            variant="outline"
          >
            Try again
          </Button>
        </Box>
      ) : null}
      <Box className="max-h-60 space-y-3 overflow-y-auto">
        {fields
          .filter((field) => !field.archivedAt)
          .map((field) => (
            <label className="flex items-center gap-3" key={field.id}>
              <Checkbox
                checked={selectedIds.includes(field.id)}
                disabled={
                  !selectedIds.includes(field.id) && selectedIds.length >= 3
                }
                onCheckedChange={(checked) => {
                  onChange(
                    checked
                      ? [...selectedIds, field.id].slice(0, 3)
                      : selectedIds.filter((id) => id !== field.id),
                  );
                }}
              />
              <CustomFieldIcon className="h-4 w-auto shrink-0" field={field} />
              <span className="min-w-0 break-words">{field.name}</span>
            </label>
          ))}
      </Box>
      {!isPending && !isError && !fields.some((field) => !field.archivedAt) ? (
        <Text color="muted">This team has no active custom fields.</Text>
      ) : null}
      {selectedIds.length ? (
        <Box className="mt-4 space-y-2">
          <Text color="muted">Selected ({selectedIds.length}/3)</Text>
          {selectedIds.map((id) => (
            <Box className="flex items-center justify-between gap-2" key={id}>
              <Text className="min-w-0 truncate">
                {knownFields.get(id)?.name ?? "Field from another team"}
              </Text>
              <Button
                aria-label={`Remove ${knownFields.get(id)?.name ?? "selected field"}`}
                color="tertiary"
                onClick={() => {
                  onChange(selectedIds.filter((selected) => selected !== id));
                }}
                variant="naked"
              >
                Remove
              </Button>
            </Box>
          ))}
        </Box>
      ) : null}
    </Box>
  );
};
