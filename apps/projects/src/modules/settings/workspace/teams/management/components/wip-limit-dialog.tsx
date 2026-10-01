"use client";

import { useState } from "react";
import { Button, Dialog, Flex, Input, Text } from "ui";
import type { State } from "@/types/states";
import { useUpdateStateMutation } from "@/lib/hooks/states/update-mutation";

export const WipLimitDialog = ({
  state,
  onClose,
}: {
  state: State;
  onClose: () => void;
}) => {
  const [value, setValue] = useState(state.wipLimit?.toString() ?? "");
  const update = useUpdateStateMutation();
  const limit = value.trim() === "" ? 0 : Number(value);
  const valid = Number.isInteger(limit) && limit >= 0 && limit <= 10000;

  return (
    <Dialog
      onOpenChange={(open) => {
        if (!open && !update.isPending) onClose();
      }}
      open
    >
      <Dialog.Content
        aria-busy={update.isPending}
        className="flex max-h-[calc(100dvh-15vw-1rem)] flex-col md:max-h-[calc(100dvh-10vw-1rem)]"
        hideClose={update.isPending}
        onEscapeKeyDown={(event) => {
          if (update.isPending) event.preventDefault();
        }}
        onInteractOutside={(event) => {
          if (update.isPending) event.preventDefault();
        }}
      >
        <Dialog.Header className="shrink-0 px-6 py-5">
          <Dialog.Title className="pr-8 text-lg">
            Work in progress limit
          </Dialog.Title>
          <Dialog.Description className="mt-2 px-0 text-base leading-6">
            Warn when {state.name} exceeds the limit. Board filters do not
            change the count.
          </Dialog.Description>
        </Dialog.Header>
        <Dialog.Body className="max-h-none min-h-0 flex-1 space-y-4">
          <Input
            autoFocus
            disabled={update.isPending}
            label="Maximum active items"
            max={10000}
            min={0}
            onChange={(event) => {
              setValue(event.target.value);
            }}
            placeholder="No limit"
            step={1}
            type="number"
            value={value}
          />
          <Text color="muted">
            Leave empty or use zero to remove the limit.
          </Text>
          {update.error ? (
            <Text color="danger" role="alert">
              {update.error.message}
            </Text>
          ) : null}
        </Dialog.Body>
        <Dialog.Footer className="shrink-0 py-4" justify="end">
          <Flex gap={2} justify="end">
            <Button
              color="tertiary"
              disabled={update.isPending}
              onClick={onClose}
            >
              Cancel
            </Button>
            <Button
              disabled={!valid || update.isPending}
              loading={update.isPending}
              onClick={() => {
                update.mutate(
                  { stateId: state.id, payload: { wipLimit: limit } },
                  { onSuccess: onClose },
                );
              }}
            >
              Save limit
            </Button>
          </Flex>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
