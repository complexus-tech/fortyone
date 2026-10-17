import { useWorkspaceSettings } from "@/lib/hooks/workspace/settings";

export const useFeatures = () => {
  const {
    data: settings = {
      objectiveEnabled: true,
      keyResultEnabled: true,
      chatEnabled: true,
    },
    isPending,
    isError,
  } = useWorkspaceSettings();

  return {
    isPending,
    isError,
    objectiveEnabled: settings.objectiveEnabled,
    keyResultEnabled: settings.keyResultEnabled,
    chatEnabled: settings.chatEnabled,
  };
};
