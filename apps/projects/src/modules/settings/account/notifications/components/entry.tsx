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
    <Flex align="center" justify="between">
      <Box>
        <Text className="font-medium">{title}</Text>
        <Text color="muted">{description}</Text>
      </Box>
      <Switch
        checked={checked}
        onCheckedChange={(value) => onChange?.(value)}
      />
    </Flex>
  );
};
