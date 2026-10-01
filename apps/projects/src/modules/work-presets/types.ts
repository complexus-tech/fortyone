import type { TaskTemplateConfiguration } from "@/shared/story/task-template";
import type { SavedViewConfiguration } from "@/shared/story/view-configuration";

export type { SavedViewConfiguration } from "@/shared/story/view-configuration";

export type { TaskTemplateConfiguration } from "@/shared/story/task-template";

export type PresetConfiguration =
  | SavedViewConfiguration
  | TaskTemplateConfiguration;
export type PresetKind = "view" | "template";
export type Preset<T extends PresetConfiguration = PresetConfiguration> = {
  id: string;
  teamId: string;
  ownerId: string;
  kind: PresetKind;
  visibility: "personal" | "team";
  name: string;
  configuration: T;
  createdAt: string;
  updatedAt: string;
  canEdit: boolean;
};
export type PresetInput = Pick<
  Preset,
  "teamId" | "kind" | "visibility" | "name" | "configuration"
>;
export type PresetPage = { items: Preset[]; nextCursor: string };
