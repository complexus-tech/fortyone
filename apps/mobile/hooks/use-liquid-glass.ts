import { useEffect, useState } from "react";
import { AccessibilityInfo, Platform } from "react-native";

const SUPPORTS_GLASS =
  Platform.OS === "ios" && Number.parseInt(String(Platform.Version), 10) >= 26;

export function useLiquidGlass() {
  const [reduceTransparency, setReduceTransparency] = useState(true);

  useEffect(() => {
    if (!SUPPORTS_GLASS) return;
    let active = true;
    let changed = false;
    const subscription = AccessibilityInfo.addEventListener(
      "reduceTransparencyChanged",
      (enabled) => {
        changed = true;
        setReduceTransparency(enabled);
      },
    );
    void AccessibilityInfo.isReduceTransparencyEnabled().then(
      (enabled) => {
        if (active && !changed) setReduceTransparency(enabled);
      },
      () => {
        // Keep the opaque fallback if the accessibility preference is unavailable.
      },
    );
    return () => {
      active = false;
      subscription.remove();
    };
  }, []);

  return SUPPORTS_GLASS && !reduceTransparency;
}
