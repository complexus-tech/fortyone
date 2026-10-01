import type { ReactNode } from "react";
import { useState } from "react";
import { Command, Dialog, Divider, Flex, Text, commandFilter } from "ui";
import { useTerminology } from "@/hooks/use-terminology-display";
import { useDebouncedCallback } from "@/hooks/debounce";
import { useSearch } from "@/modules/search/hooks/use-search";
import type { SearchResponse } from "@/modules/search/types";
import { getStoryPath } from "@/shared/routing/story";
import { CommandSearchResults } from "./command-search-results";

const SEARCH_DEBOUNCE_MS = 250;
const SEARCH_MIN_LENGTH = 2;
const SEARCH_RESULT_LIMIT = 5;

type CommandSelection = {
  context: string;
  value: string;
};

export type CommandPaletteGroup = {
  group: string;
  items: {
    label: string;
    icon: ReactNode;
    shortcut: ReactNode;
    disabled?: boolean;
    action: () => void | Promise<void>;
  }[];
};

const getMatchingCommands = (
  commands: CommandPaletteGroup[],
  query: string,
) => {
  if (!query) return commands;

  return commands
    .map((command) => ({
      ...command,
      items: command.items.filter(
        (item) => commandFilter(item.label, query) > 0,
      ),
    }))
    .filter((command) => command.items.length > 0);
};

const getFirstSearchResultValue = (
  results: SearchResponse | undefined,
  query: string,
) => {
  const story = results?.stories.at(0);
  if (story) return `story ${story.id}`;

  const objective = results?.objectives.at(0);
  if (objective) return `objective ${objective.id}`;

  return `search ${query}`;
};

export const CommandPaletteContent = ({
  commands,
  onNavigate,
}: {
  commands: CommandPaletteGroup[];
  onNavigate: (path: string) => void;
}) => {
  const { getTermDisplay } = useTerminology();
  const storyTerm = getTermDisplay("storyTerm", { variant: "plural" });
  const objectiveTerm = getTermDisplay("objectiveTerm", { variant: "plural" });
  const searchLabel = `Search ${storyTerm}, ${objectiveTerm}, or commands`;
  const [inputValue, setInputValue] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const [selection, setSelection] = useState<CommandSelection>({
    context: "",
    value: "",
  });
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

  const matchingCommands = getMatchingCommands(commands, normalizedInput);
  const enabledCommandValues = matchingCommands.flatMap((command) =>
    command.items
      .filter((item) => !item.disabled)
      .map((item) => `command ${item.label}`),
  );
  const settledSearchResults =
    hasSettledQuery && !isSearchFetching && !isSearchError
      ? searchResults
      : undefined;
  const firstSearchResultValue = getFirstSearchResultValue(
    settledSearchResults,
    normalizedInput,
  );
  const selectableValues = new Set(enabledCommandValues);
  if (hasSearchQuery) {
    selectableValues.add(`search ${normalizedInput}`);
    settledSearchResults?.stories
      .slice(0, SEARCH_RESULT_LIMIT)
      .forEach((story) => {
        selectableValues.add(`story ${story.id}`);
      });
    settledSearchResults?.objectives
      .slice(0, SEARCH_RESULT_LIMIT)
      .forEach((objective) => {
        selectableValues.add(`objective ${objective.id}`);
      });
  }
  const firstSelectableValue =
    enabledCommandValues.at(0) ??
    (hasSearchQuery ? firstSearchResultValue : "");
  const selectionContext = `${normalizedInput}:${firstSelectableValue}`;
  let selectedValue = firstSelectableValue;
  if (
    selection.context === selectionContext &&
    selectableValues.has(selection.value)
  ) {
    selectedValue = selection.value;
  }

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
      <Command
        loop
        onValueChange={(value: string) => {
          setSelection({ context: selectionContext, value });
        }}
        shouldFilter={false}
        value={selectedValue}
      >
        <Command.Input
          className="my-2.5 text-2xl antialiased"
          icon={null}
          onValueChange={handleInputValueChange}
          placeholder={`${searchLabel}…`}
          value={inputValue}
        />
        <Divider className="my-2.5" />
        <Command.List className="mt-0 max-h-140 w-full overflow-y-auto border-0 bg-transparent px-3 pt-2 pb-0 shadow-none backdrop-blur-none dark:bg-transparent">
          {matchingCommands.map((command) => (
            <Command.Group
              className="mb-4 px-0"
              heading={
                <Text className="mb-1.5 pl-3 dark:antialiased" color="muted">
                  {command.group}
                </Text>
              }
              key={command.group}
            >
              {command.items.map((item) => (
                <Command.Item
                  className="justify-between rounded-lg p-3 text-[1.1rem] opacity-85"
                  disabled={item.disabled}
                  key={item.label}
                  onSelect={item.action}
                  value={`command ${item.label}`}
                >
                  <Flex
                    align="center"
                    className="font-medium antialiased"
                    gap={3}
                  >
                    {item.icon}
                    {item.label}
                  </Flex>
                  {item.shortcut}
                </Command.Item>
              ))}
            </Command.Group>
          ))}
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
              results={settledSearchResults}
            />
          ) : null}
        </Command.List>
      </Command>
    </Dialog.Body>
  );
};
