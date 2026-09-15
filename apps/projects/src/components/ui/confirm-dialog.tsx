"use client";

import { useId, useState } from "react";
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
  normalizeConfirmPhrase?: boolean;
  errorMessage?: string;
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
 * @param normalizeConfirmPhrase - Ignore casing and surrounding whitespace when matching the phrase
 * @param errorMessage - An error to show without dismissing the dialog
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
  normalizeConfirmPhrase = false,
  errorMessage,
  isLoading = false,
  loadingText = "Confirming...",
  hideClose = false,
}: ConfirmDialogProps) => {
  const [phrase, setPhrase] = useState("");
  const phraseInputId = useId();
  const hasConfirmedPhrase = normalizeConfirmPhrase
    ? phrase.trim().toLowerCase() === confirmPhrase.trim().toLowerCase()
    : phrase === confirmPhrase;

  const handleClose = () => {
    if (isLoading) return;
    setPhrase("");
    onClose?.();
  };

  return (
    <Dialog
      onOpenChange={(open) => {
        if (!open) handleClose();
      }}
      open={isOpen}
    >
      <Dialog.Content
        aria-busy={isLoading}
        hideClose={hideClose || isLoading}
        onEscapeKeyDown={(event) => {
          if (isLoading) event.preventDefault();
        }}
        onInteractOutside={(event) => {
          if (isLoading) event.preventDefault();
        }}
      >
        <Dialog.Header>
          <Dialog.Title className="px-6 pt-0.5 text-lg">{title}</Dialog.Title>
        </Dialog.Header>
        <Dialog.Body>
          <Dialog.Description asChild>
            <Text className="px-0" color="muted" fontSize="md">
              {description}
            </Text>
          </Dialog.Description>
          {confirmPhrase ? (
            <Box className="mt-3">
              <label
                className="text-text-muted mb-2 block"
                htmlFor={phraseInputId}
              >
                Please enter{" "}
                <Text as="span">&ldquo;{confirmPhrase}&rdquo;</Text> to confirm
              </label>
              <Input
                className="rounded-lg"
                disabled={isLoading}
                id={phraseInputId}
                onChange={(e) => {
                  setPhrase(e.target.value);
                }}
                placeholder={confirmPhrase}
                type="text"
                value={phrase}
              />
            </Box>
          ) : null}
          {errorMessage ? (
            <Text className="mt-3" color="danger" role="alert">
              {errorMessage}
            </Text>
          ) : null}
        </Dialog.Body>
        <Dialog.Footer className="justify-end gap-3 border-0 pt-2">
          <Button
            className="px-4"
            color="tertiary"
            disabled={isLoading}
            onClick={() => {
              if (isLoading) return;
              onCancel?.();
              handleClose();
            }}
          >
            {cancelText}
          </Button>
          <Button
            className="px-4"
            disabled={isLoading || !hasConfirmedPhrase}
            loading={isLoading}
            loadingText={loadingText}
            onClick={() => {
              if (!isLoading && hasConfirmedPhrase) onConfirm();
            }}
          >
            {confirmText}
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
