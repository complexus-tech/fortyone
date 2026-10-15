import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { DesktopPreferences } from "./bridge";
import { getDesktopBridge } from "./bridge";
import { useDesktop } from "./desktop-provider";

const PREFERENCES_KEY = ["desktop", "preferences"] as const;

export const useDesktopPreferences = () => {
  const { isDesktop } = useDesktop();
  const bridge = isDesktop ? getDesktopBridge() : null;
  const supported = Boolean(bridge?.getPreferences && bridge.updatePreferences);
  const client = useQueryClient();
  const preferences = useQuery({
    queryKey: PREFERENCES_KEY,
    queryFn: () => {
      if (!bridge?.getPreferences)
        throw new Error("Update FortyOne to manage desktop preferences.");
      return bridge.getPreferences();
    },
    enabled: supported,
    staleTime: 30_000,
  });
  const update = useMutation({
    mutationFn: async (next: Partial<DesktopPreferences>) => {
      if (!bridge?.updatePreferences)
        throw new Error("Update FortyOne to manage desktop preferences.");
      if (next.notifications && bridge.requestNotificationPermission) {
        const permission = await bridge.requestNotificationPermission();
        if (permission !== "granted") {
          throw new Error(
            "Allow FortyOne notifications in your system settings, then try again.",
          );
        }
      }
      return bridge.updatePreferences(next);
    },
    onSuccess: (saved) => client.setQueryData(PREFERENCES_KEY, saved),
  });
  return { supported, preferences, update };
};
