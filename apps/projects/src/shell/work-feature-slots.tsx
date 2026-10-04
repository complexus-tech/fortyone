"use client";

import type { ReactNode } from "react";
import { CustomFieldsCreationSlot } from "@/modules/custom-fields/public/creation";
import {
  CustomFieldDisplayPicker,
  CustomFieldsBoardProvider,
  StoryCustomFieldBadges,
} from "@/modules/custom-fields/public/display";
import { WorkflowCountsProvider } from "@/modules/stories/public/workflow-counts";
import { CreationTemplatePicker } from "@/modules/work-presets/public/template-picker";
import { BoardPropertySlotsProvider } from "@/shared/story/board-property-slots";
import type { BoardPropertySlots } from "@/shared/story/board-property-slots";
import { CreationPropertySlotsProvider } from "@/shared/story/creation-property-slots";
import type { CreationPropertySlots } from "@/shared/story/creation-property-slots";
import { MayaSkillPicker } from "@/modules/maya-skills/public";
import { MayaSkillSlotProvider } from "@/shared/maya/skill-slot";
import { ViewFavoritesProvider } from "@/shared/views/favorites-slot";
import { SaveViewAction } from "@/modules/work-presets/public/views";
import { SaveViewProvider } from "@/shared/views/save-slot";
import { FavoritesSidebar } from "./favorites-sidebar";

const BOARD_SLOTS: BoardPropertySlots = {
  Provider: CustomFieldsBoardProvider,
  Badges: StoryCustomFieldBadges,
  DisplayPicker: CustomFieldDisplayPicker,
  WorkflowCounts: WorkflowCountsProvider,
};

const CREATION_SLOTS: CreationPropertySlots = {
  Properties: CustomFieldsCreationSlot,
  TemplatePicker: CreationTemplatePicker,
};

/** Application composition keeps shared controls independent of resource owners. */
export const WorkFeatureSlots = ({ children }: { children: ReactNode }) => (
  <BoardPropertySlotsProvider slots={BOARD_SLOTS}>
    <CreationPropertySlotsProvider slots={CREATION_SLOTS}>
      <MayaSkillSlotProvider Picker={MayaSkillPicker}>
        <ViewFavoritesProvider Favorites={FavoritesSidebar}>
          <SaveViewProvider SaveView={SaveViewAction}>
            {children}
          </SaveViewProvider>
        </ViewFavoritesProvider>
      </MayaSkillSlotProvider>
    </CreationPropertySlotsProvider>
  </BoardPropertySlotsProvider>
);
