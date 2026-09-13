import { QueryState } from "@/components/ui/query-state";
import React, { useEffect, useState } from "react";
import { SafeContainer, StoriesSkeleton } from "@/components/ui";
import { Header } from "./components/header";
import { SearchResults } from "./components/search-results";
import { useSearch } from "./hooks";
import type { SearchQueryParams } from "./types";

import { ObjectivesSkeleton } from "@/modules/objectives/components";
import {
  KeyboardAwareScrollView,
  KeyboardToolbar,
  KeyboardController,
} from "react-native-keyboard-controller";

export const Search = () => {
  const [searchType, setSearchType] = useState<"stories" | "objectives">(
    "stories",
  );
  const [searchQuery, setSearchQuery] = useState("");
  const {
    data: results,
    isPending,
    error,
    refetch,
  } = useSearch({ query: searchQuery, type: searchType });

  useEffect(() => {
    KeyboardController.preload(); // warms up keyboard before it opens
  }, []);

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
      {isPending && searchQuery ? (
        searchType === "stories" ? (
          <StoriesSkeleton count={8} />
        ) : (
          <ObjectivesSkeleton count={8} />
        )
      ) : error ? (
        <QueryState
          title="Search could not be completed"
          message={error.message}
          onRetry={() => {
            void refetch();
          }}
        />
      ) : null}
      <KeyboardAwareScrollView
        contentContainerStyle={{
          paddingBottom: 80,
        }}
        bottomOffset={62}
        style={{ flex: 1 }}
      >
        {!isPending && !error && results ? (
          <SearchResults results={results} type={searchType} />
        ) : null}
      </KeyboardAwareScrollView>
      <KeyboardToolbar doneText="Close" showArrows={false} />
    </SafeContainer>
  );
};
