import type { ReactNode } from "react";
import { useState } from "react";
import { Box, Command, Dialog, Divider } from "ui";
import { useDebouncedCallback } from "@/hooks/debounce";
import { useSearch } from "@/modules/search/hooks/use-search";
import { getStoryPath } from "@/modules/story/utils/story-url";
import { CommandSearchResults } from "./command-search-results";

const SEARCH_DEBOUNCE_MS = 250;
const SEARCH_MIN_LENGTH = 2;
const SEARCH_RESULT_LIMIT = 5;

export const CommandPaletteContent = ({
  children,
  onNavigate,
}: {
  children: ReactNode;
  onNavigate: (path: string) => void;
}) => {
  const [inputValue, setInputValue] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const { callback: queueSearch, cancel: cancelSearch } =
    useDebouncedCallback<string>(setSearchQuery, SEARCH_DEBOUNCE_MS);

  const normalizedInput = inputValue.trim();
  const hasSearchQuery = normalizedInput.length >= SEARCH_MIN_LENGTH;
  const hasSettledQuery = hasSearchQuery && searchQuery === normalizedInput;
  const {
    data: searchResults,
    isError: isSearchError,
    isFetching: isSearchFetching,
  } = useSearch({
    pageSize: SEARCH_RESULT_LIMIT,
    query: searchQuery.length >= SEARCH_MIN_LENGTH ? searchQuery : "",
    type: "all",
  });

  const handleInputValueChange = (value: string) => {
    setInputValue(value);

    const nextQuery = value.trim();
    if (nextQuery.length < SEARCH_MIN_LENGTH) {
      cancelSearch();
      setSearchQuery("");
      return;
    }

    queueSearch(nextQuery);
  };

  return (
    <Dialog.Body className="px-0 pt-2 pb-0">
      <Command>
        <Command.Input
          className="my-2.5 text-2xl antialiased"
          icon={null}
          onValueChange={handleInputValueChange}
          placeholder="Search tasks, objectives, or commands…"
          value={inputValue}
        />
        <Divider className="my-2.5" />
        <Box className="max-h-140 overflow-y-auto px-3 pt-2">
          {hasSearchQuery ? (
            <CommandSearchResults
              hasSettledQuery={Boolean(hasSettledQuery && !isSearchFetching)}
              isError={Boolean(hasSettledQuery && isSearchError)}
              isLoading={!hasSettledQuery || isSearchFetching}
              onSelectObjective={(objective) => {
                onNavigate(
                  `/teams/${encodeURIComponent(objective.teamId)}/objectives/${encodeURIComponent(objective.id)}`,
                );
              }}
              onSelectStory={(story) => {
                onNavigate(getStoryPath(story));
              }}
              onViewAll={() => {
                const params = new URLSearchParams({
                  query: normalizedInput,
                  type: "all",
                });
                onNavigate(`/search?${params.toString()}`);
              }}
              query={normalizedInput}
              results={hasSettledQuery ? searchResults : undefined}
            />
          ) : null}
          {children}
        </Box>
      </Command>
    </Dialog.Body>
  );
};
