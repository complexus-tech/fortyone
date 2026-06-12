"use client";

import { Box, Text, Button } from "ui";
import { useState } from "react";
import { generateRandomColor } from "lib";
import { TagsIcon } from "icons";
import { useLabels } from "@/lib/hooks/labels";
import type { Label } from "@/types";
import { SectionHeader } from "../../components";
import { WorkspaceLabel } from "./components/label";

export const WorkspaceLabelsSettings = () => {
  const { data: labels = [] } = useLabels();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [newLabel, setNewLabel] = useState<Partial<Label> | null>(null);

  const handleCreateLabel = () => {
    setNewLabel({
      id: "new",
      name: "",
      color: generateRandomColor({
        exclude: labels.map((label) => label.color),
      }),
      createdAt: new Date().toISOString(),
    });
    setIsCreateOpen(true);
  };

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-medium">
        Labels
      </Text>

      <Box className="border-border bg-surface rounded-2xl border">
        <SectionHeader
          action={
            <Button className="shrink-0" onClick={handleCreateLabel}>
              Create Label
            </Button>
          }
          description="Create and manage labels to categorize stories."
          title="Workspace level labels"
        />

        <Box>
          {labels.length === 0 && !isCreateOpen ? (
            <Box className="px-6 py-8 text-center">
              <TagsIcon className="mx-auto mb-3 h-9" />
              <Text className="font-medium">No labels created yet</Text>
              <Text className="mt-1" color="muted">
                Create labels to help organize and categorize your stories
              </Text>
            </Box>
          ) : (
            <Box className="divide-border divide-y">
              {isCreateOpen && newLabel ? (
                <WorkspaceLabel
                  {...newLabel}
                  setIsCreateOpen={setIsCreateOpen}
                />
              ) : null}
              {labels.map((label) => (
                <WorkspaceLabel key={label.id} {...label} />
              ))}
            </Box>
          )}
        </Box>
      </Box>
    </Box>
  );
};
