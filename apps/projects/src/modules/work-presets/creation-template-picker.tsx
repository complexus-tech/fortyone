"use client";

import type { CreationTemplatePickerProps } from "@/shared/story/creation-property-slots";
import type { TaskTemplateConfiguration } from "./types";
import { PresetPicker } from "./preset-picker";

export const CreationTemplatePicker = ({
  teamId,
  disabled,
  onSelect,
}: CreationTemplatePickerProps) => (
  <PresetPicker
    disabled={disabled}
    hideWhenEmpty
    kind="template"
    label="Use template"
    onSelect={(preset) => {
      onSelect(preset.configuration as TaskTemplateConfiguration);
    }}
    teamId={teamId}
  />
);
