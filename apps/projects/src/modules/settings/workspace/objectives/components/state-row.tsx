"use client";

import { useEffect, useRef, type FormEvent } from "react";
import { useState } from "react";
import { Box, Button, Flex, Menu, Text } from "ui";
import {
  CheckIcon,
  CloseIcon,
  DeleteIcon,
  // DragIcon,
  EditIcon,
  MoreHorizontalIcon,
  SuccessIcon,
} from "icons";
import { useUpdateObjectiveStatusMutation } from "@/modules/objectives/hooks/statuses";
import { StoryStatusIcon } from "@/components/ui";
import type { ObjectiveStatus } from "@/modules/objectives/types";

type StateRowProps = {
  status: ObjectiveStatus;
  onDelete: (status: ObjectiveStatus) => void;
  isNew?: boolean;
  onCreateCancel?: () => void;
  onCreate?: (status: ObjectiveStatus) => void;
};

export const StateRow = ({
  status,
  onDelete,
  isNew,
  onCreateCancel,
  onCreate,
}: StateRowProps) => {
  const updateMutation = useUpdateObjectiveStatusMutation();
  const [isEditing, setIsEditing] = useState(isNew);
  const [form, setForm] = useState({ name: status.name });
  const inputRef = useRef<HTMLInputElement>(null);

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (isNew) {
      onCreate?.({
        ...status,
        name: form.name,
        // color: form.color,
      });
      return;
    }
    updateMutation.mutate({
      statusId: status.id,
      payload: form,
    });
    setIsEditing(false);
  };

  const handleMakeDefault = () => {
    updateMutation.mutate({
      statusId: status.id,
      payload: { isDefault: true },
    });
  };

  const handleCancelEditing = () => {
    setIsEditing(false);
    if (isNew) {
      onCreateCancel?.();
    } else {
      setForm({ name: status.name });
    }
  };

  useEffect(() => {
    if (isEditing) {
      inputRef.current?.focus();
    }
  }, [isEditing]);

  return (
    <form
      className="flex h-16 w-full items-center justify-between rounded-[0.45rem] bg-gray-50 px-3 dark:bg-dark-100/70"
      onSubmit={handleSubmit}
    >
      <Flex align="center" gap={2}>
        {/* <DragIcon strokeWidth={4} /> */}
        <Box className="rounded-[0.4rem] bg-gray-100/60 p-2 dark:bg-dark-50/40">
          <StoryStatusIcon category={status.category} />
        </Box>
        <Box>
          <input
            className="w-max bg-transparent font-medium placeholder:text-gray focus:outline-none dark:placeholder:text-gray-300"
            onChange={(e) => {
              setForm({ ...form, name: e.target.value });
            }}
            placeholder="Status name..."
            readOnly={!isEditing}
            ref={inputRef}
            value={form.name}
          />
        </Box>
      </Flex>
      <Flex align="center" gap={2}>
        {isEditing ? (
          <>
            <Button
              color="tertiary"
              onClick={handleCancelEditing}
              rounded="full"
              size="sm"
              type="button"
              variant="naked"
            >
              <CloseIcon />
            </Button>
            <Button color="tertiary" rounded="full" size="sm" variant="naked">
              <CheckIcon />
            </Button>
          </>
        ) : (
          <Flex align="center" gap={2}>
            {status.isDefault ? <Text color="muted">Default</Text> : null}
            <Menu>
              <Menu.Button>
                <Button
                  color="tertiary"
                  rounded="full"
                  size="sm"
                  variant="naked"
                >
                  <MoreHorizontalIcon />
                </Button>
              </Menu.Button>
              <Menu.Items className="w-44">
                <Menu.Group>
                  {(status.category === "backlog" ||
                    status.category === "unstarted") &&
                    !status.isDefault && (
                      <Menu.Item onSelect={handleMakeDefault}>
                        <SuccessIcon className="h-[1.15rem]" />
                        Make default
                      </Menu.Item>
                    )}

                  <Menu.Item
                    onSelect={() => {
                      setIsEditing(true);
                    }}
                  >
                    <EditIcon className="h-[1.15rem]" />
                    Edit
                  </Menu.Item>
                  <Menu.Item
                    onSelect={() => {
                      onDelete(status);
                    }}
                  >
                    <DeleteIcon className="h-[1.15rem]" />
                    Delete...
                  </Menu.Item>
                </Menu.Group>
              </Menu.Items>
            </Menu>
          </Flex>
        )}
      </Flex>
    </form>
  );
};
