import { Box, Flex, Skeleton } from "ui";
import { BodyContainer } from "@/components/shared/body";
import { RowWrapper } from "@/components/ui/row-wrapper";
import { ObjectivesHeader } from "./header";

export const ObjectivesSkeleton = () => {
  return (
    <>
      <ObjectivesHeader />
      <BodyContainer>
        {Array.from({ length: 8 }).map((_, index) => (
          <RowWrapper
            className="justify-between px-5 py-3 md:px-12"
            key={index}
          >
            <Box className="flex min-w-0 flex-1 items-center gap-2">
              <Skeleton className="h-5 w-12 shrink-0" />
              <Skeleton className="h-6 w-56" />
            </Box>
            <Flex align="center" gap={2}>
              <Box className="shrink-0">
                <Skeleton className="h-7 w-20 rounded-lg" />
              </Box>
              <Box className="shrink-0">
                <Skeleton className="h-7 w-20 rounded-lg" />
              </Box>
              <Box className="shrink-0">
                <Skeleton className="h-7 w-20 rounded-lg" />
              </Box>
              <Box className="hidden shrink-0 items-center gap-1.5 px-1 sm:flex">
                <Skeleton className="h-4 w-4 rounded-full" />
                <Skeleton className="h-5 w-10" />
              </Box>
              <Box className="hidden shrink-0 md:block">
                <Skeleton className="h-7 w-20 rounded-lg" />
              </Box>
              <Box className="hidden shrink-0 md:block">
                <Skeleton className="h-7 w-7 rounded-full" />
              </Box>
            </Flex>
          </RowWrapper>
        ))}
      </BodyContainer>
    </>
  );
};
