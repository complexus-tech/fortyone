"use client";

import type { ReactNode } from "react";
import { useState } from "react";
import { CheckIcon, DeleteIcon, EditIcon, OKRIcon } from "icons";
import { ContextMenu } from "ui";
import { ConfirmDialog } from "@/components/ui";
import { ContextMenuItem } from "@/components/ui/story/context-menu-item";
import { useTerminology, useUserRole } from "@/hooks";
import { useDeleteKeyResultMutation } from "@/modules/objectives/hooks/use-delete-key-result-mutation";
import type { KeyResult } from "@/modules/objectives/types";
import { UpdateKeyResultDialog } from "@/modules/objectives/stories/overview/update-key-result-dialog";

export const KeyResultContextMenu = ({
  children,
  keyResult,
  onOpenDetails,
}: {
  children: ReactNode;
  keyResult: KeyResult;
  onOpenDetails: () => void;
}) => {
  const { getTermDisplay } = useTerminology();
  const { userRole } = useUserRole();
  const { mutate: deleteKeyResult } = useDeleteKeyResultMutation();
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [updateMode, setUpdateMode] = useState<"other" | "progress">("other");
  const [isUpdateOpen, setIsUpdateOpen] = useState(false);
  const canEdit = userRole !== "guest";

  const openEditor = (mode: "other" | "progress") => {
    if (!canEdit) return;
    setUpdateMode(mode);
    setIsUpdateOpen(true);
  };

  return (
    <>
      <ContextMenu>
        <ContextMenu.Trigger>{children}</ContextMenu.Trigger>
        <ContextMenu.Items className="w-56">
          <ContextMenu.Group>
            <ContextMenuItem
              icon={<OKRIcon strokeWidth={2} />}
              label="Open details"
              onSelect={onOpenDetails}
            />
          </ContextMenu.Group>
          <ContextMenu.Separator />
          <ContextMenu.Group>
            <ContextMenuItem
              disabled={!canEdit}
              icon={<EditIcon />}
              label="Edit..."
              onSelect={() => {
                openEditor("other");
              }}
            />
            <ContextMenuItem
              disabled={!canEdit}
              icon={<CheckIcon />}
              label="Update progress..."
              onSelect={() => {
                openEditor("progress");
              }}
            />
          </ContextMenu.Group>
          <ContextMenu.Separator />
          <ContextMenu.Group>
            <ContextMenuItem
              disabled={!canEdit}
              icon={<DeleteIcon className="text-danger dark:text-danger" />}
              label="Delete"
              onSelect={() => {
                setIsDeleteOpen(true);
              }}
            />
          </ContextMenu.Group>
        </ContextMenu.Items>
      </ContextMenu>

      <ConfirmDialog
        confirmText="Yes, Delete"
        description={`Are you sure you want to delete this ${getTermDisplay(
          "keyResultTerm",
        )}? This action cannot be undone.`}
        isOpen={isDeleteOpen}
        onClose={() => {
          setIsDeleteOpen(false);
        }}
        onConfirm={() => {
          deleteKeyResult({
            keyResultId: keyResult.id,
            objectiveId: keyResult.objectiveId,
          });
        }}
        title={`Delete ${getTermDisplay("keyResultTerm", {
          capitalize: true,
        })}`}
      />
      <UpdateKeyResultDialog
        isOpen={isUpdateOpen}
        keyResult={keyResult}
        onOpenChange={setIsUpdateOpen}
        updateMode={updateMode}
      />
    </>
  );
};
