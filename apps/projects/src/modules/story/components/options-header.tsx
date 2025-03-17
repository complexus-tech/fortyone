"use client";
import { Button, Container, Dialog, Flex, Text, Tooltip } from "ui";
import { CopyIcon, DeleteIcon, GitIcon, UndoIcon } from "icons";
import { toast } from "sonner";
import { useState } from "react";
import { useCopyToClipboard, useTerminology } from "@/hooks";
import { useStoryById } from "@/modules/story/hooks/story";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useRestoreStoryMutation } from "@/modules/story/hooks/restore-mutation";
import { useProfile } from "@/lib/hooks/profile";
import { slugify } from "@/utils";
import { useAutomationPreferences } from "@/lib/hooks/users/preferences";
import { useTeamStatuses } from "@/lib/hooks/statuses";
import { useDeleteStoryMutation } from "../hooks/delete-mutation";
import { useUpdateStoryMutation } from "../hooks/update-mutation";

export const OptionsHeader = ({
  isAdminOrOwner,
  storyId,
}: {
  isAdminOrOwner: boolean;
  storyId: string;
}) => {
  const { data: currentUser } = useProfile();
  const { data } = useStoryById(storyId);
  const { id, teamId, title, sequenceId, deletedAt } = data!;
  const [isOpen, setIsOpen] = useState(false);
  const { data: teams = [] } = useTeams();
  const [_, copyText] = useCopyToClipboard();
  const { code } = teams.find((team) => team.id === teamId)!;
  const isDeleted = Boolean(deletedAt);
  const { mutate: deleteStory } = useDeleteStoryMutation();
  const { mutate: updateStory } = useUpdateStoryMutation();
  const { data: statuses } = useTeamStatuses(teamId);
  const { mutateAsync } = useRestoreStoryMutation();
  const { getTermDisplay } = useTerminology();
  const { data: automationPreferences } = useAutomationPreferences();

  const generateGitBranchName = () => {
    const branchName =
      `${currentUser?.username}/${code}-${sequenceId}-${slugify(
        title.slice(0, 32),
      )}`.toLowerCase();
    return branchName.replace(/-$/, "");
  };

  const copyBranchName = async () => {
    await copyText(generateGitBranchName());
    toast.info(generateGitBranchName(), {
      description: "Git branch name copied to clipboard",
    });

    const startedStatus = statuses?.find(
      (status) => status.category === "started",
    );
    const updatePayload: { assigneeId?: string; statusId?: string } = {};
    if (automationPreferences?.assignSelfOnBranchCopy) {
      updatePayload.assigneeId = currentUser?.id;
    }
    if (automationPreferences?.moveStoryToStartedOnBranch && startedStatus) {
      updatePayload.statusId = startedStatus.id;
    }
    if (Object.keys(updatePayload).length > 0) {
      updateStory({ storyId: id, payload: updatePayload });
    }
  };

  const handleDelete = () => {
    deleteStory(id);
    setIsOpen(false);
  };

  const restoreStory = async () => {
    await mutateAsync(id);
  };
  return (
    <>
      <Container className="flex h-16 w-full items-center justify-between md:px-6">
        <Text color="muted" fontWeight="semibold" transform="uppercase">
          {code}-{sequenceId}
        </Text>
        <Flex gap={2}>
          <Tooltip
            title={`Copy ${getTermDisplay("storyTerm", { capitalize: true })} link`}
          >
            <Button
              color="tertiary"
              leftIcon={<CopyIcon />}
              onClick={async () => {
                await copyText(window.location.href);
                toast.info("Success", {
                  description: `${getTermDisplay("storyTerm", { capitalize: true })} link copied to clipboard`,
                });
              }}
              suppressHydrationWarning
              variant="naked"
            >
              <span className="sr-only">
                Copy {getTermDisplay("storyTerm")} link
              </span>
            </Button>
          </Tooltip>
          <Tooltip title="Copy git branch name">
            <Button
              color="tertiary"
              leftIcon={<GitIcon />}
              onClick={copyBranchName}
              variant="naked"
            >
              <span className="sr-only">Copy git branch name</span>
            </Button>
          </Tooltip>

          {isDeleted ? (
            <Tooltip
              title={
                isAdminOrOwner
                  ? "Restore Story"
                  : "You are not allowed to restore this story"
              }
            >
              <Button
                color="tertiary"
                disabled={!isAdminOrOwner}
                leftIcon={<UndoIcon />}
                onClick={() => {
                  if (isAdminOrOwner) {
                    restoreStory();
                  }
                }}
                variant="naked"
              >
                <span className="sr-only">Restore story</span>
              </Button>
            </Tooltip>
          ) : (
            <Tooltip
              title={
                isAdminOrOwner
                  ? "Delete Story"
                  : "You are not allowed to delete this story"
              }
            >
              <Button
                color="tertiary"
                disabled={!isAdminOrOwner}
                leftIcon={<DeleteIcon />}
                onClick={() => {
                  if (isAdminOrOwner) {
                    setIsOpen(true);
                  }
                }}
                variant="naked"
              >
                <span className="sr-only">Delete story</span>
              </Button>
            </Tooltip>
          )}
        </Flex>
      </Container>
      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content>
          <Dialog.Header className="flex items-center justify-between px-6 pt-6">
            <Dialog.Title className="flex items-center gap-1 text-lg">
              Are you sure you want to delete this story?
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="pt-0">
            <Text color="muted">
              This story will be moved to the recycle bin and will be
              permanently deleted after 30 days. You can restore it at any time
              before that.
            </Text>
            <Flex align="center" className="mt-4" gap={2} justify="end">
              <Button
                color="tertiary"
                onClick={() => {
                  setIsOpen(false);
                }}
              >
                Cancel
              </Button>
              <Button
                leftIcon={
                  <DeleteIcon className="text-white dark:text-gray-200" />
                }
                onClick={handleDelete}
              >
                Delete
              </Button>
            </Flex>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
