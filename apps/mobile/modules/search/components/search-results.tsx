import React from "react";
import { ScrollView } from "react-native";
import { SafeContainer, Text } from "@/components/ui";
import { Card as ObjectiveCard } from "@/modules/objectives/components/card";
import type { SearchResponse } from "../types";

type SearchResultsProps = {
  results: SearchResponse;
  type: "all" | "stories" | "objectives";
};

export const SearchResults = ({ results, type }: SearchResultsProps) => {
  const showStories = type === "all" || type === "stories";
  const showObjectives = type === "all" || type === "objectives";

  if (type === "all") {
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
          {type === "all" && (
            <Text fontSize="lg" fontWeight="semibold" className="px-4 py-2">
              Stories ({results.totalStories})
            </Text>
          )}
          {results.stories.map((story) => (
            <Text key={story.id} className="px-4 py-2 border-t border-gray-100">
              {story.title}
            </Text>
          ))}
        </>
      )}

      {showObjectives && results.objectives.length > 0 && (
        <>
          {type === "all" && (
            <Text fontSize="lg" fontWeight="semibold" className="px-4 py-2">
              Objectives ({results.totalObjectives})
            </Text>
          )}
          {results.objectives.map((objective) => (
            <ObjectiveCard key={objective.id} objective={objective} />
          ))}
        </>
      )}

      {type !== "all" && (
        <>
          {type === "stories" && results.stories.length === 0 && (
            <Text color="muted" className="mt-4 text-center">
              No stories found
            </Text>
          )}
          {type === "objectives" && results.objectives.length === 0 && (
            <Text color="muted" className="mt-4 text-center">
              No objectives found
            </Text>
          )}
        </>
      )}
    </ScrollView>
  );
};
