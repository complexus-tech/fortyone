import { useWorkspaceSettings } from "@/lib/hooks/workspace/settings";

export const useFeatures = () => {
  const {
    data: settings = {
      objectiveEnabled: true,
      keyResultEnabled: true,
    },
  } = useWorkspaceSettings();

  return {
    objectiveEnabled: settings.objectiveEnabled,
    keyResultEnabled: settings.keyResultEnabled,
  };
};
