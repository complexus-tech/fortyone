"use client";

import type { FormEvent } from "react";
import { useState } from "react";
import { Box, Button, Dialog, Flex, Input, Text } from "ui";
import { StrategyDescriptionEditor } from "./strategy-description-editor";

type StrategyEditorDialogProps = {
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
  title: string;
  nameLabel: string;
  initialName?: string;
  initialDescription?: string | null;
  onSave: (name: string, description: string | null) => void;
  isPending?: boolean;
};

type StrategyEditorFormProps = Pick<
  StrategyEditorDialogProps,
  "isPending" | "nameLabel" | "onOpenChange" | "onSave" | "title"
> & {
  initialDescription: string;
  initialName: string;
};

const StrategyEditorForm = ({
  initialDescription,
  initialName,
  isPending,
  nameLabel,
  onOpenChange,
  onSave,
  title,
}: StrategyEditorFormProps) => {
  const [name, setName] = useState(initialName);
  const [description, setDescription] = useState<string | null>(
    initialDescription || null,
  );

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const trimmedName = name.trim();
    if (!trimmedName) return;
    onSave(trimmedName, description);
  };

  return (
    <form onSubmit={handleSubmit}>
      <Dialog.Header>
        <Dialog.Title className="px-6 pt-0.5 text-lg">{title}</Dialog.Title>
      </Dialog.Header>
      <Dialog.Body>
        <Flex className="gap-5" direction="column">
          <Input
            autoFocus
            className="bg-surface-muted/80 dark:bg-surface-muted/80"
            label={nameLabel}
            onChange={(event) => {
              setName(event.target.value);
            }}
            required
            value={name}
          />
          <Box>
            <Text className="mb-2">Description</Text>
            <StrategyDescriptionEditor
              ariaLabel="Strategic pillar description"
              className="min-h-32"
              content={initialDescription}
              contentClassName="min-h-32"
              editable
              onChange={setDescription}
              placeholder="Add context so the strategy is easy to understand"
            />
          </Box>
        </Flex>
      </Dialog.Body>
      <Dialog.Footer className="justify-end gap-2">
        <Button
          color="tertiary"
          onClick={() => {
            onOpenChange(false);
          }}
          type="button"
        >
          Cancel
        </Button>
        <Button
          color="invert"
          disabled={isPending || !name.trim()}
          type="submit"
        >
          Save
        </Button>
      </Dialog.Footer>
    </form>
  );
};

export const StrategyEditorDialog = ({
  initialDescription = "",
  initialName = "",
  isOpen,
  isPending,
  nameLabel,
  onOpenChange,
  onSave,
  title,
}: StrategyEditorDialogProps) => (
  <Dialog onOpenChange={onOpenChange} open={isOpen}>
    <Dialog.Content size="lg">
      <StrategyEditorForm
        initialDescription={initialDescription ?? ""}
        initialName={initialName}
        isPending={isPending}
        nameLabel={nameLabel}
        onOpenChange={onOpenChange}
        onSave={onSave}
        title={title}
      />
    </Dialog.Content>
  </Dialog>
);
