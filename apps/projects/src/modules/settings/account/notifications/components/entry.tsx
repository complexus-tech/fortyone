import { Box, Flex, Text, Switch } from "ui";

type EntryProps = {
  title: string;
  description: string;
  checked?: boolean;
  onChange?: (checked: boolean) => void;
};

export const Entry = ({
  title,
  description,
  checked = false,
  onChange,
}: EntryProps) => {
  return (
    <Flex align="center" gap={2} justify="between">
      <Box>
        <Text className="font-medium">{title}</Text>
        <Text className="line-clamp-2" color="muted">
          {description}
        </Text>
      </Box>
      <Switch
        checked={checked}
        className="shrink-0"
        onCheckedChange={(value) => onChange?.(value)}
      />
    </Flex>
  );
};
