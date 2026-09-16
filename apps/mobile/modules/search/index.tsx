import { QueryState } from "@/components/ui/query-state";
import React, { useEffect, useMemo, useState } from "react";
import { ScrollView, View } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { SafeContainer, StoriesSkeleton, Text } from "@/components/ui";
import { useFeatures, useTheme, useTerminology } from "@/hooks";
import { themeColors } from "@/constants/colors";
import { Header } from "./components/header";
import { SearchInput } from "./components/search-input";
import { SafeAreaView as NativeTabSafeAreaView } from "react-native-screens/experimental";
import { SearchResults } from "./components/search-results";
import { useSearch } from "./hooks";
import { mergeSearchPages } from "./pagination";

import { ObjectivesSkeleton } from "@/modules/objectives/components";
import {
  KeyboardController,
  KeyboardAvoidingView,
} from "react-native-keyboard-controller";

export const Search = () => {
  const { resolvedTheme } = useTheme();
  const { getTermDisplay } = useTerminology();
  const [searchType, setSearchType] = useState<"stories" | "objectives">(
    "stories",
  );
  const [input, setInput] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const { objectiveEnabled } = useFeatures();
  const effectiveType = objectiveEnabled ? searchType : "stories";
  const trimmedInput = input.trim();
  const effectiveQuery = trimmedInput ? searchQuery : "";
  const debouncing = trimmedInput !== effectiveQuery;
  useEffect(() => {
    const timer = setTimeout(() => setSearchQuery(trimmedInput), 300);
    return () => clearTimeout(timer);
  }, [trimmedInput]);
  const {
    data,
    isFetching,
    isFetchingNextPage,
    hasNextPage,
    fetchNextPage,
    isFetchNextPageError,
    isPending,
    error,
    refetch,
  } = useSearch({ query: effectiveQuery, type: effectiveType, pageSize: 20 });
  const results = useMemo(() => mergeSearchPages(data?.pages), [data?.pages]);

  useEffect(() => {
    KeyboardController.preload(); // warms up keyboard before it opens
  }, []);

  return (
    <SafeContainer isFull>
      <NativeTabSafeAreaView
        style={{ flex: 1 }}
        edges={{ top: false, bottom: true, left: false, right: false }}
      >
        <Header
          objectivesEnabled={objectiveEnabled}
          searchType={effectiveType}
          setSearchType={setSearchType}
        />
        <KeyboardAvoidingView
          automaticOffset
          behavior="padding"
          style={{ flex: 1 }}
        >
          <View style={{ flex: 1 }}>
            {!trimmedInput ? (
              <ScrollView
                contentInsetAdjustmentBehavior="automatic"
                keyboardShouldPersistTaps="handled"
                keyboardDismissMode="on-drag"
                showsVerticalScrollIndicator={false}
                contentContainerStyle={{
                  flexGrow: 1,
                  alignItems: "center",
                  paddingHorizontal: 20,
                  paddingVertical: 24,
                }}
              >
                <View style={{ flex: 1 }} />
                <View
                  className="mb-[16px] size-[48px] items-center justify-center rounded-2xl"
                  style={{
                    backgroundColor: themeColors[resolvedTheme].surfaceMuted,
                  }}
                >
                  <Ionicons
                    name="search-outline"
                    size={24}
                    color={themeColors[resolvedTheme].textMuted}
                  />
                </View>
                <Text fontSize="xl" fontWeight="semibold" className="mb-2">
                  Find your way
                </Text>
                <Text color="muted" align="center" style={{ maxWidth: 290 }}>
                  Search {getTermDisplay("storyTerm", { variant: "plural" })}
                  {objectiveEnabled
                    ? ` and ${getTermDisplay("objectiveTerm", { variant: "plural" })}`
                    : ""}{" "}
                  by name or keyword.
                </Text>
                <View style={{ flex: 3 }} />
              </ScrollView>
            ) : debouncing || isPending ? (
              effectiveType === "stories" ? (
                <StoriesSkeleton count={5} />
              ) : (
                <ObjectivesSkeleton count={5} />
              )
            ) : error && !results ? (
              <QueryState
                title="Search could not be completed"
                message={error.message}
                onRetry={() => {
                  void refetch();
                }}
              />
            ) : results ? (
              <SearchResults
                key={`${effectiveType}:${effectiveQuery}`}
                results={results}
                type={effectiveType}
                query={effectiveQuery}
                hasMore={hasNextPage}
                loadingMore={isFetchingNextPage}
                error={error}
                onRetry={() => {
                  void refetch();
                }}
                loadMoreError={isFetchNextPageError}
                onLoadMore={() => {
                  if (!isFetching && hasNextPage) void fetchNextPage();
                }}
              />
            ) : null}
          </View>
          <View
            style={{ paddingHorizontal: 16, paddingTop: 8, paddingBottom: 12 }}
          >
            <SearchInput
              value={input}
              onChangeText={setInput}
              onSubmit={() => setSearchQuery(trimmedInput)}
              searching={Boolean(trimmedInput) && (debouncing || isFetching)}
              searchType={effectiveType}
            />
          </View>
        </KeyboardAvoidingView>
      </NativeTabSafeAreaView>
    </SafeContainer>
  );
};
