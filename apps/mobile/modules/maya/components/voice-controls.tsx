import { useEffect } from "react";
import {
  Animated,
  Pressable,
  StyleSheet,
  useAnimatedValue,
  View,
} from "react-native";
import { useReducedMotion } from "react-native-reanimated";
import { Ionicons } from "@expo/vector-icons";
import { Text } from "@/components/ui";
import { WebIcon } from "@/components/icons/web-icon";
import { themeColors, colors } from "@/constants/colors";
import { useTheme } from "@/hooks";
import { ComposerSurface } from "./composer-surface";

export function VoiceControls({
  status,
  muted,
  speaking,
  remainingSeconds,
  onMute,
  onStop,
}: {
  status: string;
  muted: boolean;
  speaking: boolean;
  remainingSeconds: number | null;
  onMute: () => void;
  onStop: () => void;
}) {
  const { resolvedTheme } = useTheme();
  const theme = themeColors[resolvedTheme];
  const reduceMotion = useReducedMotion();
  const opacity = useAnimatedValue(1);
  useEffect(() => {
    if (!speaking || reduceMotion) {
      opacity.setValue(1);
      return;
    }
    const motion = Animated.loop(
      Animated.sequence([
        Animated.timing(opacity, {
          toValue: 0.4,
          duration: 450,
          useNativeDriver: true,
          isInteraction: false,
        }),
        Animated.timing(opacity, {
          toValue: 1,
          duration: 450,
          useNativeDriver: true,
          isInteraction: false,
        }),
      ]),
    );
    motion.start();
    return () => motion.stop();
  }, [speaking, reduceMotion, opacity]);
  const connecting = status === "connecting";
  const remaining =
    remainingSeconds === null
      ? null
      : Math.max(0, Math.floor(remainingSeconds));
  return (
    <ComposerSurface>
      <View style={styles.panel}>
        <Animated.View
          accessible={false}
          accessibilityElementsHidden
          importantForAccessibility="no-hide-descendants"
          style={[styles.wave, { opacity }]}
        >
          <WebIcon
            name="voice"
            size={22}
            color={connecting || muted ? theme.textMuted : theme.foreground}
          />
        </Animated.View>
        <View style={styles.status}>
          <Text
            numberOfLines={1}
            ellipsizeMode="tail"
            accessibilityLiveRegion="polite"
            style={styles.statusText}
          >
            {connecting
              ? "Connecting…"
              : muted
                ? "Microphone muted"
                : speaking
                  ? "Maya is speaking"
                  : "Listening"}
          </Text>
          {!connecting && remaining !== null ? (
            <Text
              color="muted"
              numberOfLines={1}
              accessibilityLabel={`${Math.floor(remaining / 60)} minutes ${remaining % 60} seconds remaining`}
              style={styles.time}
            >
              {`${Math.floor(remaining / 60)}:${String(remaining % 60).padStart(2, "0")}`}
            </Text>
          ) : null}
        </View>
        <Pressable
          accessibilityRole="button"
          accessibilityLabel={muted ? "Unmute microphone" : "Mute microphone"}
          onPress={onMute}
          disabled={connecting}
          accessibilityState={{ disabled: connecting, selected: muted }}
          style={({ pressed }) => [
            styles.control,
            { opacity: connecting ? 0.35 : pressed ? 0.55 : 1 },
          ]}
        >
          <Ionicons
            name={muted ? "mic-off-outline" : "mic-outline"}
            size={22}
            color={theme.foreground}
          />
        </Pressable>
        <Pressable
          accessibilityRole="button"
          accessibilityLabel="End voice conversation"
          onPress={onStop}
          style={({ pressed }) => [
            styles.control,
            { opacity: pressed ? 0.65 : 1 },
          ]}
        >
          <View
            pointerEvents="none"
            style={[styles.end, { backgroundColor: colors.danger }]}
          >
            <WebIcon name="close" size={20} color={colors.dangerForeground} />
          </View>
        </Pressable>
      </View>
    </ComposerSurface>
  );
}
const styles = StyleSheet.create({
  panel: {
    minHeight: 48,
    flexDirection: "row",
    alignItems: "center",
    padding: 2,
    gap: 2,
  },
  status: {
    flex: 1,
    minWidth: 0,
    flexDirection: "row",
    alignItems: "center",
    gap: 6,
  },
  statusText: { flex: 1, fontSize: 16, lineHeight: 24 },
  time: { fontSize: 12, lineHeight: 18, fontVariant: ["tabular-nums"] },
  wave: {
    width: 44,
    height: 44,
    alignItems: "center",
    justifyContent: "center",
  },
  control: {
    width: 44,
    height: 44,
    borderRadius: 22,
    alignItems: "center",
    justifyContent: "center",
  },
  end: {
    width: 28,
    height: 28,
    borderRadius: 14,
    alignItems: "center",
    justifyContent: "center",
  },
});
