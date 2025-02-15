"use client";

import type { FormEvent } from "react";
import { useState } from "react";
import { Box, Text, Button, Input, ColorPicker } from "ui";
import { SearchIcon } from "icons";
import { toast } from "sonner";
import { useTeams } from "@/modules/teams/hooks/teams";
import { SectionHeader } from "@/modules/settings/components";
import type { CreateTeamInput } from "@/modules/teams/types";
import { useCreateTeamMutation } from "@/modules/teams/hooks/use-create-team";
import { RowWrapper } from "@/components/ui";
import { WorkspaceTeam } from "../components/team";

const initialForm = {
  name: "",
  code: "",
  color: "#2563eb",
};

export const TeamsList = () => {
  const { data: teams = [] } = useTeams();
  const [search, setSearch] = useState("");
  const [form, setForm] = useState<CreateTeamInput>(initialForm);
  const createTeam = useCreateTeamMutation();

  const filteredTeams = teams.filter(
    (team) =>
      team.name.toLowerCase().includes(search.toLowerCase()) ||
      team.code.toLowerCase().includes(search.toLowerCase()),
  );

  const handleSubmit = async (e: FormEvent) => {
    const teamCodes = teams.map((team) => team.code);
    if (teamCodes.includes(form.code)) {
      toast.warning("Validation error", {
        description: "Team code already exists",
      });
      return;
    }
    e.preventDefault();
    await createTeam.mutateAsync(form);
    setForm(initialForm);
  };

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Teams
      </Text>
      <Box className="mb-5 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Create a new team to organize your workspace."
          title="Create new team"
        />
        <form onSubmit={handleSubmit}>
          <RowWrapper className="px-6">
            <Text color="muted">Team name</Text>
            <Input
              className="h-[2.5rem] w-72"
              name="name"
              onChange={(e) => {
                setForm({ ...form, name: e.target.value });
              }}
              placeholder="eg. Marketing"
              required
              value={form.name}
            />
          </RowWrapper>
          <RowWrapper className="px-6">
            <Text color="muted">Team code</Text>
            <Input
              className="h-[2.5rem] w-28"
              name="code"
              onChange={(e) => {
                setForm({ ...form, code: e.target.value });
              }}
              placeholder="ENG"
              required
              value={form.code}
            />
          </RowWrapper>
          <RowWrapper className="px-6">
            <Text color="muted">Team color</Text>
            <Box className="rounded-full border border-gray-100 bg-gray-50 dark:border-dark-50 dark:bg-dark-100">
              <ColorPicker
                onChange={(value) => {
                  setForm({ ...form, color: value });
                }}
                value={form.color}
              />
            </Box>
          </RowWrapper>

          <Box className="px-6 py-4">
            <Button
              loading={createTeam.isPending}
              loadingText="Creating team..."
              type="submit"
            >
              Create Team
            </Button>
          </Box>
        </form>
      </Box>
      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Manage your teams and their members."
          title="Team Management"
        />
        <Box className="border-b border-gray-100 px-6 py-4 dark:border-dark-100">
          <Box className="relative max-w-md">
            <SearchIcon className="text-gray-400 absolute left-3 top-1/2 h-4 -translate-y-1/2" />
            <Input
              className="pl-9"
              onChange={(e) => {
                setSearch(e.target.value);
              }}
              placeholder="Search teams..."
              type="search"
              value={search}
            />
          </Box>
        </Box>

        {filteredTeams.map((team) => (
          <WorkspaceTeam {...team} key={team.id} />
        ))}
      </Box>
    </Box>
  );
};
