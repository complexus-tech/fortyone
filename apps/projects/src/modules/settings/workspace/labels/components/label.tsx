import { Box, Flex, Button, ColorPicker, Text, Menu } from "ui";
import {
  CheckIcon,
  CloseIcon,
  DeleteIcon,
  EditIcon,
  MoreHorizontalIcon,
} from "icons";
import type { FormEvent } from "react";
import { useRef, useState, useEffect } from "react";
import { format } from "date-fns";
import type { Label } from "@/types";
import { ConfirmDialog } from "@/components/ui";
import { useCreateLabelMutation } from "@/lib/hooks/create-label-mutation";

export const WorkspaceLabel = ({
  id,
  name,
  color,
  createdAt,
  setIsCreateOpen,
}: Partial<Label> & { setIsCreateOpen?: (value: boolean) => void }) => {
  const { mutateAsync: createLabel } = useCreateLabelMutation();
  const isNew = id === "new";
  const [isEditing, setIsEditing] = useState(isNew);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [form, setForm] = useState({ name: name || "", color: color || "" });
  const inputRef = useRef<HTMLInputElement>(null);

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setIsEditing(false);
    if (isNew) {
      await createLabel({ ...form }).then(() => {
        setIsCreateOpen?.(false);
      });
    } else {
      // TODO: update label
    }
  };

  const handleCancelEditing = () => {
    setIsEditing(false);
    setIsCreateOpen?.(false);
  };

  const handleDeleteLabel = () => {
    setIsDeleteOpen(false);
  };

  useEffect(() => {
    if (isEditing) {
      inputRef.current?.focus();
    }
  }, [isEditing]);

  return (
    <Box className="px-6 py-2.5 transition-colors hover:bg-gray-50 dark:hover:bg-dark-300">
      <form
        className="flex items-center justify-between"
        onSubmit={handleSubmit}
      >
        <Flex align="center" gap={2}>
          <ColorPicker
            onChange={(value) => {
              setForm({ ...form, color: value });
            }}
            onClick={() => {
              setIsEditing(true);
            }}
            value={form.color}
          />
          <input
            className="bg-transparent font-medium placeholder:text-gray focus:outline-none dark:placeholder:text-gray-300"
            onChange={(e) => {
              setForm({ ...form, name: e.target.value });
            }}
            placeholder="Label name..."
            readOnly={!isEditing}
            ref={inputRef}
            value={form.name}
          />
        </Flex>
        <Flex align="center" gap={1}>
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
            <>
              <Text className="text-[0.95rem]" color="muted">
                {format(new Date(createdAt!), "MMM d, yyyy")}
              </Text>
              <Menu>
                <Menu.Button asChild>
                  <Button
                    color="tertiary"
                    rounded="full"
                    size="sm"
                    variant="naked"
                  >
                    <MoreHorizontalIcon />
                  </Button>
                </Menu.Button>
                <Menu.Items align="end" className="w-32">
                  <Menu.Group>
                    <Menu.Item
                      onSelect={() => {
                        setIsEditing(true);
                      }}
                    >
                      <EditIcon />
                      Edit...
                    </Menu.Item>
                  </Menu.Group>
                  <Menu.Separator />
                  <Menu.Group>
                    <Menu.Item
                      onSelect={() => {
                        setIsDeleteOpen(true);
                      }}
                    >
                      <DeleteIcon />
                      Delete
                    </Menu.Item>
                  </Menu.Group>
                </Menu.Items>
              </Menu>
            </>
          )}
        </Flex>
      </form>
      <ConfirmDialog
        confirmText="Yes, go ahead"
        description="Are you sure you want to delete this label? This action cannot be undone."
        isOpen={isDeleteOpen}
        onClose={() => {
          setIsDeleteOpen(false);
        }}
        onConfirm={handleDeleteLabel}
        title="Delete Label"
      />
    </Box>
  );
};
