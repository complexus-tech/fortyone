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
    <Box className="flex aspect-square items-center rounded-full bg-white p-[0.3rem] dark:bg-dark-300">
      <Avatar
        name="Joseph Mukorivo"
        size="xs"
        src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
      />
    </Box>
    <Text className="ml-1" fontWeight="medium">
      {user}
    </Text>
    <Text className="text-[0.95rem]" color="muted">
      {action}
    </Text>
    <Text className="text-[0.95rem]" fontWeight="medium">
      {prevValue}
    </Text>
    <Text className="text-[0.95rem]" color="muted">
      to
    </Text>
    <Text className="text-[0.95rem]" fontWeight="medium">
      {newValue}
    </Text>
    <Text className="text-[0.95rem]" color="muted">
      ·
    </Text>
    <Text className="text-[0.95rem]" color="muted">
      {timestamp}
    </Text>
  </Flex>
);
