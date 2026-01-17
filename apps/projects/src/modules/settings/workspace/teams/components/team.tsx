import { Box, Flex, Text, Button, Menu, TimeAgo } from "ui";
import { DeleteIcon, MoreHorizontalIcon, SettingsIcon } from "icons";
import { useState } from "react";
import Link from "next/link";
import type { Team } from "@/modules/teams/types";
import { ConfirmDialog, RowWrapper, TeamColor } from "@/components/ui";
import { useDeleteTeamMutation } from "@/modules/teams/hooks/delete-team-mutation";
import { useWorkspacePath } from "@/hooks";

export const WorkspaceTeam = ({
  id,
  name,
  color,
  code,
  createdAt,
  memberCount,
}: Team) => {
  const { mutate: deleteTeam } = useDeleteTeamMutation();
  const { withWorkspace } = useWorkspacePath();
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);

  const handleDeleteTeam = () => {
    deleteTeam(id);
    setIsDeleteOpen(false);
  };

  return (
    <RowWrapper className="last-of-type:border-b-0 md:px-6">
      <Flex align="center" gap={3}>
        <Box className="flex size-8 shrink-0 items-center justify-center rounded-[0.6rem] bg-surface-muted/80">
          <TeamColor color={color} />
        </Box>
        <Link href={withWorkspace(`/settings/workspace/teams/${id}`)}>
          <Text className="font-medium">{name}</Text>
          <Text className="text-[0.95rem]" color="muted">
            {code}
          </Text>
        </Link>
      </Flex>
      <Flex align="center" gap={3}>
        <Text className="md:w-32" color="muted">
          {memberCount} Member{memberCount === 1 ? "" : "s"}
        </Text>
        <Text as="span" className="hidden w-32 md:inline" color="muted">
          <TimeAgo timestamp={createdAt} />
        </Text>
        <Menu>
          <Menu.Button>
            <Button
              aria-label="More options"
              asIcon
              color="tertiary"
              rounded="full"
              size="sm"
              variant="naked"
            >
              <MoreHorizontalIcon className="h-5" />
            </Button>
          </Menu.Button>
          <Menu.Items align="end" className="w-44">
            <Menu.Group>
              <Menu.Item className="p-0">
                <Link
                  className="flex items-center gap-2 px-2 py-1.5"
                  href={withWorkspace(`/settings/workspace/teams/${id}`)}
                  prefetch
                >
                  <SettingsIcon />
                  Team settings
                </Link>
              </Menu.Item>
            </Menu.Group>
            <Menu.Separator />
            <Menu.Group>
              <Menu.Item
                onSelect={() => {
                  setIsDeleteOpen(true);
                }}
              >
                <DeleteIcon />
                Delete team...
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
      <ConfirmDialog
        confirmPhrase="delete team"
        description="Are you sure you want to delete this team? This action will remove all members and stories from the team and cannot be undone."
        isOpen={isDeleteOpen}
        onClose={() => {
          setIsDeleteOpen(false);
        }}
        onConfirm={handleDeleteTeam}
        title="Delete team"
      />
    </RowWrapper>
  );
};
