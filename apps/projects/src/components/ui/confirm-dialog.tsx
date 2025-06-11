"use client";

import { useState } from "react";
import { Box, Button, Dialog, Input, Text } from "ui";

type ConfirmDialogProps = {
  isOpen: boolean;
  onClose?: () => void;
  onCancel?: () => void;
  onConfirm: () => void;
  title: string;
  description: string;
  confirmText?: string;
  cancelText?: string;
  confirmPhrase?: string;
  isLoading?: boolean;
  loadingText?: string;
  hideClose?: boolean;
};

/**
 * ConfirmDialog
 *
 * @param isOpen - Whether the dialog is open
 * @param onClose - Function to call when the dialog is closed
 * @param onCancel - Function to call when the dialog is cancelled. if provided its only triggered on button click
 * @param onConfirm - Function to call when the dialog is confirmed
 * @param title - The title of the dialog
 * @param description - The description of the dialog
 * @param confirmText - The text of the confirm button
 * @param cancelText - The text of the cancel button
 * @param confirmPhrase - The phrase to confirm the action
 * @param isLoading - display a loading state on the confirm button
 * @param loadingText - The text of the loading button
 * @param hideClose - Whether to hide the close button
 */
export const ConfirmDialog = ({
  isOpen,
  onClose,
  onCancel,
  onConfirm,
  title,
  description,
  confirmText = "Confirm",
  cancelText = "Cancel",
  confirmPhrase = "",
  isLoading = false,
  loadingText = "Confirming...",
  hideClose = false,
}: ConfirmDialogProps) => {
  const [phrase, setPhrase] = useState("");
  return (
    <Dialog onOpenChange={onClose} open={isOpen}>
      <Dialog.Content hideClose={hideClose}>
        <Dialog.Header>
          <Dialog.Title className="px-6 pt-0.5 text-lg">{title}</Dialog.Title>
        </Dialog.Header>
        <Dialog.Body>
          <Text color="muted">{description}</Text>
          {confirmPhrase ? (
            <Box className="mt-3">
              <Text className="mb-2" color="muted">
                Please enter{" "}
                <Text as="span">&ldquo;{confirmPhrase}&rdquo;</Text> to confirm
              </Text>
              <Input
                className="rounded-lg"
                onChange={(e) => {
                  setPhrase(e.target.value);
                }}
                placeholder={confirmPhrase}
                type="text"
                value={phrase}
              />
            </Box>
          ) : null}
        </Dialog.Body>
        <Dialog.Footer className="justify-end gap-3 border-0 pt-2">
          <Button className="px-4" color="tertiary" onClick={onCancel}>
            {cancelText}
          </Button>
          <Button
            className="px-4"
            disabled={phrase !== confirmPhrase}
            loading={isLoading}
            loadingText={loadingText}
            onClick={onConfirm}
          >
            {confirmText}
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
