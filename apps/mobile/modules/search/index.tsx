import React, { useState } from "react";
import { SafeContainer, Text, Tabs } from "@/components/ui";
import { Header } from "./components/header";
import { SearchResults } from "./components/search-results";
import { useSearch } from "./hooks";
import type { SearchQueryParams } from "./types";

type SearchTab = "all" | "stories" | "objectives";

export const Search = () => {
  const [activeTab, setActiveTab] = useState<SearchTab>("all");
  const [searchQuery, setSearchQuery] = useState("");

  const { data: results, isPending } = useSearch({ query: searchQuery });

  const handleSearch = (params: SearchQueryParams) => {
    setSearchQuery(params.query || "");
  };

  return (
    <SafeContainer>
      <Header onSearch={handleSearch} />

      <Tabs
        defaultValue={activeTab}
        onValueChange={(value) => setActiveTab(value as SearchTab)}
      >
        <Tabs.List>
          <Tabs.Tab value="all">All</Tabs.Tab>
          <Tabs.Tab value="stories">Stories</Tabs.Tab>
          <Tabs.Tab value="objectives">Objectives</Tabs.Tab>
        </Tabs.List>

        <Tabs.Panel value="all">
          {!searchQuery ? (
            <Text color="muted" className="mt-8 text-center">
              Start typing to search for stories and objectives
            </Text>
          ) : isPending ? (
            <Text color="muted" className="mt-4 text-center">
              Searching...
            </Text>
          ) : results ? (
            <SearchResults results={results} type="all" />
          ) : null}
        </Tabs.Panel>

        <Tabs.Panel value="stories">
          {!searchQuery ? (
            <Text color="muted" className="mt-8 text-center">
              Start typing to search for stories
            </Text>
          ) : isPending ? (
            <Text color="muted" className="mt-4 text-center">
              Searching...
            </Text>
          ) : results ? (
            <SearchResults results={results} type="stories" />
          ) : null}
        </Tabs.Panel>

        <Tabs.Panel value="objectives">
          {!searchQuery ? (
            <Text color="muted" className="mt-8 text-center">
              Start typing to search for objectives
            </Text>
          ) : isPending ? (
            <Text color="muted" className="mt-4 text-center">
              Searching...
            </Text>
          ) : results ? (
            <SearchResults results={results} type="objectives" />
          ) : null}
        </Tabs.Panel>
      </Tabs>
    </SafeContainer>
  );
};
