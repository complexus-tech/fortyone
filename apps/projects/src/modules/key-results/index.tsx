"use client";

import { useState } from "react";
import { Box, BreadCrumbs, Button, Flex, Skeleton, Text } from "ui";
import { OKRIcon } from "icons";
import {
  BodyContainer,
  HeaderContainer,
  MobileMenuButton,
} from "@/components/shared";
import { BoardDividedPanel, RowWrapper } from "@/components/ui";
import { RoadmapEmptyIllustration } from "@/components/ui/illustrations/empty-state-illustrations";
import { useMediaQuery, useTerminology } from "@/hooks";
import { useMembers } from "@/lib/hooks/members";
import { useTeams } from "@/modules/teams/hooks/teams";
import {
  KeyResultsFilterButton,
  KeyResultsFilterBar,
} from "./components/key-results-filter-button";
import { KeyResultsList } from "./components/key-results-list";
import { KeyResultsReportSidebar } from "./components/key-results-report-sidebar";
import { KeyResultsToolbar } from "./components/key-results-toolbar";
import { useWorkspaceKeyResultsInfinite } from "./hooks/use-workspace-key-results-infinite";
import { getActiveKeyResultFilterCount } from "./key-results-filter-utils";
import type { KeyResultFilters } from "./types";
import { groupKeyResultsByObjective } from "./utils";

const KeyResultsHeader = ({
  filters,
  setFilters,
}: {
  filters: KeyResultFilters;
  setFilters: (filters: KeyResultFilters) => void;
}) => {
  const { getTermDisplay } = useTerminology();

  return (
    <HeaderContainer className="justify-between gap-4">
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
      <Flex align="center" gap={2}>
        <KeyResultsFilterButton filters={filters} setFilters={setFilters} />
      </Flex>
    </HeaderContainer>
  );
};

const KeyResultsSkeleton = () => (
  <BodyContainer>
    {Array.from({ length: 3 }).map((_, groupIndex) => (
      <Box key={groupIndex}>
        <Box className="border-border bg-surface-muted/85 border-b-[0.5px] px-5 py-3 md:px-12">
          <Flex align="center" justify="between">
            <Skeleton className="h-5 w-64" />
            <Skeleton className="h-5 w-24" />
          </Flex>
        </Box>
        {Array.from({ length: groupIndex === 0 ? 3 : 2 }).map((_, rowIndex) => (
          <RowWrapper className="pointer-events-none gap-4 py-3" key={rowIndex}>
            <Flex align="center" className="min-w-0 flex-1" gap={2}>
              <Skeleton className="size-5 shrink-0" />
              <Skeleton className="h-5 w-56" />
            </Flex>
            <Skeleton className="h-5 w-40" />
          </RowWrapper>
        ))}
      </Box>
    ))}
  </BodyContainer>
);

export const WorkspaceKeyResultsPage = () => {
  const { getTermDisplay } = useTerminology();
  const isMobile = useMediaQuery("(max-width: 768px)");
  const [filters, setFilters] = useState<KeyResultFilters>({});
  const [selectedKeyResultIds, setSelectedKeyResultIds] = useState<Set<string>>(
    () => new Set(),
  );
  const { data: members = [] } = useMembers();
  const { data: teams = [] } = useTeams();
  const {
    data,
    fetchNextPage,
    hasNextPage,
    isError,
    isFetchingNextPage,
    isPending,
    refetch,
  } = useWorkspaceKeyResultsInfinite({
    ...filters,
    pageSize: 50,
    orderBy: "objective_name",
    orderDirection: "asc",
  });

  const keyResults = data?.pages.flatMap((page) => page.keyResults) ?? [];
  const groups = groupKeyResultsByObjective(keyResults);
  const memberById = new Map(members.map((member) => [member.id, member]));
  const teamColorById = new Map(teams.map((team) => [team.id, team.color]));
  const totalCount = data?.pages[0]?.totalCount ?? 0;
  const activeFilterCount = getActiveKeyResultFilterCount(filters);
  const keyResultLabel = getTermDisplay("keyResultTerm", {
    variant: "plural",
  });

  const header = (
    <>
      <KeyResultsHeader
        filters={filters}
        setFilters={(nextFilters) => {
          setFilters(nextFilters);
          setSelectedKeyResultIds(new Set());
        }}
      />
      <KeyResultsFilterBar
        filters={filters}
        setFilters={(nextFilters) => {
          setFilters(nextFilters);
          setSelectedKeyResultIds(new Set());
        }}
      />
    </>
  );

  if (isPending) {
    return (
      <>
        {header}
        <KeyResultsSkeleton />
      </>
    );
  }

  if (isError) {
    return (
      <>
        {header}
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
      </>
    );
  }

  if (keyResults.length === 0 && activeFilterCount === 0) {
    return (
      <>
        {header}
        <BodyContainer className="flex items-center justify-center">
          <Flex align="center" direction="column">
            <RoadmapEmptyIllustration />
            <Text className="mt-8 mb-3" fontSize="3xl">
              No {keyResultLabel} yet
            </Text>
            <Text className="max-w-md text-center" color="muted">
              Add measurable {keyResultLabel} to an objective to track whether
              your strategy is working.
            </Text>
          </Flex>
        </BodyContainer>
      </>
    );
  }

  const list = (
    <BodyContainer className="overflow-x-auto">
      {keyResults.length > 0 ? (
        <KeyResultsList
          groups={groups}
          memberById={memberById}
          selectedKeyResultIds={selectedKeyResultIds}
          setSelectedKeyResultIds={setSelectedKeyResultIds}
          teamColorById={teamColorById}
        />
      ) : (
        <Flex
          align="center"
          className="min-h-64"
          direction="column"
          justify="center"
        >
          <Text fontSize="lg" fontWeight="medium">
            No matching {keyResultLabel}
          </Text>
          <Text className="mt-1" color="muted">
            Clear or adjust the filters to see more results.
          </Text>
          <Button
            className="mt-4"
            color="tertiary"
            onClick={() => {
              setFilters({});
              setSelectedKeyResultIds(new Set());
            }}
            size="sm"
          >
            Clear filters
          </Button>
        </Flex>
      )}
      {hasNextPage ? (
        <Flex className="py-6" justify="center">
          <Button
            color="tertiary"
            loading={isFetchingNextPage}
            loadingText={`Loading ${keyResultLabel}...`}
            onClick={() => {
              void fetchNextPage();
            }}
          >
            Load more
          </Button>
        </Flex>
      ) : null}
      {selectedKeyResultIds.size > 0 ? (
        <KeyResultsToolbar
          clearSelection={() => {
            setSelectedKeyResultIds(new Set());
          }}
          selectedKeyResults={keyResults.filter(({ id }) =>
            selectedKeyResultIds.has(id),
          )}
        />
      ) : null}
    </BodyContainer>
  );

  return (
    <>
      {header}
      {isMobile ? (
        <>
          {list}
          <KeyResultsReportSidebar
            keyResults={keyResults}
            memberById={memberById}
            teamColorById={teamColorById}
            totalCount={totalCount}
          />
        </>
      ) : (
        <BoardDividedPanel autoSaveId="workspace:key-results:divided-panel">
          <BoardDividedPanel.MainPanel>{list}</BoardDividedPanel.MainPanel>
          <BoardDividedPanel.SideBar isExpanded>
            <KeyResultsReportSidebar
              keyResults={keyResults}
              memberById={memberById}
              teamColorById={teamColorById}
              totalCount={totalCount}
            />
          </BoardDividedPanel.SideBar>
        </BoardDividedPanel>
      )}
    </>
  );
};
