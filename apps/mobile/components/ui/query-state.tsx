import { ActivityIndicator, View } from "react-native";
import { Button } from "./button";
import { Text } from "./text";

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
    className="flex-1 items-center justify-center gap-4 px-6 py-8"
    accessibilityLiveRegion="polite"
  >
    {loading ? <ActivityIndicator /> : null}
    <Text fontWeight="semibold" className="text-center">
      {title}
    </Text>
    {message ? (
      <Text color="muted" className="text-center">
        {message}
      </Text>
    ) : null}
    {onRetry ? <Button onPress={onRetry}>{retryLabel}</Button> : null}
  </View>
);
