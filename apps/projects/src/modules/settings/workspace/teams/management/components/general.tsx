"use client";

import type { FormEvent } from "react";
import { useMemo, useState } from "react";
import {
  Box,
  Input,
  Button,
  Text,
  Flex,
  Switch,
  ColorPicker,
  Wrapper,
} from "ui";
import { WarningIcon } from "icons";
import type { Team } from "@/modules/teams/types";
import { useUpdateTeamMutation } from "@/modules/teams/hooks/use-update-team";
import { SectionHeader } from "@/modules/settings/components/section-header";
import { FeatureGuard } from "@/components/ui";
import { useUserRole } from "@/hooks";

export const GeneralSettings = ({ team }: { team: Team }) => {
  const [form, setForm] = useState({
    name: team.name,
    code: team.code,
    color: team.color,
    isPrivate: team.isPrivate,
  });
  const { userRole } = useUserRole();
  const updateTeam = useUpdateTeamMutation(team.id);

  const formatCode = (name: string) => {
    return name
      .toUpperCase()
      .replace(/[^A-Z0-9]/g, "") // Remove non-alphanumeric chars
      .slice(0, 3); // Take first 3 characters
  };

  const hasChanged = useMemo(() => {
    return (
      form.name !== team.name ||
      form.code !== team.code ||
      form.color !== team.color ||
      form.isPrivate !== team.isPrivate
    );
  }, [form, team]);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    await updateTeam.mutateAsync(form);
  };

  return (
    <Box className="rounded-2xl border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
      <SectionHeader
        description="Basic information about your team."
        title="General Information"
      />

      <form
        className="divide-y-[0.5px] divide-gray-100 dark:divide-dark-100"
        onSubmit={handleSubmit}
      >
        <Box className="flex flex-col gap-4 px-6 py-4 md:flex-row md:items-center md:justify-between">
          <Box>
            <Text>Team name</Text>
            <Text color="muted" fontSize="sm">
              Choose a descriptive name for your team
            </Text>
          </Box>
          <Input
            className="h-[2.5rem] md:w-80"
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
              Short prefix for team&apos;s story IDs (e.g., ENG-123)
            </Text>
          </Box>
          <Input
            className="h-[2.5rem] w-28"
            maxLength={3}
            minLength={2}
            name="code"
            onChange={(e) => {
              setForm({ ...form, code: formatCode(e.target.value) });
            }}
            placeholder="ENG"
            required
            value={form.code}
          />
        </Flex>
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
              <Wrapper className="flex items-center justify-between gap-2 border border-warning bg-warning/10 p-4 dark:border-warning/20 dark:bg-warning/10">
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
                  <Button color="warning" href="/settings/workspace/billing">
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
