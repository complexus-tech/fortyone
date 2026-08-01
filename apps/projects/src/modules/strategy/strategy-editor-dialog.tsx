"use client";

import type { FormEvent } from "react";
import { useState } from "react";
import { Button, Dialog, Flex, Input, TextArea } from "ui";

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
  const [description, setDescription] = useState(initialDescription);

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const trimmedName = name.trim();
    if (!trimmedName) return;
    onSave(trimmedName, description.trim() || null);
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
          <TextArea
            className="bg-surface-muted/60 dark:bg-surface-muted/60 min-h-28 resize-none py-3 leading-6"
            label="Description"
            onChange={(event) => {
              setDescription(event.target.value);
            }}
            placeholder="Add context so the strategy is easy to understand"
            value={description}
          />
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
    <Dialog.Content>
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
