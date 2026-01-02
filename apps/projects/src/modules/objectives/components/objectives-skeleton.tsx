import { Box, Flex, Skeleton } from "ui";
import { BodyContainer } from "@/components/shared/body";
import { RowWrapper } from "@/components/ui/row-wrapper";
import { TableHeader } from "./heading";
import { ObjectivesHeader } from "./header";

export const ObjectivesSkeleton = ({ isInTeam }: { isInTeam?: boolean }) => {
  return (
    <>
      <ObjectivesHeader />
      <BodyContainer className="h-[calc(100dvh-3.7rem)]">
        <TableHeader isInTeam={isInTeam} />
        {Array.from({ length: 8 }).map((_, index) => (
          <RowWrapper
            className="justify-between px-5 py-3 md:px-12"
            key={index}
          >
            <Box className="flex shrink-0 items-center gap-2 md:w-[300px]">
              <Flex
                align="center"
                className="size-8 shrink-0 rounded-[0.6rem] bg-surface-muted"
                justify="center"
              >
                <Skeleton className="h-4 w-4" />
              </Flex>
              <Skeleton className="h-6 w-56" />
            </Box>
            <Flex align="center" gap={4}>
              {!isInTeam && (
                <Box className="hidden w-[45px] shrink-0 items-center gap-1.5 md:flex">
                  <Skeleton className="h-3 w-3 rounded-full" />
                  <Skeleton className="h-5 w-8" />
                </Box>
              )}
              <Box className="hidden w-[40px] shrink-0 items-center md:flex">
                <Skeleton className="h-6 w-6 rounded-full" />
              </Box>
              <Box className="hidden w-[60px] shrink-0 items-center gap-1.5 pl-0.5 md:flex">
                <Skeleton className="h-4 w-4 rounded-full" />
                <Skeleton className="h-5 w-10" />
              </Box>
              <Box className="hidden w-[120px] shrink-0 md:block">
                <Skeleton className="h-5 w-24" />
              </Box>
              <Box className="hidden w-[100px] shrink-0 md:block">
                <Skeleton className="h-5 w-20" />
              </Box>
              <Box className="hidden w-[100px] shrink-0 md:block">
                <Skeleton className="h-5 w-24" />
              </Box>
              <Box className="shrink-0 md:w-[120px]">
                <Skeleton className="h-5 w-20" />
              </Box>
            </Flex>
          </RowWrapper>
        ))}
      </BodyContainer>
    </>
  );
};
