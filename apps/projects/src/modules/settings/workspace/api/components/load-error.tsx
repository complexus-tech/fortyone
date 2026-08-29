import { Button, Flex, Text } from "ui";

export const LoadError = ({
  label,
  onRetry,
}: {
  label: string;
  onRetry: () => void;
}) => (
  <Flex align="center" className="gap-3 px-6 py-5" justify="between" wrap>
    <Text color="muted">Could not load {label}.</Text>
    <Button color="tertiary" onClick={onRetry} size="sm" variant="outline">
      Try again
    </Button>
  </Flex>
);
