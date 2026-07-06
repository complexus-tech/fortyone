import { Box, Flex } from "ui";

const ROW_WIDTHS = ["w-32", "w-40", "w-28", "w-36"];

export const MenuLoadingSkeleton = ({
  avatar = false,
  rows = 4,
  trailing = true,
}: {
  avatar?: boolean;
  rows?: number;
  trailing?: boolean;
}) => (
  <Box className="space-y-1.5">
    {Array.from({ length: rows }).map((_, index) => (
      <Flex
        align="center"
        className="h-9 animate-pulse justify-between rounded-md px-2"
        gap={3}
        key={index}
      >
        <Flex align="center" className="min-w-0 flex-1" gap={2}>
          <Box
            className={`bg-skeleton shrink-0 ${
              avatar ? "size-6 rounded-full" : "size-4 rounded"
            }`}
          />
          <Box
            className={`bg-skeleton h-3 rounded ${ROW_WIDTHS[index % ROW_WIDTHS.length]}`}
          />
        </Flex>
        {trailing ? (
          <Box className="bg-skeleton h-3 w-5 shrink-0 rounded" />
        ) : null}
      </Flex>
    ))}
  </Box>
);
