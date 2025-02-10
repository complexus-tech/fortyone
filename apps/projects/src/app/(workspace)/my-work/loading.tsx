import { Box, Flex, Skeleton } from "ui";
import { HeaderContainer } from "@/components/shared";
import { MainTabs } from "./tabs";

const HeaderSkeleton = () => {
  return (
    <HeaderContainer className="justify-between">
      <Flex align="center" gap={2}>
        <Skeleton className="h-8 w-32" />
        <Skeleton className="h-6 w-24 rounded-full" />
      </Flex>
      <Flex align="center" gap={2}>
        <Skeleton className="h-8 w-32" />
        <span className="text-gray-200 dark:text-dark-100">|</span>
        <Skeleton className="h-8 w-24" />
      </Flex>
    </HeaderContainer>
  );
};

export default function Loading() {
  return (
    <Box className="h-[calc(100vh-4rem)]">
      <HeaderSkeleton />
      <Box className="sticky top-0 z-10 flex h-[3.7rem] w-full flex-col justify-center border-b-[0.5px] border-gray-100/60 bg-white dark:border-dark-100 dark:bg-dark-300">
        <MainTabs />
      </Box>
      <Box className="mt-4 space-y-3 px-4">
        {Array.from({ length: 8 }).map((_, i) => (
          <Flex align="center" className="w-full" gap={4} key={i}>
            <Skeleton className="h-12 flex-1" />
            <Flex align="center" gap={3}>
              <Skeleton className="h-8 w-24" />
              <Skeleton className="h-8 w-24" />
              <Skeleton className="h-8 w-8 rounded-full" />
            </Flex>
          </Flex>
        ))}
      </Box>
    </Box>
  );
}
