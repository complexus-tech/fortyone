"use client";

import { format } from "date-fns";
import { OKRIcon } from "icons";
import Link from "next/link";
import {
  Box,
  BreadCrumbs,
  Button,
  Flex,
  ProgressBar,
  Skeleton,
  Text,
} from "ui";
import {
  BodyContainer,
  HeaderContainer,
  MobileMenuButton,
} from "@/components/shared";
import { RowWrapper } from "@/components/ui";
import { useTerminology, useWorkspacePath } from "@/hooks";
import type { KeyResultWithTeam } from "./types";
import { useWorkspaceKeyResultsInfinite } from "./hooks/use-workspace-key-results-infinite";
import { formatKeyResultValue, getKeyResultProgress } from "./utils";

const formatTargetDate = (value: string | null | undefined) => {
  if (!value) return "No due date";

  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? "No due date"
    : format(date, "MMM d, yy");
};

const KeyResultsHeader = () => {
  const { getTermDisplay } = useTerminology();

  return (
    <HeaderContainer>
      <Flex align="center" gap={2}>
        <MobileMenuButton />
        <BreadCrumbs
          breadCrumbs={[
            {
              name: getTermDisplay("keyResultTerm", {
                variant: "plural",
                capitalize: true,
              }),
              icon: <OKRIcon />,
            },
          ]}
        />
      </Flex>
    </HeaderContainer>
  );
};

const KeyResultsTableHeader = () => (
  <Box className="border-border bg-surface/90 sticky top-0 z-1 hidden border-b-[0.5px] py-3 md:block">
    <Flex align="center" className="min-w-4xl px-12">
      <Text className="min-w-64 flex-1" color="muted" fontWeight="medium">
        Key result
      </Text>
      <Text
        className="hidden w-64 shrink-0 lg:block"
        color="muted"
        fontWeight="medium"
      >
        Objective
      </Text>
      <Text className="w-36 shrink-0" color="muted" fontWeight="medium">
        Team
      </Text>
      <Text className="w-44 shrink-0" color="muted" fontWeight="medium">
        Progress
      </Text>
      <Text
        className="hidden w-36 shrink-0 xl:block"
        color="muted"
        fontWeight="medium"
      >
        Current / target
      </Text>
      <Text
        className="hidden w-28 shrink-0 lg:block"
        color="muted"
        fontWeight="medium"
      >
        Due
      </Text>
    </Flex>
  </Box>
);

const KeyResultRow = ({ keyResult }: { keyResult: KeyResultWithTeam }) => {
  const { withWorkspace } = useWorkspacePath();
  const progress = getKeyResultProgress(keyResult);
  const objectiveHref = withWorkspace(
    `/teams/${keyResult.teamId}/objectives/${keyResult.objectiveId}`,
  );

  return (
    <RowWrapper className="min-w-4xl gap-4 py-3">
      <Box className="min-w-64 flex-1">
        <Link
          className="flex min-w-0 items-center gap-2 hover:opacity-90"
          href={objectiveHref}
          prefetch
        >
          <Flex
            align="center"
            className="bg-surface-muted size-8 shrink-0 rounded-lg"
            justify="center"
          >
            <OKRIcon className="h-4" />
          </Flex>
          <Box className="min-w-0">
            <Text className="truncate">{keyResult.name}</Text>
            <Text className="truncate lg:hidden" color="muted" fontSize="sm">
              {keyResult.objectiveName}
            </Text>
          </Box>
        </Link>
      </Box>
      <Box className="hidden w-64 shrink-0 lg:block">
        <Link
          className="block truncate hover:underline"
          href={objectiveHref}
          prefetch
        >
          {keyResult.objectiveName}
        </Link>
      </Box>
      <Text className="w-36 shrink-0 truncate" color="muted">
        {keyResult.teamName}
      </Text>
      <Flex align="center" className="w-44 shrink-0" gap={2}>
        <ProgressBar className="w-24" progress={progress} />
        <Text className="w-10 shrink-0" fontWeight="medium">
          {progress}%
        </Text>
      </Flex>
      <Text className="hidden w-36 shrink-0 xl:block">
        {formatKeyResultValue(
          keyResult.currentValue,
          keyResult.measurementType,
        )}{" "}
        /{" "}
        {formatKeyResultValue(keyResult.targetValue, keyResult.measurementType)}
      </Text>
      <Text className="hidden w-28 shrink-0 lg:block" color="muted">
        {formatTargetDate(keyResult.endDate)}
      </Text>
    </RowWrapper>
  );
};

const KeyResultsSkeleton = () => (
  <BodyContainer>
    <KeyResultsTableHeader />
    {Array.from({ length: 8 }).map((_, index) => (
      <RowWrapper className="min-w-4xl gap-4 py-3" key={index}>
        <Flex align="center" className="min-w-64 flex-1" gap={2}>
          <Skeleton className="size-8 shrink-0" />
          <Skeleton className="h-5 w-56" />
        </Flex>
        <Skeleton className="hidden h-5 w-64 lg:block" />
        <Skeleton className="h-5 w-36" />
        <Skeleton className="h-5 w-44" />
        <Skeleton className="hidden h-5 w-36 xl:block" />
        <Skeleton className="hidden h-5 w-28 lg:block" />
      </RowWrapper>
    ))}
  </BodyContainer>
);

export const WorkspaceKeyResultsPage = () => {
  const { getTermDisplay } = useTerminology();
  const {
    data,
    fetchNextPage,
    hasNextPage,
    isError,
    isFetchingNextPage,
    isPending,
    refetch,
  } = useWorkspaceKeyResultsInfinite({
    pageSize: 50,
    orderBy: "updated_at",
    orderDirection: "desc",
  });

  const keyResults = data?.pages.flatMap((page) => page.keyResults) ?? [];
  const totalCount = data?.pages[0]?.totalCount ?? 0;
  const keyResultLabel = getTermDisplay("keyResultTerm", {
    variant: "plural",
  });

  const renderContent = () => {
    if (isPending) {
      return <KeyResultsSkeleton />;
    }

    if (isError) {
      return (
        <BodyContainer className="flex items-center justify-center">
          <Flex align="center" direction="column" gap={4}>
            <Text fontSize="xl" fontWeight="medium">
              We couldn&apos;t load your {keyResultLabel}
            </Text>
            <Text color="muted">Try again to refresh this workspace view.</Text>
            <Button
              color="tertiary"
              onClick={() => {
                void refetch();
              }}
            >
              Try again
            </Button>
          </Flex>
        </BodyContainer>
      );
    }

    if (keyResults.length === 0) {
      return (
        <BodyContainer className="flex items-center justify-center">
          <Flex align="center" direction="column">
            <OKRIcon className="h-12 w-auto" strokeWidth={1.5} />
            <Text className="mt-8 mb-3" fontSize="3xl">
              No {keyResultLabel} yet
            </Text>
            <Text className="max-w-md text-center" color="muted">
              Add measurable {keyResultLabel} to an objective to track whether
              your strategy is working.
            </Text>
          </Flex>
        </BodyContainer>
      );
    }

    return (
      <BodyContainer className="overflow-x-auto">
        <Box className="border-border border-b-[0.5px] px-5 py-3 md:px-12">
          <Text color="muted">
            {totalCount} {keyResultLabel} across all teams
          </Text>
        </Box>
        <KeyResultsTableHeader />
        {keyResults.map((keyResult) => (
          <KeyResultRow key={keyResult.id} keyResult={keyResult} />
        ))}
        {hasNextPage ? (
          <Flex className="py-6" justify="center">
            <Button
              color="tertiary"
              loading={isFetchingNextPage}
              loadingText="Loading key results..."
              onClick={() => {
                void fetchNextPage();
              }}
            >
              Load more
            </Button>
          </Flex>
        ) : null}
      </BodyContainer>
    );
  };

  return (
    <>
      <KeyResultsHeader />
      {renderContent()}
    </>
  );
};
