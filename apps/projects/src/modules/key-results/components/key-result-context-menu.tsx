"use client";

import type { ReactNode } from "react";
import { useState } from "react";
import Link from "next/link";
import {
  CheckIcon,
  DeleteIcon,
  EditIcon,
  ExternalLinkIcon,
  LinkIcon,
  OKRIcon,
} from "icons";
import { toast } from "sonner";
import { ContextMenu, Text } from "ui";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { ContextMenuItem } from "@/components/ui/story/context-menu-item";
import { useTerminology, useUserRole } from "@/hooks";
import { useCopyToClipboard } from "@/hooks/clipboard";
import { useDeleteKeyResultMutation } from "@/modules/objectives/hooks/use-delete-key-result-mutation";
import type { KeyResult } from "@/modules/objectives/types";
import { UpdateKeyResultDialog } from "@/modules/objectives/stories/overview/update-key-result-dialog";

export const KeyResultContextMenu = ({
  children,
  keyResult,
  objectivePath,
  onOpenDetails,
}: {
  children: ReactNode;
  keyResult: KeyResult;
  objectivePath?: string;
  onOpenDetails: () => void;
}) => {
  const { getTermDisplay } = useTerminology();
  const { userRole } = useUserRole();
  const [, copyToClipboard] = useCopyToClipboard();
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

  const copyLink = async () => {
    if (!objectivePath) return;

    const url = new URL(objectivePath, window.location.origin);
    url.searchParams.set("tab", "overview");
    url.searchParams.set("keyResultId", keyResult.id);
    const copied = await copyToClipboard(url.toString());

    if (copied) {
      toast.success("Link copied to clipboard");
    } else {
      toast.error("Unable to copy link", {
        description: "Please allow clipboard access and try again.",
      });
    }
  };

  return (
    <>
      <ContextMenu>
        <ContextMenu.Trigger
          onKeyDown={(event) => {
            if (
              event.defaultPrevented ||
              !(
                event.key === "ContextMenu" ||
                (event.shiftKey && event.key === "F10")
              )
            ) {
              return;
            }

            event.preventDefault();
            event.stopPropagation();
            const bounds = event.currentTarget.getBoundingClientRect();
            // Radix anchors context menus to pointer coordinates, including
            // when keyboard users request the focused card's actions.
            event.currentTarget.dispatchEvent(
              new MouseEvent("contextmenu", {
                bubbles: true,
                cancelable: true,
                clientX: bounds.left + Math.min(12, bounds.width / 2),
                clientY: bounds.top + Math.min(12, bounds.height / 2),
              }),
            );
          }}
        >
          <div>{children}</div>
        </ContextMenu.Trigger>
        <ContextMenu.Items className="w-56">
          <ContextMenu.Group>
            <ContextMenuItem
              icon={<OKRIcon strokeWidth={2} />}
              label="Open details"
              onSelect={onOpenDetails}
            />
            {objectivePath ? (
              <>
                <ContextMenu.Item asChild>
                  <Link href={objectivePath} prefetch={false}>
                    <span className="text-text-muted flex w-6 shrink-0">
                      <ExternalLinkIcon />
                    </span>
                    <Text>Open objective</Text>
                  </Link>
                </ContextMenu.Item>
                <ContextMenuItem
                  icon={<LinkIcon />}
                  label="Copy link"
                  onSelect={() => {
                    void copyLink();
                  }}
                />
              </>
            ) : null}
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
                if (canEdit) setIsDeleteOpen(true);
              }}
            />
          </ContextMenu.Group>
        </ContextMenu.Items>
      </ContextMenu>

      {isDeleteOpen ? (
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
            if (!canEdit) return;
            deleteKeyResult({
              keyResultId: keyResult.id,
              objectiveId: keyResult.objectiveId,
            });
          }}
          title={`Delete ${getTermDisplay("keyResultTerm", {
            capitalize: true,
          })}`}
        />
      ) : null}
      {isUpdateOpen ? (
        <UpdateKeyResultDialog
          isOpen={isUpdateOpen}
          keyResult={keyResult}
          onOpenChange={setIsUpdateOpen}
          updateMode={updateMode}
        />
      ) : null}
    </>
  );
};
