import type { ReactNode } from "react";
import { StyleSheet, View } from "react-native";
import { Text } from "./Text";

export type ScreenHeaderProps = {
  title: string;
  subtitle?: string;
  leading?: ReactNode;
  trailing?: ReactNode;
  compact?: boolean;
};

/** Owns its page gutters; place inside a full-width screen container. */
export const ScreenHeader = ({
  title,
  subtitle,
  leading,
  trailing,
  compact = false,
}: ScreenHeaderProps) => (
  <View style={styles.container}>
    <View style={styles.toolbar}>
      {leading ? <View style={styles.actions}>{leading}</View> : null}
      <View style={styles.heading}>
        <Text
          accessibilityRole="header"
          fontSize={compact ? "xl" : "3xl"}
          fontWeight="bold"
          style={compact ? undefined : styles.pageTitle}
          numberOfLines={2}
        >
          {title}
        </Text>
        {subtitle ? (
          <Text color="muted" fontSize="sm" style={styles.subtitle}>
            {subtitle}
          </Text>
        ) : null}
      </View>
      {trailing ? <View style={styles.actions}>{trailing}</View> : null}
    </View>
  </View>
);

const styles = StyleSheet.create({
  pageTitle: { fontSize: 36, lineHeight: 43 },
  container: { paddingHorizontal: 20, paddingBottom: 12 },
  toolbar: {
    minHeight: 56,
    flexDirection: "row",
    alignItems: "center",
    gap: 8,
  },
  heading: { flex: 1, minWidth: 0 },
  subtitle: { marginTop: 4 },
  actions: {
    flexDirection: "row",
    alignItems: "center",
    gap: 4,
    flexShrink: 0,
  },
});
