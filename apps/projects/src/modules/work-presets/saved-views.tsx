"use client";

import { useState } from "react";
import { useUserRole } from "@/hooks/role";
import { usePresetMutations } from "./hooks";
import { PresetNameDialog } from "./name-dialog";
import { PresetPicker } from "./preset-picker";
import type { SavedViewConfiguration } from "./types";

export const SavedViews = ({
  teamId,
  configuration,
  onApply,
}: {
  teamId: string;
  configuration: SavedViewConfiguration;
  onApply: (configuration: SavedViewConfiguration) => void;
}) => {
  const [saving, setSaving] = useState(false);
  const { userRole } = useUserRole();
  const { create } = usePresetMutations(teamId, "view");
  return (
    <>
      <PresetPicker
        kind="view"
        label="Views"
        onSaveCurrent={
          userRole !== "guest"
            ? () => {
                setSaving(true);
              }
            : undefined
        }
        onSelect={(preset) => {
          onApply(preset.configuration as SavedViewConfiguration);
        }}
        teamId={teamId}
      />
      {saving ? (
        <PresetNameDialog
          onOpenChange={setSaving}
          onSave={(name, visibility) =>
            create.mutateAsync({
              teamId,
              kind: "view",
              name,
              visibility,
              configuration,
            })
          }
          open
          shared
          title="Save current view"
        />
      ) : null}
    </>
  );
};
