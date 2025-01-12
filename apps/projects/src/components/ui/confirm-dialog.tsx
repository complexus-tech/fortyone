"use client";

import { Button, Dialog, Text } from "ui";

type ConfirmDialogProps = {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title: string;
  description: string;
  confirmText?: string;
  cancelText?: string;
};

export const ConfirmDialog = ({
  isOpen,
  onClose,
  onConfirm,
  title,
  description,
  confirmText = "Confirm",
  cancelText = "Cancel",
}: ConfirmDialogProps) => {
  return (
    <Dialog onOpenChange={onClose} open={isOpen}>
      <Dialog.Content>
        <Dialog.Header>
          <Dialog.Title className="px-6 pt-0.5 text-lg">{title}</Dialog.Title>
        </Dialog.Header>
        <Dialog.Body>
          <Text color="muted">{description}</Text>
        </Dialog.Body>
        <Dialog.Footer className="justify-end gap-3 border-0 pt-2">
          <Button className="px-4" color="tertiary" onClick={onClose}>
            {cancelText}
          </Button>
          <Button onClick={onConfirm}>{confirmText}</Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
