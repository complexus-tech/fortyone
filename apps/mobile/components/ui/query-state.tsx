import { ActivityIndicator, StyleSheet, View } from "react-native";
import { colors } from "@/constants/colors";
import { Button } from "./Button";
import { Text } from "./Text";

type QueryStateProps = {
  title: string;
  message?: string;
  loading?: boolean;
  onRetry?: () => void;
  retryLabel?: string;
};

export const QueryState = ({
  title,
  message,
  loading = false,
  onRetry,
  retryLabel = "Try again",
}: QueryStateProps) => (
  <View
    style={styles.container}
    accessibilityLiveRegion="polite"
    accessibilityState={{ busy: loading }}
  >
    {loading ? <ActivityIndicator color={colors.primary} /> : null}
    <Text fontSize="lg" fontWeight="semibold" align="center">
      {title}
    </Text>
    {message ? (
      <Text color="muted" align="center" style={styles.message}>
        {message}
      </Text>
    ) : null}
    {onRetry ? (
      <Button
        color="tertiary"
        fullWidth={false}
        disabled={loading}
        style={styles.retry}
        onPress={onRetry}
      >
        {retryLabel}
      </Button>
    ) : null}
  </View>
);

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    gap: 12,
    paddingHorizontal: 24,
    paddingVertical: 40,
  },
  message: { maxWidth: 320 },
  retry: { marginTop: 8, minWidth: 120 },
});
