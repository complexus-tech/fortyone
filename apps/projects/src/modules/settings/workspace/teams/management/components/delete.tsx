import { Box, Button, Text } from "ui";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { SectionHeader } from "@/modules/settings/components/section-header";
import { ConfirmDialog } from "@/components/ui";
import { useDeleteTeamMutation } from "@/modules/teams/hooks/delete-team-mutation";
import type { Team } from "@/modules/teams/types";

export const DeleteTeam = ({ team }: { team: Team }) => {
  const router = useRouter();
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const { mutate: deleteTeam, isPending } = useDeleteTeamMutation();

  const handleDelete = () => {
    deleteTeam(team.id, {
      onSuccess: () => {
        router.push("/settings/workspace/teams");
      },
    });
  };
  return (
    <div>
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
    </div>
  );
};
