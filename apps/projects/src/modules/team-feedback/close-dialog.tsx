"use client";

import { useState } from "react";
import { Button, Dialog, Flex, Text, TextArea } from "ui";

const MAX_PUBLIC_EXPLANATION_LENGTH = 1000;

export const CloseTeamFeedbackDialog = ({
  isLoading,
  onCancel,
  onConfirm,
}: {
  isLoading: boolean;
  onCancel: () => void;
  onConfirm: (publicExplanation: string | null) => void;
}) => {
  const [publicExplanation, setPublicExplanation] = useState("");

  return (
    <Dialog
      onOpenChange={(open) => {
        if (!open && !isLoading) onCancel();
      }}
      open
    >
      <Dialog.Content className="max-w-lg">
        <Dialog.Header>
          <Dialog.Title className="px-6 pt-1 text-lg">
            Close feedback
          </Dialog.Title>
        </Dialog.Header>
        <Dialog.Body className="space-y-4">
          <Text className="leading-6" color="muted">
            Closing removes this item from the active feedback queue. It remains
            available in the feedback portal.
          </Text>
          <TextArea
            className="min-h-28 resize-y px-3 py-2 leading-6"
            label="Public explanation (optional)"
            maxLength={MAX_PUBLIC_EXPLANATION_LENGTH}
            onChange={(event) => {
              setPublicExplanation(event.target.value);
            }}
            placeholder="Explain why this request is being closed"
            value={publicExplanation}
          />
          <Text className="text-sm leading-5" color="muted">
            Followers are notified only when you include an explanation. Leave
            this blank to close the request quietly.
          </Text>
        </Dialog.Body>
        <Dialog.Footer>
          <Flex className="w-full" gap={3} justify="end">
            <Button color="tertiary" disabled={isLoading} onClick={onCancel}>
              Cancel
            </Button>
            <Button
              color="primary"
              loading={isLoading}
              loadingText="Closing..."
              onClick={() => {
                onConfirm(publicExplanation.trim() || null);
              }}
            >
              Close feedback
            </Button>
          </Flex>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
