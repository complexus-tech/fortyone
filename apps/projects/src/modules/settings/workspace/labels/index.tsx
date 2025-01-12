"use client";

import { Box, Flex, Text, Button, Input, ColorPicker, TimeAgo } from "ui";
import { useState } from "react";
import { SearchIcon } from "icons";
import type { Label } from "@/types";
import { useLabels } from "@/lib/hooks/labels";
import { SectionHeader } from "../../components";

export const WorkspaceLabelsSettings = () => {
  const { data: labels = [] } = useLabels();
  const [searchQuery, setSearchQuery] = useState("");
  const [editingLabel, setEditingLabel] = useState<Label | null>(null);
  const [editedName, setEditedName] = useState("");
  const [editedColor, setEditedColor] = useState("");

  const handleStartEdit = (label: Label) => {
    setEditingLabel(label);
    setEditedName(label.name);
    setEditedColor(label.color);
  };

  const handleSaveEdit = () => {
    if (!editingLabel) return;
    // Handle save edit
    setEditingLabel(null);
  };

  const handleCancelEdit = () => {
    setEditingLabel(null);
  };

  const filteredLabels = labels.filter((label) =>
    label.name.toLowerCase().includes(searchQuery.toLowerCase()),
  );

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Workspace level labels
      </Text>

      <Flex className="mb-6" justify="between">
        <Box className="relative w-[400px]">
          <SearchIcon className="text-gray-500 absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2" />
          <Input
            className="pl-9"
            onChange={(e) => {
              setSearchQuery(e.target.value);
            }}
            placeholder="Filter by name..."
            type="search"
            value={searchQuery}
          />
        </Box>
        <Button>New label</Button>
      </Flex>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Labels help you categorize and organize stories."
          title="Story labels"
        />

        <Box className="border-b border-gray-100 px-6 py-4 dark:border-dark-100">
          <Flex>
            <Text className="w-[400px] font-medium">Name</Text>
            <Text className="w-[200px] font-medium">Usage</Text>
            <Text className="w-[200px] font-medium">Created</Text>
          </Flex>
        </Box>

        <Box className="divide-y divide-gray-100 dark:divide-dark-100">
          {filteredLabels.map((label) => (
            <Box
              className="px-6 py-4 transition-colors hover:bg-gray-50 dark:hover:bg-dark-300"
              key={label.id}
            >
              <Flex>
                <Box className="w-[400px]">
                  {editingLabel?.id === label.id ? (
                    <Flex align="center" gap={3}>
                      <ColorPicker
                        onChange={(color) => {
                          setEditedColor(color);
                        }}
                        value={editedColor}
                      />
                      <Input
                        className="w-[300px]"
                        onChange={(e) => {
                          setEditedName(e.target.value);
                        }}
                        value={editedName}
                      />
                      <Button
                        className="ml-2"
                        onClick={() => {
                          handleSaveEdit();
                        }}
                        size="sm"
                      >
                        Save
                      </Button>
                      <Button
                        color="tertiary"
                        onClick={() => {
                          handleCancelEdit();
                        }}
                        size="sm"
                        variant="outline"
                      >
                        Cancel
                      </Button>
                    </Flex>
                  ) : (
                    <Flex
                      align="center"
                      className="cursor-pointer"
                      gap={3}
                      onClick={() => {
                        handleStartEdit(label);
                      }}
                    >
                      <Box
                        className="size-3 rounded-full"
                        style={{ backgroundColor: label.color }}
                      />
                      <Text className="font-medium">{label.name}</Text>
                    </Flex>
                  )}
                </Box>
                <Text className="text-gray-500 w-[200px]">{0} issues</Text>
                <Text className="text-gray-500 w-[200px]">
                  <TimeAgo timestamp={label.createdAt} />
                </Text>
              </Flex>
            </Box>
          ))}
        </Box>
      </Box>
    </Box>
  );
};
