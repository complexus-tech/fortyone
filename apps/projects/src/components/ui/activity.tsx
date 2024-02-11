import { Box, Flex, Text, Avatar } from "ui";

export type ActivityProps = {
  id: number;
  user: string;
  action: string;
  prevValue: string;
  newValue: string;
  timestamp: string;
};

export const Activity = ({
  user,
  action,
  prevValue,
  newValue,
  timestamp,
}: ActivityProps) => (
  <Flex align="center" className="z-[1]" gap={1}>
    <Box className="flex aspect-square items-center rounded-full bg-white p-[0.3rem] dark:bg-dark">
      <Avatar
        name="Joseph Mukorivo"
        size="sm"
        src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
      />
    </Box>
    <Text className="ml-2" fontWeight="medium">
      {user}
    </Text>
    <Text color="muted">{action}</Text>
    <Text fontWeight="medium">{prevValue}</Text>
    <Text color="muted">to</Text>
    <Text fontWeight="medium">{newValue}</Text>
    <Text color="muted">·</Text>
    <Text color="muted">{timestamp}</Text>
  </Flex>
);
