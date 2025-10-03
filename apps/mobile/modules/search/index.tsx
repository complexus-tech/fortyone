import React, { useState } from "react";
import { SafeContainer, Text, Tabs } from "@/components/ui";
import { Header } from "./components/header";
import { SearchResults } from "./components/search-results";
import { useSearch } from "./hooks";
import type { SearchQueryParams, SearchResponse } from "./types";

type SearchTab = "stories" | "objectives";

type TabContentProps = {
  searchQuery: string;
  isPending: boolean;
  results: SearchResponse | null | undefined;
  type: SearchTab;
};

const TabContent = ({
  searchQuery,
  isPending,
  results,
  type,
}: TabContentProps) => {
  if (!searchQuery) {
    return (
      <Text color="muted" className="mt-8 text-center">
        Start typing to search for stories and objectives
      </Text>
    );
  }

  if (isPending) {
    return (
      <Text color="muted" className="mt-4 text-center">
        Searching...
      </Text>
    );
  }

  if (results) {
    return <SearchResults results={results} type={type} />;
  }

  return null;
};

export const Search = () => {
  const [activeTab, setActiveTab] = useState<SearchTab>("stories");
  const [searchQuery, setSearchQuery] = useState("");

  const { data: results, isPending } = useSearch({ query: searchQuery });

  const handleSearch = (params: SearchQueryParams) => {
    setSearchQuery(params.query || "");
  };

  // remove the tabs

  return (
    <SafeContainer isFull>
      <Header onSearch={handleSearch} />
      <Tabs
        defaultValue={activeTab}
        onValueChange={(value) => setActiveTab(value as SearchTab)}
      >
        <Tabs.List>
          <Tabs.Tab value="stories">Stories</Tabs.Tab>
          <Tabs.Tab value="objectives">Objectives</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="stories">
          <TabContent
            searchQuery={searchQuery}
            isPending={isPending}
            results={results}
            type="stories"
          />
        </Tabs.Panel>
        <Tabs.Panel value="objectives">
          <TabContent
            searchQuery={searchQuery}
            isPending={isPending}
            results={results}
            type="objectives"
          />
        </Tabs.Panel>
      </Tabs>
    </SafeContainer>
  );
};
