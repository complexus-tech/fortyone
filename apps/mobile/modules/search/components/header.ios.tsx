import React, { useState } from "react";
import { TextInput, Pressable, useWindowDimensions } from "react-native";
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
  TextField,
  VStack,
} from "@expo/ui/swift-ui";
import {
  border,
  frame,
  glassEffect,
  opacity,
  padding,
} from "@expo/ui/swift-ui/modifiers";
import { useTheme } from "@/hooks";
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
  const { resolvedTheme } = useTheme();
  const [searchTerm, setSearchTerm] = useState("");
  const { getTermDisplay } = useTerminology();
  const { width } = useWindowDimensions();

  const handleSubmit = () => {
    if (searchTerm.trim()) {
      onSearch({ query: searchTerm.trim() });
    }
  };

  return (
    <Col className="mb-2" asContainer>
      <Row justify="between" align="center" className="mb-2">
        <Row className="flex-1">
          <Text fontSize="3xl" fontWeight="semibold">
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
                  color={
                    resolvedTheme === "light"
                      ? colors.dark.DEFAULT
                      : colors.gray[200]
                  }
                  weight="medium"
                  size={16}
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
                    resolvedTheme === "light"
                      ? colors.dark.DEFAULT
                      : colors.gray[200]
                  }
                  size={11}
                />
              </HStack>
            </ContextMenuButton>
          </HStack>
        </Host>
      </Row>
      <Row
        className="bg-gray-100/60 dark:bg-dark-200 rounded-full pl-3 pr-2.5 mt-3"
        align="center"
        gap={2}
      >
        <SymbolView
          name="magnifyingglass"
          size={20}
          tintColor={
            resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[300]
          }
        />
        <TextInput
          className="flex-1 h-12 font-medium text-[17px] dark:text-white"
          placeholder={`Search ${getTermDisplay("storyTerm", {
            variant: "plural",
          })} and ${getTermDisplay("objectiveTerm", {
            variant: "plural",
          })}...`}
          placeholderTextColor={
            resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[300]
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
                resolvedTheme === "light"
                  ? colors.gray.DEFAULT
                  : colors.gray[300]
              }
            />
          </Pressable>
        )}
      </Row>
      {/* <Host
        matchContents
        style={{
          width: width - 32,
        }}
      >
        <HStack
          modifiers={[
            glassEffect({
              glass: {
                interactive: true,
                variant: "regular",
              },
            }),
          ]}
        >
          <Image
            systemName="magnifyingglass"
            size={16}
            modifiers={[padding({ leading: 12 }), opacity(0.4)]}
          />
          <VStack modifiers={[padding({ vertical: 12, horizontal: 6 })]}>
            <TextField
              autocorrection={false}
              placeholder="Search..."
              onChangeText={() => {}}
            />
          </VStack>
          <Image
            systemName="xmark.circle.fill"
            size={16}
            modifiers={[padding({ trailing: 12 }), opacity(0.4)]}
          />
        </HStack>
      </Host> */}
    </Col>
  );
};
