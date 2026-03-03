"use client";

import { useRef, useState, type FormEvent } from "react";
import { Box, Button, Flex, Menu, Text, ColorPicker } from "ui";
import {
  CheckIcon,
  CloseIcon,
  DeleteIcon,
  DragIcon,
  EditIcon,
  MoreHorizontalIcon,
  SuccessIcon,
} from "icons";
import { useSortable } from "@dnd-kit/sortable";
import { useUpdateObjectiveStatusMutation } from "@/modules/objectives/hooks/statuses";
import type { ObjectiveStatus } from "@/modules/objectives/types";
import { ObjectiveStatusIcon } from "@/components/ui/objective-status-icon";

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
  const [isManualEditing, setIsManualEditing] = useState(false);
  const isEditing = Boolean(isNew) || isManualEditing;
  const [form, setForm] = useState({ name: status.name, color: status.color });
  const inputRef = useRef<HTMLInputElement>(null);

  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: status.id,
    disabled: isNew || isEditing,
  });

  const style = {
    transform: transform
      ? `translate3d(${transform.x}px, ${transform.y}px, 0)`
      : undefined,
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (isNew) {
      onCreate?.({
        ...status,
        name: form.name,
        color: form.color,
      });
      return;
    }
    updateMutation.mutate({
      statusId: status.id,
      payload: form,
    });
    setIsManualEditing(false);
  };

  const handleMakeDefault = () => {
    updateMutation.mutate({
      statusId: status.id,
      payload: { isDefault: true },
    });
  };

  const handleCancelEditing = () => {
    if (isNew) {
      onCreateCancel?.();
    } else {
      setIsManualEditing(false);
      setForm({ name: status.name, color: status.color });
    }
  };

  return (
    <form
      className="bg-surface-muted flex h-16 w-full items-center justify-between rounded-lg px-3"
      onSubmit={handleSubmit}
      ref={setNodeRef}
      style={style}
    >
      <Flex align="center" gap={2}>
        <button
          type="button"
          {...attributes}
          {...listeners}
          className="cursor-grab active:cursor-grabbing"
          style={{ touchAction: "none" }}
        >
          <DragIcon strokeWidth={4} />
        </button>

        {isEditing ? (
          <ColorPicker
            onChange={(value) => {
              setForm({ ...form, color: value });
            }}
            value={form.color}
          />
        ) : (
          <Box className="bg-surface/40 rounded-md p-2">
            <ObjectiveStatusIcon statusId={status.id} />
          </Box>
        )}

        <Box>
          <input
            className="placeholder:text-foreground w-max bg-transparent font-medium focus:outline-none"
            onChange={(e) => {
              setForm({ ...form, name: e.target.value });
            }}
            placeholder="Status name..."
            readOnly={!isEditing}
            ref={(element) => {
              inputRef.current = element;
            }}
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
                      setIsManualEditing(true);
                      requestAnimationFrame(() => {
                        inputRef.current?.focus();
                      });
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
