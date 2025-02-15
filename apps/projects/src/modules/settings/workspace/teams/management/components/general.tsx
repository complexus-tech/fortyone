"use client";

import { useState } from "react";
import { Box, Input, Button, Text } from "ui";
import { toast } from "sonner";
import type { Team } from "@/modules/teams/types";
import { useUpdateTeamMutation } from "@/modules/teams/hooks/use-update-team";
import { useDeleteTeamMutation } from "@/modules/teams/hooks/use-delete-team";
import { ConfirmDialog } from "@/components/ui";
import { SectionHeader } from "@/modules/settings/components/section-header";

export const GeneralSettings = ({ team }: { team: Team }) => {
  const [form, setForm] = useState({
    name: team.name,
    description: team.description,
    code: team.code,
    color: team.color,
  });
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);

  const updateTeam = useUpdateTeamMutation(team.id);
  const deleteTeam = useDeleteTeamMutation();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await updateTeam.mutateAsync(form);
  };

  const handleDelete = async () => {
    await deleteTeam.mutateAsync(team.id);
    // Redirect to teams list
    window.location.href = "/settings/workspace/teams";
  };

  return (
    <Box>
      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Basic information about your team."
          title="General Information"
        />

        <form className="p-6" onSubmit={handleSubmit}>
          <Box className="space-y-4">
            <Input
              label="Team Name"
              name="name"
              onChange={(e) => {
                setForm({ ...form, name: e.target.value });
              }}
              placeholder="Engineering Team"
              required
              value={form.name}
            />
            <Input
              label="Description"
              name="description"
              onChange={(e) => {
                setForm({ ...form, description: e.target.value });
              }}
              placeholder="Core engineering team responsible for product development"
              value={form.description}
            />
            <Input
              label="Team Code"
              name="code"
              onChange={(e) => {
                setForm({ ...form, code: e.target.value });
              }}
              placeholder="ENG"
              required
              value={form.code}
            />
            <Box>
              <Text className="text-gray-700 mb-2 block text-sm font-medium dark:text-gray-200">
                Team Color
              </Text>
              <Input
                name="color"
                onChange={(e) => {
                  setForm({ ...form, color: e.target.value });
                }}
                placeholder="#2563eb"
                type="color"
                value={form.color}
              />
            </Box>
          </Box>
          <Button className="mt-4" loading={updateTeam.isPending} type="submit">
            Save Changes
          </Button>
        </form>
      </Box>

      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Permanently delete your team and all of its data."
          title="Delete Team"
        />
        <Box className="p-6">
          <Text color="muted">
            Once you delete a team, there is no going back. Please be certain.
          </Text>
          <Button
            className="mt-4"
            color="danger"
            onClick={() => {
              setIsDeleteOpen(true);
            }}
            variant="outline"
          >
            Delete Team
          </Button>
        </Box>
      </Box>

      <ConfirmDialog
        confirmText="Delete team"
        description="Are you sure you want to delete this team? All of the team's data will be permanently removed. This action cannot be undone."
        isOpen={isDeleteOpen}
        onClose={() => {
          setIsDeleteOpen(false);
        }}
        onConfirm={handleDelete}
        title="Delete team"
      />
    </Box>
  );
};
