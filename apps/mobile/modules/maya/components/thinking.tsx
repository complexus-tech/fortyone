import { useEffect, useState } from "react";
import {
  AccessibilityInfo,
  Animated,
  StyleSheet,
  useAnimatedValue,
} from "react-native";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks/theme";

export function MayaThinking() {
  const { resolvedTheme } = useTheme();
  const [reduceMotion, setReduceMotion] = useState(true);
  const opacity = useAnimatedValue(1);

  useEffect(() => {
    let mounted = true;
    let preferenceChanged = false;
    const subscription = AccessibilityInfo.addEventListener(
      "reduceMotionChanged",
      (enabled) => {
        preferenceChanged = true;
        if (mounted) setReduceMotion(enabled);
      },
    );
    void AccessibilityInfo.isReduceMotionEnabled()
      .then((enabled) => {
        if (mounted && !preferenceChanged) setReduceMotion(enabled);
      })
      .catch(() => {
        // Keep the label static if the platform cannot read this preference.
        if (mounted && !preferenceChanged) setReduceMotion(true);
      });

    return () => {
      mounted = false;
      subscription.remove();
    };
  }, []);

  useEffect(() => {
    opacity.setValue(1);
    if (reduceMotion) return;

    const animation = Animated.loop(
      Animated.sequence([
        Animated.timing(opacity, {
          toValue: 0.9,
          duration: 900,
          useNativeDriver: true,
          isInteraction: false,
        }),
        Animated.timing(opacity, {
          toValue: 1,
          duration: 900,
          useNativeDriver: true,
          isInteraction: false,
        }),
      ]),
    );
    animation.start();

    return () => animation.stop();
  }, [opacity, reduceMotion]);

  return (
    <Animated.Text
      accessibilityLiveRegion="polite"
      style={[
        styles.text,
        { color: themeColors[resolvedTheme].textMuted, opacity },
      ]}
    >
      Maya is thinking…
    </Animated.Text>
  );
}

const styles = StyleSheet.create({
  text: {
    fontSize: 16,
    lineHeight: 25,
  },
});
