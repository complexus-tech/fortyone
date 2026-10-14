"use client";

import { useState } from "react";
import {
  AiIcon,
  CopyIcon,
  DeleteIcon,
  MoreHorizontalIcon,
  StrategyIcon,
} from "icons";
import { Button, Flex, Menu, Tooltip } from "ui";
import { toast } from "sonner";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { useChatActions } from "@/context/chat-context";
import {
  useCopyToClipboard,
  useTerminology,
  useUserRole,
  useWorkspacePath,
} from "@/hooks";
import { useIsAdminOrOwner } from "@/hooks/owner";
import { useDeleteObjectiveMutation } from "../hooks/delete-mutation";
import type { Objective } from "../types";

export const HeaderActions = ({
  objective,
}: {
  objective: Pick<Objective, "id" | "name" | "createdBy">;
}) => {
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [, copyText] = useCopyToClipboard();
  const { getTermDisplay } = useTerminology();
  const { withWorkspace } = useWorkspacePath();
  const { userRole } = useUserRole();
  const { isAdminOrOwner } = useIsAdminOrOwner(objective.createdBy);
  const { openChat } = useChatActions();
  const deleteMutation = useDeleteObjectiveMutation();

  const handleCopyLink = async () => {
    const copied = await copyText(window.location.href);
    if (!copied) {
      toast.error("Unable to copy link", {
        description: "Please try again or copy the link from your address bar.",
      });
      return;
    }

    toast.info("Success", {
      description: `${getTermDisplay("objectiveTerm", { capitalize: true })} link copied to clipboard`,
    });
  };

  return (
    <>
      <Flex align="center" className="shrink-0" gap={1}>
        {userRole !== "guest" && (
          <Tooltip title="Intelligently add work items to the objective">
            <Button
              aria-label="Smart fill"
              color="tertiary"
              leftIcon={<AiIcon className="text-primary size-5" />}
              onClick={() => {
                openChat(
                  `Suggest work items that can be added to the objective "${objective.name}" from the team`,
                );
              }}
              size="sm"
              variant="naked"
            >
              <span className="hidden 2xl:inline">Smart fill</span>
            </Button>
          </Tooltip>
        )}
        <Tooltip title="Open strategy map">
          <Button
            color="tertiary"
            href={withWorkspace("/strategy")}
            leftIcon={<StrategyIcon className="size-5" />}
            size="sm"
            variant="naked"
          >
            <span className="sr-only 2xl:not-sr-only">Map view</span>
          </Button>
        </Tooltip>
        <Menu>
          <Menu.Button>
            <Button
              aria-label="More objective actions"
              asIcon
              color="tertiary"
              size="sm"
              variant="naked"
            >
              <MoreHorizontalIcon className="size-5" />
            </Button>
          </Menu.Button>
          <Menu.Items align="end" className="w-48">
            <Menu.Group>
              <Menu.Item onSelect={handleCopyLink}>
                <CopyIcon className="size-4" />
                Copy link
              </Menu.Item>
            </Menu.Group>
            {isAdminOrOwner ? (
              <>
                <Menu.Separator />
                <Menu.Group>
                  <Menu.Item
                    onSelect={() => {
                      setIsDeleteOpen(true);
                    }}
                  >
                    <DeleteIcon className="size-4" />
                    Delete
                  </Menu.Item>
                </Menu.Group>
              </>
            ) : null}
          </Menu.Items>
        </Menu>
      </Flex>
      <ConfirmDialog
        confirmText="Yes, Delete"
        description={`Are you sure you want to delete this ${getTermDisplay("objectiveTerm")}? This action is irreversible.`}
        isLoading={deleteMutation.isPending}
        isOpen={isDeleteOpen}
        loadingText="Deleting..."
        onClose={() => {
          setIsDeleteOpen(false);
        }}
        onConfirm={() => {
          if (!isAdminOrOwner || deleteMutation.isPending) return;
          deleteMutation.mutate(objective.id);
        }}
        title={`Delete ${getTermDisplay("objectiveTerm")}`}
      />
    </>
  );
};
