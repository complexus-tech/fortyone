"use client";

import { Box, Text, Button } from "ui";
import { useEffect, useState } from "react";
import { generateRandomColor } from "lib";
import { TagsIcon } from "icons";
import { useLabels } from "@/lib/hooks/labels";
import type { Label } from "@/types";
import { SectionHeader } from "../../components";
import { WorkspaceLabel } from "./components/label";

export const WorkspaceLabelsSettings = () => {
  const { data: labels = [] } = useLabels();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const newLabelTemplate: Partial<Label> = {
    id: "new",
    name: "",
    color: generateRandomColor({ exclude: labels.map((l) => l.color) }),
    createdAt: new Date().toISOString(),
  };
  const [newLabel, setNewLabel] = useState(newLabelTemplate);

  useEffect(() => {
    if (isCreateOpen) {
      setNewLabel((prev) => ({
        ...prev,
        color: generateRandomColor({ exclude: labels.map((l) => l.color) }),
      }));
    }
  }, [isCreateOpen, labels]);

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-medium">
        Labels
      </Text>

      <Box className="rounded-2xl border border-border bg-surface">
        <SectionHeader
          action={
            <Button
              className="shrink-0"
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
          {labels.length === 0 && !isCreateOpen ? (
            <Box className="px-6 py-8 text-center">
              <TagsIcon className="mx-auto mb-3 h-9" />
              <Text className="font-medium">No labels created yet</Text>
              <Text className="mt-1" color="muted">
                Create labels to help organize and categorize your stories
              </Text>
            </Box>
          ) : (
            <Box className="divide-y divide-gray-100 dark:divide-dark-100">
              {isCreateOpen ? (
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
