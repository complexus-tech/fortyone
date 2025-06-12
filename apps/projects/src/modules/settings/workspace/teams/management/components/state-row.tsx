"use client";

import { useState, useRef, useEffect } from "react";
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
import type { FormEvent } from "react";
import { toast } from "sonner";
import { useSortable } from "@dnd-kit/sortable";
import { cn } from "lib";
import { StoryStatusIcon } from "@/components/ui";
import { useUpdateStateMutation } from "@/lib/hooks/states/update-mutation";
import type { State } from "@/types/states";

type StateRowProps = {
  state: State;
  onDelete: (state: State) => void;
  isNew?: boolean;
  onCreateCancel?: () => void;
  onCreate?: (state: State) => void;
  storyCount?: number;
};

export const StateRow = ({
  state,
  onDelete,
  isNew,
  onCreateCancel,
  onCreate,
  storyCount,
}: StateRowProps) => {
  const updateMutation = useUpdateStateMutation();
  const [isEditing, setIsEditing] = useState(isNew);
  const [form, setForm] = useState({ name: state.name, color: state.color });
  const inputRef = useRef<HTMLInputElement>(null);

  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: state.id,
    disabled: isNew || isEditing,
  });

  const style = {
    transform: transform
      ? `translate3d(${transform.x}px, ${transform.y}px, 0)`
      : undefined,
    transition,
  };

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (isNew) {
      onCreate?.({
        ...state,
        name: form.name,
        color: form.color,
      });
      return;
    }
    updateMutation.mutate({
      stateId: state.id,
      payload: form,
    });
    setIsEditing(false);
  };

  const handleCancelEditing = () => {
    setIsEditing(false);
    if (isNew) {
      onCreateCancel?.();
    } else {
      setForm({ name: state.name, color: state.color });
    }
  };

  const handleMakeDefault = () => {
    updateMutation.mutate({
      stateId: state.id,
      payload: { isDefault: true },
    });
  };

  useEffect(() => {
    if (isEditing) {
      inputRef.current?.focus();
    }
  }, [isEditing]);

  return (
    <form
      className={cn(
        "flex h-16 w-full items-center justify-between rounded-[0.45rem] bg-gray-50 px-3 dark:bg-dark-100/70",
        {
          "opacity-80 backdrop-blur": isDragging,
          "shadow-lg": isDragging,
        },
      )}
      onSubmit={handleSubmit}
      ref={setNodeRef}
      style={style}
    >
      <Flex align="center" gap={2}>
        <DragIcon
          className={cn("cursor-grab", {
            "cursor-grabbing": isDragging,
            "opacity-80": isNew || isEditing,
            "cursor-not-allowed": isNew || isEditing,
          })}
          strokeWidth={4}
          {...attributes}
          {...listeners}
        />
        {isEditing ? (
          <ColorPicker
            onChange={(value) => {
              setForm({ ...form, color: value });
            }}
            value={form.color}
          />
        ) : (
          <Box className="rounded-[0.4rem] bg-gray-100/60 p-2 dark:bg-dark-50/40">
            <StoryStatusIcon category={state.category} statusId={state.id} />
          </Box>
        )}

        <Box>
          <input
            className={cn(
              "bg-transparent font-medium placeholder:text-gray focus:outline-none dark:placeholder:text-gray-300",
              {
                "my-0.5 rounded-lg border border-gray-100 bg-white/50 px-3 py-1 dark:border-gray-300/20 dark:bg-dark/10":
                  isEditing,
              },
            )}
            onChange={(e) => {
              setForm({ ...form, name: e.target.value });
            }}
            placeholder="State name..."
            readOnly={!isEditing}
            ref={inputRef}
            value={form.name}
          />
          {storyCount && !isEditing ? (
            <Text color="muted" fontSize="sm">
              {storyCount} {storyCount === 1 ? "story" : "stories"}
            </Text>
          ) : null}
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
            {state.isDefault ? <Text color="muted">Default</Text> : null}
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
                  {(state.category === "backlog" ||
                    state.category === "unstarted") &&
                    !state.isDefault && (
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
                      if (storyCount) {
                        toast.warning(`Cannot delete status "${form.name}"`, {
                          description:
                            "Move all stories to another status before deleting, or delete the stories first.",
                        });
                      } else {
                        onDelete(state);
                      }
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
