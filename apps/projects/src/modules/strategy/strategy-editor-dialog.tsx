"use client";

import type { FormEvent } from "react";
import { useState } from "react";
import { Button, Dialog, Input } from "ui";
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
      <Dialog.Header className="flex items-center justify-between px-6 pt-5 pb-1">
        <Dialog.Title className="text-lg">{title}</Dialog.Title>
        <Dialog.Close />
      </Dialog.Header>
      <Dialog.Body className="pt-3 pb-3">
        <Input
          aria-label={nameLabel}
          autoFocus
          className="h-auto border-0 bg-transparent px-0 pt-1 pb-1 text-2xl leading-tight font-medium focus-visible:ring-0 dark:bg-transparent"
          maxLength={200}
          onChange={(event) => {
            setName(event.target.value);
          }}
          placeholder={nameLabel}
          required
          value={name}
        />
        <StrategyDescriptionEditor
          ariaLabel={`${nameLabel} description`}
          className="min-h-24"
          content={initialDescription}
          editable
          onChange={setDescription}
          placeholder="Add context so the strategy is easy to understand"
        />
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
    <Dialog.Content className="max-w-4xl" hideClose>
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
