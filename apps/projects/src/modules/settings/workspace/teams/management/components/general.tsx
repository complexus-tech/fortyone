"use client";

import type { FormEvent } from "react";
import { useState } from "react";
import {
  Box,
  Input,
  Button,
  Text,
  Flex,
  Switch,
  ColorPicker,
  Wrapper,
  Select,
} from "ui";
import { WarningIcon } from "icons";
import type { Team } from "@/modules/teams/types";
import { useUpdateTeamMutation } from "@/modules/teams/hooks/use-update-team";
import { SectionHeader } from "@/modules/settings/components/section-header";
import { FeatureGuard } from "@/components/ui";
import { useUserRole, useWorkspacePath } from "@/hooks";
import {
  TEAM_STORY_TERMS,
  useTerminology,
} from "@/hooks/use-terminology-display";

const formatTeamCode = (name: string) =>
  name
    .toUpperCase()
    .replace(/[^A-Z0-9]/g, "")
    .slice(0, 3);

export const GeneralSettings = ({ team }: { team: Team }) => {
  const [form, setForm] = useState({
    name: team.name,
    code: team.code,
    color: team.color,
    isPrivate: team.isPrivate,
    storyTerm: team.storyTerm ?? null,
  });
  const { userRole } = useUserRole();
  const { withWorkspace } = useWorkspacePath();
  const updateTeam = useUpdateTeamMutation(team.id);
  const { getTermDisplay } = useTerminology(null);

  const hasChanged =
    form.name !== team.name ||
    form.code !== team.code ||
    form.color !== team.color ||
    form.isPrivate !== team.isPrivate ||
    form.storyTerm !== (team.storyTerm ?? null);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    await updateTeam.mutateAsync({
      ...form,
      code: form.code.trim(),
      name: form.name.trim(),
    });
  };

  return (
    <Box className="border-border bg-surface rounded-2xl border">
      <SectionHeader
        description="Name, appearance, and privacy for your team."
        title="Team details"
      />

      <form className="divide-border divide-y-[0.5px]" onSubmit={handleSubmit}>
        <Box className="flex flex-col gap-4 px-6 py-4 md:flex-row md:items-center md:justify-between">
          <Box>
            <Text>Team name</Text>
            <Text color="muted" fontSize="sm">
              Choose a descriptive name for your team
            </Text>
          </Box>
          <Input
            aria-label="Team name"
            className="h-10 md:w-64"
            maxLength={24}
            minLength={3}
            name="name"
            onChange={(e) => {
              setForm({ ...form, name: e.target.value });
            }}
            placeholder="eg. Growth, Product, Operations"
            required
            value={form.name}
          />
        </Box>
        <Flex align="center" className="px-6 py-4" justify="between">
          <Box>
            <Text>Team code</Text>
            <Text color="muted" fontSize="sm">
              Prefix for this team&apos;s work IDs (e.g., ENG-123)
            </Text>
          </Box>
          <Input
            aria-label="Team code"
            className="h-10 w-28"
            maxLength={3}
            minLength={2}
            name="code"
            onChange={(e) => {
              setForm({ ...form, code: formatTeamCode(e.target.value) });
            }}
            placeholder="ENG"
            required
            value={form.code}
          />
        </Flex>
        <Box className="flex flex-col gap-4 px-6 py-4 md:flex-row md:items-center md:justify-between">
          <Box>
            <Text id={`team-work-naming-${team.id}`}>Work naming</Text>
            <Text color="muted">Name this team&apos;s work.</Text>
          </Box>
          <Select
            onValueChange={(value) => {
              const term = TEAM_STORY_TERMS.find(
                (option) => option.value === value,
              );
              setForm((current) => ({
                ...current,
                storyTerm: term?.value ?? null,
              }));
            }}
            value={form.storyTerm ?? "workspace-default"}
          >
            <Select.Trigger
              aria-labelledby={`team-work-naming-${team.id}`}
              className="w-max max-w-full text-base"
            >
              <Select.Input />
            </Select.Trigger>
            <Select.Content>
              <Select.Option className="text-base" value="workspace-default">
                Workspace default (
                {getTermDisplay("storyTerm", { capitalize: true })})
              </Select.Option>
              {TEAM_STORY_TERMS.map((term) => (
                <Select.Option
                  className="text-base"
                  key={term.value}
                  value={term.value}
                >
                  {term.label}
                </Select.Option>
              ))}
            </Select.Content>
          </Select>
        </Box>
        <Flex align="center" className="px-6 py-4" justify="between">
          <Box>
            <Text>Team color</Text>
            <Text color="muted" fontSize="sm">
              Used to identify the team in the workspace
            </Text>
          </Box>
          <ColorPicker
            onChange={(value) => {
              setForm({ ...form, color: value });
            }}
            value={form.color}
          />
        </Flex>
        <FeatureGuard
          fallback={
            <Box className="px-6 py-4">
              <Wrapper className="border-warning bg-warning/10 dark:border-warning/20 dark:bg-warning/10 flex items-center justify-between gap-2 border p-4">
                <Flex align="center" gap={2}>
                  <WarningIcon className="text-warning dark:text-warning" />
                  <Text>
                    {userRole === "admin"
                      ? "Upgrade"
                      : "Ask your admin to upgrade"}{" "}
                    to a higher plan to create private teams
                  </Text>
                </Flex>
                {userRole === "admin" && (
                  <Button
                    color="warning"
                    href={withWorkspace("/settings/workspace/billing")}
                  >
                    Upgrade now
                  </Button>
                )}
              </Wrapper>
            </Box>
          }
          feature="privateTeams"
        >
          <Flex align="center" className="gap-3 px-6 py-4" justify="between">
            <Box>
              <Text>Private team</Text>
              <Text className="max-w-xl" color="muted" fontSize="sm">
                Private teams are only visible to members of the team. Admin and
                team leads can add members to private teams.
              </Text>
            </Box>
            <Switch
              aria-label="Private team"
              checked={form.isPrivate}
              className="shrink-0"
              name="isPrivate"
              onCheckedChange={(checked) => {
                setForm({ ...form, isPrivate: checked });
              }}
            />
          </Flex>
        </FeatureGuard>
        {hasChanged ? (
          <Box className="px-6 py-4">
            <Button loading={updateTeam.isPending} type="submit">
              Save Changes
            </Button>
          </Box>
        ) : null}
      </form>
    </Box>
  );
};
