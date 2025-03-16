import { useWorkspaceSettings } from "@/lib/hooks/workspace/settings";

export const useFeatures = () => {
  const {
    data: settings = {
      sprintEnabled: true,
      objectiveEnabled: true,
      keyResultEnabled: true,
    },
  } = useWorkspaceSettings();

  return {
    sprintEnabled: settings.sprintEnabled,
    objectiveEnabled: settings.objectiveEnabled,
    keyResultEnabled: settings.keyResultEnabled,
  };
};
