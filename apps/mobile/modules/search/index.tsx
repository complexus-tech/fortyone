import React, { useState } from "react";
import { SafeContainer, Text, StoriesSkeleton, Col } from "@/components/ui";
import { Header } from "./components/header";
import { SearchResults } from "./components/search-results";
import { useSearch } from "./hooks";
import type { SearchQueryParams } from "./types";

import { ObjectivesSkeleton } from "@/modules/objectives/components";

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
        <Col align="center" justify="center" className="flex-1" asContainer>
          <Text color="muted" className="mt-8 text-center">
            Start typing to search for stories and objectives
          </Text>
        </Col>
      ) : isPending ? (
        searchType === "stories" ? (
          <StoriesSkeleton count={8} />
        ) : (
          <ObjectivesSkeleton count={8} />
        )
      ) : results ? (
        <SearchResults results={results} type={searchType} />
      ) : null}
    </SafeContainer>
  );
};
