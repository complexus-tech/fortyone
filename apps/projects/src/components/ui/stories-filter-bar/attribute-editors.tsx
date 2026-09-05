import { useDeferredValue, useState, type UIEvent } from "react";
import { useParams } from "next/navigation";
import { Box, Button, Calendar, Command, Divider, Flex, Text } from "ui";
import { CheckIcon, EstimateIcon, TagsIcon } from "icons";
import { formatISO } from "date-fns";
import { LABEL_MENU_PAGE_SIZE, useLabelsInfinite } from "@/lib/hooks/labels";
import {
  formatEstimate,
  getEstimateOptions,
  type EstimateScheme,
} from "@/lib/estimate";
import { MenuLoadingSkeleton } from "../menu-loading-skeleton";
import { normalizeArrayFilter, shouldFetchNextPage } from "./filter-model";
import type { StoriesFilterEditorProps } from "./types";

export const LabelEditor = ({
  filters,
  setFilters,
}: StoriesFilterEditorProps) => {
  const { teamId } = useParams<{ teamId?: string }>();
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const { data, fetchNextPage, hasNextPage, isFetching, isFetchingNextPage } =
    useLabelsInfinite({ search: deferredQuery, teamId }, LABEL_MENU_PAGE_SIZE);
  const labels = data?.pages.flatMap((page) => page.labels) ?? [];
  const isLoadingLabels = isFetching && !isFetchingNextPage;

  const toggleLabel = (labelId: string) => {
    const selected = filters.labelIds ?? [];
    const labelIds = selected.includes(labelId)
      ? selected.filter((id) => id !== labelId)
      : [...selected, labelId];
    setFilters({ ...filters, labelIds: normalizeArrayFilter(labelIds) });
  };

  const handleScroll = (event: UIEvent<HTMLDivElement>) => {
    if (
      shouldFetchNextPage(
        event.currentTarget,
        Boolean(hasNextPage),
        isFetchingNextPage,
      )
    ) {
      void fetchNextPage();
    }
  };

  return (
    <Command>
      <Command.Input
        autoFocus
        onValueChange={setQuery}
        placeholder="Search labels..."
        value={query}
      />
      <Divider className="my-2" />
      <Command.List className="w-full border-0 bg-transparent p-0 shadow-none backdrop-blur-none dark:bg-transparent">
        {!isLoadingLabels ? (
          <Command.Empty className="px-3 py-2 text-left text-base">
            <Text color="muted">No labels found.</Text>
          </Command.Empty>
        ) : null}
        <Command.Group
          className="max-h-80 overflow-y-auto"
          onScroll={handleScroll}
        >
          {isLoadingLabels ? (
            <Command.Loading className="p-2">
              <MenuLoadingSkeleton rows={5} />
            </Command.Loading>
          ) : null}
          {labels.map((label, idx) => (
            <Command.Item
              active={Boolean(filters.labelIds?.includes(label.id))}
              className="justify-between gap-4"
              key={label.id}
              onSelect={() => {
                toggleLabel(label.id);
              }}
              value={label.name}
            >
              <Flex align="center" className="min-w-0 flex-1" gap={2}>
                <TagsIcon
                  className="h-4 w-auto"
                  style={{ color: label.color }}
                />
                <Text className="max-w-48 truncate">{label.name}</Text>
              </Flex>
              <Flex align="center" className="shrink-0" gap={2}>
                {filters.labelIds?.includes(label.id) ? (
                  <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                ) : null}
                <Text color="muted">{idx}</Text>
              </Flex>
            </Command.Item>
          ))}
          {isFetchingNextPage ? (
            <Command.Loading className="p-2">
              <MenuLoadingSkeleton rows={2} />
            </Command.Loading>
          ) : null}
        </Command.Group>
      </Command.List>
    </Command>
  );
};

export const EstimateEditor = ({
  estimateScheme,
  filters,
  setFilters,
}: StoriesFilterEditorProps & { estimateScheme: EstimateScheme }) => {
  const options = getEstimateOptions(estimateScheme);

  const toggleEstimate = (estimateValue: number) => {
    const selected = filters.estimateValues ?? [];
    const estimateValues = selected.includes(estimateValue)
      ? selected.filter((value) => value !== estimateValue)
      : [...selected, estimateValue];
    setFilters({
      ...filters,
      estimateValues: normalizeArrayFilter(estimateValues),
    });
  };

  return (
    <Command>
      <Command.Input autoFocus placeholder="Change complexity..." />
      <Divider className="my-2" />
      <Command.List className="w-full border-0 bg-transparent p-0 shadow-none backdrop-blur-none dark:bg-transparent">
        <Command.Empty className="px-3 py-2 text-left text-base">
          <Text color="muted">No complexity found.</Text>
        </Command.Empty>
        <Command.Group>
          {options.map(({ label, value }, idx) => (
            <Command.Item
              active={Boolean(filters.estimateValues?.includes(value))}
              className="justify-between gap-4"
              key={value}
              onSelect={() => {
                toggleEstimate(value);
              }}
              value={label}
            >
              <Box className="grid min-w-0 flex-1 grid-cols-[24px_minmax(0,1fr)] items-center">
                <EstimateIcon className="text-text-secondary h-4 w-auto" />
                <Text className="truncate">
                  {formatEstimate(estimateScheme, value, "full")}
                </Text>
              </Box>
              <Flex align="center" className="shrink-0" gap={2}>
                {filters.estimateValues?.includes(value) ? (
                  <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                ) : null}
                <Text color="muted">{idx}</Text>
              </Flex>
            </Command.Item>
          ))}
        </Command.Group>
      </Command.List>
    </Command>
  );
};

export const DateEditor = ({
  field,
  filters,
  setFilters,
}: StoriesFilterEditorProps & { field: "startDate" | "endDate" }) => {
  const selectedDate = filters[field] ? new Date(filters[field]) : undefined;

  return (
    <Box>
      <Calendar
        className="px-3 py-3 shadow-none"
        mode="single"
        onDayClick={(date) => {
          setFilters({
            ...filters,
            [field]: formatISO(date, { representation: "date" }),
          });
        }}
        selected={selectedDate}
      />
      {selectedDate ? (
        <Button
          className="mx-3 mb-2 w-[calc(100%-1.5rem)] justify-start"
          color="tertiary"
          onClick={() => {
            setFilters({ ...filters, [field]: null });
          }}
          size="sm"
          variant="naked"
        >
          Clear date
        </Button>
      ) : null}
    </Box>
  );
};
