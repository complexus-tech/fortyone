import { useWorkspaceSettings } from "@/lib/hooks/workspace/settings";

export const useFeatures = () => {
  const {
    data: settings = {
      objectiveEnabled: true,
      keyResultEnabled: true,
    },
    isPending,
  } = useWorkspaceSettings();

  return {
    isPending,
    objectiveEnabled: settings.objectiveEnabled,
    keyResultEnabled: settings.keyResultEnabled,
  };
};
