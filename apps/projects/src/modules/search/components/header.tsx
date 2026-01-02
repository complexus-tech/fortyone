"use client";
import { Flex } from "ui";
import { SearchIcon } from "icons";
import { useEffect, useRef, useState } from "react";
import { useQueryState } from "nuqs";
import { HeaderContainer } from "@/components/shared";
import { useTerminology } from "@/hooks";
import { NewStoryButton } from "@/components/ui";
import type { SearchQueryParams } from "../types";

export const Header = ({
  onSearch,
}: {
  onSearch: (params: SearchQueryParams) => void;
}) => {
  const inputRef = useRef<HTMLInputElement>(null);
  const { getTermDisplay } = useTerminology();
  const [query, setQuery] = useQueryState("query", {
    defaultValue: "",
  });
  const [searchTerm, setSearchTerm] = useState(query || "");

  const placeholder = `Search for ${getTermDisplay("storyTerm", {
    variant: "plural",
  })} and ${getTermDisplay("objectiveTerm", { variant: "plural" })}...`;

  useEffect(() => {
    if (inputRef.current) {
      inputRef.current.focus();
    }
  }, []);

  return (
    <HeaderContainer className="justify-between">
      <Flex align="center" className="w-full" gap={3}>
        <SearchIcon className="relative -top-px h-4" />
        <form
          className="w-full"
          onSubmit={(e) => {
            e.preventDefault();
            setQuery(searchTerm);
            onSearch({ query: searchTerm });
          }}
        >
          <input
            autoFocus
            className="w-9/12 bg-transparent placeholder:text-text-muted/90"
            onChange={(e) => {
              setSearchTerm(e.target.value);
            }}
            placeholder={placeholder}
            ref={inputRef}
            type="text"
            value={searchTerm}
          />
        </form>
      </Flex>
      <NewStoryButton />
    </HeaderContainer>
  );
};
