"use client";

import type { FormEvent } from "react";
import { useState } from "react";
import { Box, Text, Button, Input, ColorPicker, Dialog, Flex } from "ui";
import { SearchIcon } from "icons";
import { toast } from "sonner";
import { useTeams } from "@/modules/teams/hooks/teams";
import { SectionHeader } from "@/modules/settings/components";
import type { CreateTeamInput } from "@/modules/teams/types";
import { useCreateTeamMutation } from "@/modules/teams/hooks/use-create-team";
import { WorkspaceTeam } from "../components/team";

const initialForm = {
  name: "",
  code: "",
  color: "#2563eb",
};

export const TeamsList = () => {
  const { data: teams = [] } = useTeams();
  const [search, setSearch] = useState("");
  const [isDialogOpen, setIsDialogOpen] = useState(false);
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
    setIsDialogOpen(false);
    setForm(initialForm);
  };

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Teams
      </Text>
      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          action={
            <Button
              onClick={() => {
                setIsDialogOpen(true);
              }}
            >
              Create Team
            </Button>
          }
          description="Manage your teams and their members."
          title="Team Management"
        />
        {teams.length > 10 && (
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
        )}

        {filteredTeams.map((team) => (
          <WorkspaceTeam {...team} key={team.id} />
        ))}
      </Box>

      <Dialog onOpenChange={setIsDialogOpen} open={isDialogOpen}>
        <Dialog.Content className="max-w-3xl">
          <Dialog.Header>
            <Dialog.Title className="px-6 pt-1 text-xl">
              Create a New Team
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Description>
            Teams help organize your workspace and manage access control. Create
            a new team to group members, and collaborate on objectives.
          </Dialog.Description>
          <Dialog.Body className="mt-3">
            <form
              className="divide-y-[0.5px] divide-gray-100 dark:divide-dark-100"
              onSubmit={handleSubmit}
            >
              <Flex align="center" className="py-3" justify="between">
                <Box>
                  <Text>Team name</Text>
                  <Text color="muted" fontSize="sm">
                    Choose a descriptive name for your team
                  </Text>
                </Box>
                <Input
                  className="h-[2.5rem] w-80"
                  name="name"
                  onChange={(e) => {
                    setForm({ ...form, name: e.target.value });
                  }}
                  placeholder="eg. Growth, Product, Operations"
                  required
                  value={form.name}
                />
              </Flex>
              <Flex align="center" className="py-3" justify="between">
                <Box>
                  <Text>Team code</Text>
                  <Text color="muted" fontSize="sm">
                    Short prefix for team&apos;s story IDs (e.g., ENG-123)
                  </Text>
                </Box>
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
              </Flex>
              <Flex align="center" className="py-3" justify="between">
                <Box>
                  <Text>Team color</Text>
                  <Text color="muted" fontSize="sm">
                    Used to identify the team in the workspace
                  </Text>
                </Box>
                <Box className="rounded-full border border-gray-100 bg-gray-50 dark:border-dark-50 dark:bg-dark-100">
                  <ColorPicker
                    onChange={(value) => {
                      setForm({ ...form, color: value });
                    }}
                    value={form.color}
                  />
                </Box>
              </Flex>

              <Box className="pt-4">
                <Button
                  loading={createTeam.isPending}
                  loadingText="Creating team..."
                  type="submit"
                >
                  Create Team
                </Button>
              </Box>
            </form>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>
    </Box>
  );
};
