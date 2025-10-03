import React, { useState } from "react";
import { SafeContainer, Text, Tabs } from "@/components/ui";
import { Header } from "./components/header";
import { SearchResults } from "./components/search-results";
import { useSearch } from "./hooks";
import type { SearchQueryParams, SearchResponse } from "./types";

type SearchTab = "all" | "stories" | "objectives";

export const Search = () => {
  const [activeTab, setActiveTab] = useState<SearchTab>("all");
  const [searchParams, setSearchParams] = useState<SearchQueryParams>({
    query: "",
    type: undefined,
  });

  const { data: results, isPending, error } = useSearch(searchParams);

  const handleSearch = (params: SearchQueryParams) => {
    setSearchParams(params);
  };

  const handleTabChange = (tab: SearchTab) => {
    setActiveTab(tab);
    const newParams: SearchQueryParams = {
      ...searchParams,
      type: tab === "all" ? undefined : tab,
    };
    setSearchParams(newParams);
  };

  // Show empty state when no search query
  if (!searchParams.query) {
    return (
      <SafeContainer>
        <Header onSearch={handleSearch} />
        <Text color="muted" className="mt-8 text-center">
          Start typing to search for stories and objectives
        </Text>
      </SafeContainer>
    );
  }

  // Show loading state
  if (isPending) {
    return (
      <SafeContainer>
        <Header onSearch={handleSearch} />
        <Text color="muted" className="mt-4 text-center">
          Searching...
        </Text>
      </SafeContainer>
    );
  }

  // Show error state
  if (error) {
    return (
      <SafeContainer>
        <Header onSearch={handleSearch} />
        <Text color="danger" className="mt-4 text-center">
          Error searching. Please try again.
        </Text>
      </SafeContainer>
    );
  }

  // Show results
  if (results) {
    return (
      <SafeContainer>
        <Header onSearch={handleSearch} />
        <Tabs defaultValue={activeTab} onValueChange={handleTabChange}>
          <Tabs.List>
            <Tabs.Tab value="all">All</Tabs.Tab>
            <Tabs.Tab value="stories">Stories</Tabs.Tab>
            <Tabs.Tab value="objectives">Objectives</Tabs.Tab>
          </Tabs.List>
          <Tabs.Panel value="all">
            <SearchResults results={results} type="all" />
          </Tabs.Panel>
          <Tabs.Panel value="stories">
            <SearchResults results={results} type="stories" />
          </Tabs.Panel>
          <Tabs.Panel value="objectives">
            <SearchResults results={results} type="objectives" />
          </Tabs.Panel>
        </Tabs>
      </SafeContainer>
    );
  }

  return null;
};
