"use client";

import { Box, Flex, Text, Button, Menu } from "ui";
import { DeleteIcon, EditIcon, MoreHorizontalIcon } from "icons";
import { useState } from "react";
import type { Label } from "@/types";
import { useLabels } from "@/lib/hooks/labels";
import { SectionHeader } from "../../components";
import { LabelDialog } from "./components/label-dialog";
import { DeleteDialog } from "./components/delete-dialog";

type LabelFormData = {
  name: string;
  color: string;
};

export const WorkspaceLabelsSettings = () => {
  const { data: labels = [] } = useLabels();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [isEditOpen, setIsEditOpen] = useState(false);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [selectedLabel, setSelectedLabel] = useState<Label | null>(null);

  const handleCreateLabel = (_data: LabelFormData) => {
    // Handle create label
    setIsCreateOpen(false);
  };

  const handleEditLabel = (_data: LabelFormData) => {
    // Handle edit label
    setIsEditOpen(false);
  };

  const handleDeleteLabel = () => {
    // Handle delete label
    setIsDeleteOpen(false);
  };

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Labels
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          action={
            <Button
              onClick={() => {
                setIsCreateOpen(true);
              }}
            >
              Create Label
            </Button>
          }
          description="Create and manage labels to categorize stories."
          title="Workspace level labels"
        />

        <Box>
          {labels.length === 0 ? (
            <Box className="px-6 py-8 text-center">
              <Text className="font-medium">No labels created yet</Text>
              <Text className="mt-1" color="muted">
                Create labels to help organize and categorize your stories
              </Text>
              <Button
                className="mt-4"
                onClick={() => {
                  setIsCreateOpen(true);
                }}
              >
                Create your first label
              </Button>
            </Box>
          ) : (
            <Box className="divide-y divide-gray-100 dark:divide-dark-100">
              {labels.map((label) => (
                <Box
                  className="px-6 py-2.5 transition-colors hover:bg-gray-50 dark:hover:bg-dark-300"
                  key={label.id}
                >
                  <Flex align="center" justify="between">
                    <Flex align="center" gap={3}>
                      <Box
                        className="size-4 rounded"
                        style={{ backgroundColor: label.color }}
                      />
                      <Text className="font-medium">{label.name}</Text>
                    </Flex>
                    <Flex align="center" gap={3}>
                      <Menu>
                        <Menu.Button>
                          <Button
                            aria-label="More options"
                            color="tertiary"
                            variant="naked"
                          >
                            <MoreHorizontalIcon className="h-4 w-4" />
                          </Button>
                        </Menu.Button>
                        <Menu.Items>
                          <Menu.Group>
                            <Menu.Item
                              onClick={() => {
                                setSelectedLabel(label);
                                setIsEditOpen(true);
                              }}
                            >
                              <EditIcon /> Edit label
                            </Menu.Item>
                          </Menu.Group>
                          <Menu.Separator />
                          <Menu.Group>
                            <Menu.Item
                              className="text-danger"
                              onClick={() => {
                                setSelectedLabel(label);
                                setIsDeleteOpen(true);
                              }}
                            >
                              <DeleteIcon className="text-danger dark:text-danger" />{" "}
                              Delete label
                            </Menu.Item>
                          </Menu.Group>
                        </Menu.Items>
                      </Menu>
                    </Flex>
                  </Flex>
                </Box>
              ))}
            </Box>
          )}
        </Box>
      </Box>

      <LabelDialog
        isOpen={isCreateOpen}
        mode="create"
        onClose={() => {
          setIsCreateOpen(false);
        }}
        onSubmit={handleCreateLabel}
      />

      <LabelDialog
        isOpen={isEditOpen}
        mode="edit"
        onClose={() => {
          setIsEditOpen(false);
          setSelectedLabel(null);
        }}
        onSubmit={handleEditLabel}
        selectedLabel={selectedLabel}
      />

      <DeleteDialog
        isOpen={isDeleteOpen}
        onClose={() => {
          setIsDeleteOpen(false);
          setSelectedLabel(null);
        }}
        onConfirm={handleDeleteLabel}
      />
    </Box>
  );
};
