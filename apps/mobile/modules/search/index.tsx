import React, { useState } from "react";
import { SafeContainer, Text } from "@/components/ui";
import { Header } from "./components/header";
import { SearchResults } from "./components/search-results";
import { useSearch } from "./hooks";
import type { SearchQueryParams } from "./types";
import { View } from "react-native";

export const Search = () => {
  const [searchType, setSearchType] = useState<"stories" | "objectives">(
    "stories"
  );
  const [searchQuery, setSearchQuery] = useState("");

  const { data: results, isPending } = useSearch({ query: searchQuery });

  const handleSearch = (params: SearchQueryParams) => {
    setSearchQuery(params.query || "");
  };

  return (
    <SafeContainer isFull>
      <Header
        onSearch={handleSearch}
        searchType={searchType}
        setSearchType={setSearchType}
      />
      {!searchQuery ? (
        <View className="flex-1 justify-center items-center">
          <Text color="muted" className="mt-8 text-center">
            Start typing to search for stories and objectives
          </Text>
        </View>
      ) : isPending ? (
        <Text color="muted" className="mt-4 text-center">
          Searching...
        </Text>
      ) : results ? (
        <SearchResults results={results} type={searchType} />
      ) : null}
    </SafeContainer>
  );
};
