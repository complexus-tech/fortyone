import React from "react";
import { ScrollView } from "react-native";
import { SafeContainer, Story, Text } from "@/components/ui";
import { Card as ObjectiveCard } from "@/modules/objectives/components/card";
import type { SearchResponse } from "../types";

type SearchResultsProps = {
  results: SearchResponse;
  type: "stories" | "objectives";
};

export const SearchResults = ({ results, type }: SearchResultsProps) => {
  const showStories = type === "stories";
  const showObjectives = type === "objectives";

  if (type === "stories" || type === "objectives") {
    const totalResults = results.totalStories + results.totalObjectives;
    if (totalResults === 0) {
      return (
        <SafeContainer>
          <Text color="muted" className="mt-4 text-center">
            No results found
          </Text>
        </SafeContainer>
      );
    }
  }

  return (
    <ScrollView showsVerticalScrollIndicator={false}>
      {showStories && results.stories.length > 0 && (
        <>
          {results.stories.map((story) => (
            <Story key={story.id} story={story} />
          ))}
        </>
      )}

      {showObjectives && results.objectives.length > 0 && (
        <>
          {results.objectives.map((objective) => (
            <ObjectiveCard key={objective.id} objective={objective} />
          ))}
        </>
      )}
    </ScrollView>
  );
};
