import type { ReactNode } from "react";
import { Box, Flex, Text } from "ui";

export const MetricCard = ({
  detail,
  icon,
  label,
  value,
}: {
  detail?: string;
  icon?: ReactNode;
  label: string;
  value: string;
}) => {
  return (
    <Box className="border-border bg-surface shadow-shadow rounded-xl border-[0.5px] p-4 shadow-sm">
      <Flex align="start" justify="between">
        <Box>
          <Text className="text-[0.92rem]" color="muted">
            {label}
          </Text>
          <Text
            className="font-heading mt-2 text-[1.65rem] leading-none"
            fontWeight="semibold"
          >
            {value}
          </Text>
        </Box>
        {icon ? (
          <Box className="bg-surface-elevated border-border flex size-9 items-center justify-center rounded-xl border-[0.5px]">
            {icon}
          </Box>
        ) : null}
      </Flex>
      {detail ? (
        <Text className="mt-3 text-[0.92rem]" color="muted">
          {detail}
        </Text>
      ) : null}
    </Box>
  );
};
