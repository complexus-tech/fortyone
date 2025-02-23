"use client";

import type { FormEvent } from "react";
import { useMemo, useState } from "react";
import { Box, Input, Button, Text, Flex, Switch, ColorPicker } from "ui";
import { useRouter } from "next/navigation";
import type { Team } from "@/modules/teams/types";
import { useUpdateTeamMutation } from "@/modules/teams/hooks/use-update-team";
import { ConfirmDialog } from "@/components/ui";
import { SectionHeader } from "@/modules/settings/components/section-header";
import { useDeleteTeamMutation } from "@/modules/teams/hooks/delete-team-mutation";

export const GeneralSettings = ({ team }: { team: Team }) => {
  const router = useRouter();
  const [form, setForm] = useState({
    name: team.name,
    code: team.code,
    color: team.color,
    isPrivate: team.isPrivate,
  });
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const { mutate: deleteTeam, isPending } = useDeleteTeamMutation();

  const updateTeam = useUpdateTeamMutation(team.id);

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

  const handleDelete = () => {
    deleteTeam(team.id, {
      onSuccess: () => {
        router.push("/settings/workspace/teams");
      },
    });
  };

  return (
    <Box>
      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Basic information about your team."
          title="General Information"
        />

        <form
          className="divide-y-[0.5px] divide-gray-100 dark:divide-dark-100"
          onSubmit={handleSubmit}
        >
          <Flex align="center" className="px-6 py-4" justify="between">
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
                setForm({ ...form, code: e.target.value });
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
            <Box className="rounded-full border border-gray-100 bg-gray-50 dark:border-dark-50 dark:bg-dark-100">
              <ColorPicker
                onChange={(value) => {
                  setForm({ ...form, color: value });
                }}
                value={form.color}
              />
            </Box>
          </Flex>
          <Flex align="center" className="px-6 py-4" justify="between">
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
              disabled={!hasChanged}
              loading={updateTeam.isPending}
              type="submit"
            >
              Save Changes
            </Button>
          </Box>
        </form>
      </Box>

      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Permanently delete your team and all of its data."
          title="Delete Team"
        />
        <Box className="px-6 py-5">
          <Text color="muted">
            Deleting a team will remove all of its data and cannot be undone.
          </Text>
          <Button
            className="mt-4"
            onClick={() => {
              setIsDeleteOpen(true);
            }}
          >
            Delete Team
          </Button>
        </Box>
      </Box>

      <ConfirmDialog
        confirmPhrase="i understand"
        confirmText="Delete team"
        description="Are you sure you want to delete this team? All of the team's data will be permanently removed. This action cannot be undone."
        isLoading={isPending}
        isOpen={isDeleteOpen}
        loadingText="Deleting team..."
        onClose={() => {
          setIsDeleteOpen(false);
        }}
        onConfirm={handleDelete}
        title="Delete team"
      />
    </Box>
  );
};
