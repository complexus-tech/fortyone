"use client";

import { useState } from "react";
import { Dialog, Box, Input, Button, Text } from "ui";
import type { CreateTeamInput } from "@/modules/teams/types";
import { useCreateTeamMutation } from "@/modules/teams/hooks/use-create-team";

export const CreateTeamDialog = ({
  isOpen,
  onClose,
}: {
  isOpen: boolean;
  onClose: () => void;
}) => {
  const [form, setForm] = useState<CreateTeamInput>({
    name: "",
    code: "",
    color: "#2563eb",
  });

  const createTeam = useCreateTeamMutation();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await createTeam.mutateAsync(form);
    onClose();
  };

  return (
    <Dialog onOpenChange={onClose} open={isOpen}>
      <Dialog.Content>
        <Dialog.Header>
          <Dialog.Title>Create Team</Dialog.Title>
        </Dialog.Header>
        <form onSubmit={handleSubmit}>
          <Box className="space-y-4 p-6">
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
          <Dialog.Footer>
            <Button onClick={onClose} variant="outline">
              Cancel
            </Button>
            <Button loading={createTeam.isPending} type="submit">
              Create Team
            </Button>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
