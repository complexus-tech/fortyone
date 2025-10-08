import React, { useState } from "react";
import { TextInput, Pressable } from "react-native";
import { Col, ContextMenuButton, Row, Text } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants/colors";
import type { SearchQueryParams } from "../types";
import {
  Host,
  HStack,
  Spacer,
  Image,
  Text as SwiftUIText,
} from "@expo/ui/swift-ui";
import { opacity } from "@expo/ui/swift-ui/modifiers";
import { useColorScheme } from "nativewind";
import { useTerminology } from "@/hooks/use-terminology";

type HeaderProps = {
  onSearch: (params: SearchQueryParams) => void;
  searchType: "stories" | "objectives";
  setSearchType: (type: "stories" | "objectives") => void;
};

export const Header = ({
  onSearch,
  searchType,
  setSearchType,
}: HeaderProps) => {
  const { colorScheme } = useColorScheme();
  const [searchTerm, setSearchTerm] = useState("");
  const { getTermDisplay } = useTerminology();

  const handleSubmit = () => {
    if (searchTerm.trim()) {
      onSearch({ query: searchTerm.trim() });
    }
  };

  return (
    <Col className="mb-2" asContainer>
      <Row justify="between" align="center">
        <Row className="flex-1">
          <Text fontSize="2xl" fontWeight="semibold">
            Search
          </Text>
        </Row>
        <Host
          matchContents
          style={{ width: 230, height: 40, position: "absolute", right: 0 }}
        >
          <HStack>
            <Spacer />
            <ContextMenuButton
              withNoHost
              actions={[
                {
                  label: getTermDisplay("storyTerm", {
                    variant: "plural",
                    capitalize: true,
                  }),
                  onPress: () => setSearchType("stories"),
                },
                {
                  label: getTermDisplay("objectiveTerm", {
                    variant: "plural",
                    capitalize: true,
                  }),
                  onPress: () => setSearchType("objectives"),
                },
              ]}
            >
              <HStack spacing={4}>
                <SwiftUIText
                  size={15}
                  color={
                    colorScheme === "light"
                      ? colors.dark.DEFAULT
                      : colors.gray[200]
                  }
                  weight="medium"
                >
                  {getTermDisplay(
                    searchType === "stories" ? "storyTerm" : "objectiveTerm",
                    {
                      variant: "plural",
                      capitalize: true,
                    }
                  )}
                </SwiftUIText>
                <Image
                  systemName="chevron.up.chevron.down"
                  modifiers={[opacity(0.6)]}
                  color={
                    colorScheme === "light"
                      ? colors.dark.DEFAULT
                      : colors.gray[200]
                  }
                  size={12}
                />
              </HStack>
            </ContextMenuButton>
          </HStack>
        </Host>
      </Row>
      <Row
        className="bg-gray-50 dark:bg-dark-100/60 rounded-full pl-3 pr-2.5 mt-3"
        align="center"
        gap={2}
      >
        <SymbolView
          name="magnifyingglass"
          size={20}
          tintColor={
            colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300]
          }
        />
        <TextInput
          className="flex-1 h-12 font-medium text-[16px] dark:text-white"
          placeholder={`Search ${getTermDisplay("storyTerm", {
            variant: "plural",
          })} and ${getTermDisplay("objectiveTerm", {
            variant: "plural",
          })}...`}
          placeholderTextColor={
            colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300]
          }
          value={searchTerm}
          onChangeText={setSearchTerm}
          onSubmitEditing={handleSubmit}
          returnKeyType="search"
          autoFocus
        />
        {searchTerm.length > 0 && (
          <Pressable
            onPress={() => {
              setSearchTerm("");
              onSearch({ query: "" });
            }}
            className="p-1"
          >
            <SymbolView
              name="xmark.circle.fill"
              size={20}
              tintColor={
                colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300]
              }
            />
          </Pressable>
        )}
      </Row>
    </Col>
  );
};
