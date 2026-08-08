import { Box, Skeleton } from "ui";

export default function Loading() {
  return (
    <Box className="flex h-dvh flex-col overflow-hidden">
      <Skeleton className="h-[3.6rem] w-full shrink-0 rounded-none" />
      <Skeleton className="min-h-0 w-full flex-1 rounded-none" />
    </Box>
  );
}
