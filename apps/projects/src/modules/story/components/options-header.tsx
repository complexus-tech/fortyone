"use client";
import { Button, Container, Dialog, Flex, Text, Tooltip } from "ui";
import { CopyIcon, DeleteIcon, GitIcon, MaximizeIcon, UndoIcon } from "icons";
import { toast } from "sonner";
import { useState } from "react";
import { useHotkeys } from "react-hotkeys-hook";
import { useCopyToClipboard, useTerminology, useUserRole, useWorkspacePath } from "@/hooks";
import { useStoryById } from "@/modules/story/hooks/story";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useRestoreStoryMutation } from "@/modules/story/hooks/restore-mutation";
import { useProfile } from "@/lib/hooks/profile";
import { slugify } from "@/utils";
import { useAutomationPreferences } from "@/lib/hooks/users/preferences";
import { useTeamStatuses } from "@/lib/hooks/statuses";
import { MobileMenuButton } from "@/components/shared";
import { useDeleteStoryMutation } from "../hooks/delete-mutation";
import { useUpdateStoryMutation } from "../hooks/update-mutation";

const isFortyOneApp = process.env.NEXT_PUBLIC_DOMAIN === "fortyone.app";

export const OptionsHeader = ({
  isAdminOrOwner,
  storyId,
  isNotifications,
  isDialog,
}: {
  isAdminOrOwner: boolean;
  storyId: string;
  isNotifications?: boolean;
  isDialog?: boolean;
}) => {
  const { data: currentUser } = useProfile();
  const { data } = useStoryById(storyId);
  const { id, teamId, title, sequenceId, deletedAt, assigneeId, statusId } =
    data!;
  const [isOpen, setIsOpen] = useState(false);
  const { data: teams = [] } = useTeams();
  const [_, copyText] = useCopyToClipboard();
  const team = teams.find((team) => team.id === teamId);
  const code = team ? team.code : "";
  const isDeleted = Boolean(deletedAt);
  const { mutate: deleteStory } = useDeleteStoryMutation();
  const { mutate: updateStory } = useUpdateStoryMutation();
  const { data: statuses } = useTeamStatuses(teamId);
  const { mutateAsync } = useRestoreStoryMutation();
  const { getTermDisplay } = useTerminology();
  const { data: automationPreferences } = useAutomationPreferences();
  const { userRole } = useUserRole();
  const { withWorkspace, workspaceSlug } = useWorkspacePath();

  const generateGitBranchName = () => {
    const branchName =
      `${currentUser?.username}/${code}-${sequenceId}-${slugify(
        title.slice(0, 32),
      )}`.toLowerCase();
    return branchName.replace(/-$/, "");
  };

  const getStoryUrl = () => {
    if (isFortyOneApp) {
      return `${window.location.origin}/story/${id}`;
    }
    return `${window.location.origin}/${workspaceSlug}/story/${id}`;
  }

  const copyBranchName = async () => {
    await copyText(generateGitBranchName());
    toast.info(generateGitBranchName(), {
      description: "Git branch name copied to clipboard",
    });

    const startedStatuses =
      statuses?.filter((status) => status.category === "started") || [];
    const updatePayload: { assigneeId?: string; statusId?: string } = {};
    const currentStatusCategory = statuses?.find(
      (status) => status.id === statusId,
    )?.category;
    if (
      automationPreferences?.assignSelfOnBranchCopy &&
      assigneeId !== currentUser?.id
    ) {
      updatePayload.assigneeId = currentUser?.id;
    }
    if (
      automationPreferences?.moveStoryToStartedOnBranch &&
      startedStatuses.length > 0 &&
      currentStatusCategory !== "started"
    ) {
      updatePayload.statusId = startedStatuses[0].id;
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

  useHotkeys("backspace, delete", () => {
    if (isAdminOrOwner) {
      setIsOpen(true);
    }
  });
  return (
    <>
      <Container className="flex h-16 w-full items-center justify-between border-b-[0.5px] border-border d md:border-b-0 md:px-6">
        <Flex align="center" gap={2}>
          <MobileMenuButton />
          <Text color="muted" fontWeight="semibold" transform="uppercase">
            {code ? (
              <>
                {code}-{sequenceId}
              </>
            ) : null}
          </Text>
        </Flex>
        <Flex align="center" gap={2}>
          {isDialog ? (
            <Tooltip side="bottom" title="Fullscreen">
              <span>
                <Button
                  asIcon
                  color="tertiary"
                  href={withWorkspace(`/story/${id}/${slugify(title)}`)}
                  leftIcon={
                    <MaximizeIcon className="h-5" strokeWidth={2.5} />
                  }
                  variant="naked"
                >
                  <span className="sr-only">Fullscreen</span>
                </Button>
              </span>
            </Tooltip>
          ) : null}

          <Tooltip
            title={`Copy ${getTermDisplay("storyTerm", { capitalize: true })} link`}
          >
            <Button
              color="tertiary"
              leftIcon={<CopyIcon />}
              onClick={async () => {
                await copyText(getStoryUrl());
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
          {userRole !== "guest" && (
            <Tooltip title="Copy git branch name">
              <Button
                color="tertiary"
                disabled={!code}
                leftIcon={<GitIcon />}
                onClick={copyBranchName}
                variant="naked"
              >
                <span className="sr-only">Copy git branch name</span>
              </Button>
            </Tooltip>
          )}

          {!isNotifications && !isDialog ? (
            <>
              {isDeleted ? (
                <Tooltip
                  title={
                    isAdminOrOwner
                      ? `Restore ${getTermDisplay("storyTerm", {
                        capitalize: true,
                      })}`
                      : `You are not allowed to restore this ${getTermDisplay("storyTerm")}`
                  }
                >
                  <Button
                    color="tertiary"
                    disabled={!isAdminOrOwner || !code}
                    leftIcon={<UndoIcon />}
                    onClick={() => {
                      if (isAdminOrOwner) {
                        restoreStory();
                      }
                    }}
                    variant="naked"
                  >
                    <span className="sr-only">
                      Restore{" "}
                      {getTermDisplay("storyTerm", {
                        capitalize: true,
                      })}
                    </span>
                  </Button>
                </Tooltip>
              ) : (
                <Tooltip
                  title={
                    isAdminOrOwner
                      ? `Delete ${getTermDisplay("storyTerm", {
                        capitalize: true,
                      })}`
                      : `You are not allowed to delete this ${getTermDisplay("storyTerm")}`
                  }
                >
                  <Button
                    color="tertiary"
                    disabled={!isAdminOrOwner || !code}
                    leftIcon={<DeleteIcon />}
                    onClick={() => {
                      if (isAdminOrOwner) {
                        setIsOpen(true);
                      }
                    }}
                    variant="naked"
                  >
                    <span className="sr-only">
                      Delete{" "}
                      {getTermDisplay("storyTerm", {
                        capitalize: true,
                      })}
                    </span>
                  </Button>
                </Tooltip>
              )}
            </>
          ) : null}
        </Flex>
      </Container>
      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content>
          <Dialog.Header className="flex items-center justify-between px-6 pt-6">
            <Dialog.Title className="flex items-center gap-1 text-lg">
              Are you sure you want to delete this {getTermDisplay("storyTerm")}
              ?
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="pt-0">
            <Text color="muted">
              This {getTermDisplay("storyTerm")} will be moved to the recycle
              bin and will be permanently deleted after 30 days. You can restore
              it at any time before that.
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
