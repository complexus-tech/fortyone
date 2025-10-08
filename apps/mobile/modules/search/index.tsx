import React, { useState } from "react";
import { SafeContainer, StoriesSkeleton } from "@/components/ui";
import { Header } from "./components/header";
import { SearchResults } from "./components/search-results";
import { useSearch } from "./hooks";
import type { SearchQueryParams } from "./types";

import { ObjectivesSkeleton } from "@/modules/objectives/components";
import {
  KeyboardAwareScrollView,
  KeyboardToolbar,
} from "react-native-keyboard-controller";

export const Search = () => {
  const [searchType, setSearchType] = useState<"stories" | "objectives">(
    "stories"
  );
  const [searchQuery, setSearchQuery] = useState("");
  const { data: results, isPending } = useSearch({ query: searchQuery });

  if (isPending && searchQuery) {
    return (
      <SafeContainer isFull>
        <Header
          onSearch={() => {}}
          searchType={searchType}
          setSearchType={setSearchType}
        />
        {searchType === "stories" ? (
          <StoriesSkeleton count={8} />
        ) : (
          <ObjectivesSkeleton count={8} />
        )}
      </SafeContainer>
    );
  }

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
      <KeyboardAwareScrollView
        contentContainerStyle={{
          paddingBottom: 80,
        }}
        bottomOffset={62}
        style={{ flex: 1 }}
      >
        {results ? <SearchResults results={results} type={searchType} /> : null}
      </KeyboardAwareScrollView>
      <KeyboardToolbar doneText="Close" />
    </SafeContainer>
  );
};
