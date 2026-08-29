"use client";

import type { ReactNode } from "react";
import { useState } from "react";
import { Button, Dialog, Text, TextArea } from "ui";
import { toast } from "sonner";
import { HealthMenu } from "@/components/ui/health-menu";
import { ObjectiveHealthIcon } from "@/components/ui/objective-health-icon";
import { useTerminology } from "@/hooks";
import { openDialogAfterMenuClose } from "@/utils/menu-dialog-state";
import { useUpdateObjectiveMutation } from "../hooks/update-mutation";
import type { ObjectiveHealth } from "../types";

export const ObjectiveHealthEditor = ({
  children,
  health,
  objectiveId,
}: {
  children: ReactNode;
  health: ObjectiveHealth;
  objectiveId: string;
}) => {
  const { getTermDisplay } = useTerminology();
  const updateMutation = useUpdateObjectiveMutation();
  const [comment, setComment] = useState("");
  const [isCommentOpen, setIsCommentOpen] = useState(false);
  const [pendingHealth, setPendingHealth] = useState<ObjectiveHealth>(null);

  const closeHealthDialog = () => {
    setIsCommentOpen(false);
    setComment("");
    setPendingHealth(null);
  };

  const handleHealthUpdate = () => {
    if (!pendingHealth) {
      toast.warning("Validation error", {
        description: "Please select a health status",
      });
      return;
    }

    const trimmedComment = comment.trim();
    if (!trimmedComment) {
      toast.warning("Validation error", {
        description: "Please provide a comment",
      });
      return;
    }

    updateMutation.mutate({
      objectiveId,
      data: { health: pendingHealth, comment: trimmedComment },
    });
    closeHealthDialog();
  };

  return (
    <>
      <HealthMenu>
        <HealthMenu.Trigger>{children}</HealthMenu.Trigger>
        <HealthMenu.Items
          health={health}
          setHealth={(nextHealth) => {
            setPendingHealth(nextHealth);
            openDialogAfterMenuClose(setIsCommentOpen);
          }}
        />
      </HealthMenu>

      <Dialog
        onOpenChange={(open) => {
          if (!open) closeHealthDialog();
        }}
        open={isCommentOpen}
      >
        <Dialog.Content>
          <Dialog.Header>
            <Dialog.Title className="flex items-center gap-2 px-6 pt-0.5 text-lg">
              Change {getTermDisplay("objectiveTerm")} health to
              <ObjectiveHealthIcon health={pendingHealth} />
              {pendingHealth}
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Description>
            Please provide a brief comment explaining why you&apos;re changing
            the objective health status.
          </Dialog.Description>
          <Dialog.Body>
            <Text className="mt-3 mb-1.5" color="muted">
              Comment*
            </Text>
            <TextArea
              className="border-border/80 resize-none rounded-2xl border bg-transparent py-4 leading-normal"
              onChange={(event) => {
                setComment(event.target.value);
              }}
              placeholder={`e.g, We're on track to complete the ${getTermDisplay("objectiveTerm")} by the end of the quarter.`}
              rows={4}
              value={comment}
            />
          </Dialog.Body>
          <Dialog.Footer className="justify-end gap-2">
            <Button
              color="tertiary"
              onClick={closeHealthDialog}
              variant="outline"
            >
              Cancel
            </Button>
            <Button
              disabled={!comment.trim() || updateMutation.isPending}
              loading={updateMutation.isPending}
              onClick={handleHealthUpdate}
            >
              Update health
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
