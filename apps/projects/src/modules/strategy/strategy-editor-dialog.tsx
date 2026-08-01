"use client";

import type { FormEvent } from "react";
import { useEffect, useState } from "react";
import { Box, Button, Dialog, Flex, Input, Text } from "ui";

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

export const StrategyEditorDialog = ({
  isOpen,
  onOpenChange,
  title,
  nameLabel,
  initialName = "",
  initialDescription = "",
  onSave,
  isPending,
}: StrategyEditorDialogProps) => {
  const [name, setName] = useState(initialName);
  const [description, setDescription] = useState(initialDescription ?? "");

  useEffect(() => {
    if (isOpen) {
      setName(initialName);
      setDescription(initialDescription ?? "");
    }
  }, [initialDescription, initialName, isOpen]);

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const trimmedName = name.trim();
    if (!trimmedName) return;
    onSave(trimmedName, description.trim() || null);
  };

  return (
    <Dialog onOpenChange={onOpenChange} open={isOpen}>
      <Dialog.Content>
        <form onSubmit={handleSubmit}>
          <Dialog.Header>
            <Dialog.Title className="px-6 pt-0.5 text-lg">{title}</Dialog.Title>
          </Dialog.Header>
          <Dialog.Body>
            <Flex className="gap-5" direction="column">
              <Input
                autoFocus
                label={nameLabel}
                onChange={(event) => {
                  setName(event.target.value);
                }}
                required
                value={name}
              />
              <Box>
                <Text className="mb-[0.35rem]">Description</Text>
                <textarea
                  className="border-input bg-background focus-visible:ring-ring min-h-28 w-full resize-none rounded-md border px-4 py-3 outline-none focus-visible:ring-2"
                  onChange={(event) => {
                    setDescription(event.target.value);
                  }}
                  placeholder="Add context so the strategy is easy to understand"
                  value={description}
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
      </Dialog.Content>
    </Dialog>
  );
};
