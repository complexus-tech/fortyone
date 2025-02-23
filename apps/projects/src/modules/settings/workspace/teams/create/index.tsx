"use client";

import type { FormEvent } from "react";
import { useState } from "react";
import { Box, Text, Button, Input, ColorPicker, Flex, Switch } from "ui";
import { toast } from "sonner";
import { useTeams } from "@/modules/teams/hooks/teams";
import { SectionHeader } from "@/modules/settings/components";
import type { CreateTeamInput } from "@/modules/teams/types";
import { useCreateTeamMutation } from "@/modules/teams/hooks/use-create-team";

const initialForm = {
  name: "",
  code: "",
  color: "#2563eb",
  isPrivate: false,
};

export const CreateTeam = () => {
  const { data: teams = [] } = useTeams();
  const [form, setForm] = useState<CreateTeamInput>(initialForm);
  const createTeam = useCreateTeamMutation();

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
        Create a new team
      </Text>
      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Create a new team to group members, and collaborate on objectives."
          title="Create a new team"
        />
        <form
          className="divide-y-[0.5px] divide-gray-100 dark:divide-dark-100"
          onSubmit={handleSubmit}
        >
          <Flex align="center" className="px-6 py-3" justify="between">
            <Box>
              <Text>Team name</Text>
              <Text color="muted" fontSize="sm">
                Choose a descriptive name for your team
              </Text>
            </Box>
            <Input
              className="h-[2.5rem] w-80"
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
          </Flex>
          <Flex align="center" className="px-6 py-3" justify="between">
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
                setForm({ ...form, code: e.target.value });
              }}
              placeholder="ENG"
              required
              value={form.code}
            />
          </Flex>
          <Flex align="center" className="px-6 py-3" justify="between">
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
          <Flex align="center" className="px-6 py-3" justify="between">
            <Box>
              <Text>Private team</Text>
              <Text className="max-w-xl" color="muted" fontSize="sm">
                Private teams are only visible to members of the team. Admin and
                team leads can add members to private teams.
              </Text>
            </Box>
            <Switch
              checked={form.isPrivate}
              name="isPrivate"
              onCheckedChange={(checked) => {
                setForm({ ...form, isPrivate: checked });
              }}
            />
          </Flex>

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
    </Box>
  );
};
