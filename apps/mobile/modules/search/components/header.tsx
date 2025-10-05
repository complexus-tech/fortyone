import React, { useState } from "react";
import { TextInput, Pressable } from "react-native";
import { Col, Row, Text } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants/colors";
import type { SearchQueryParams } from "../types";

type HeaderProps = {
  onSearch: (params: SearchQueryParams) => void;
};

export const Header = ({ onSearch }: HeaderProps) => {
  const [searchTerm, setSearchTerm] = useState("");

  const handleSubmit = () => {
    if (searchTerm.trim()) {
      onSearch({ query: searchTerm.trim() });
    }
  };

  return (
    <Col className="mb-2" asContainer>
      <Row justify="between" align="center">
        <Text fontSize="2xl" fontWeight="semibold" className="mb-2 flex-1">
          Search
        </Text>
      </Row>
      <Row
        className="bg-gray-50 dark:bg-dark-300 rounded-2xl px-2"
        align="center"
        gap={2}
      >
        <SymbolView
          name="magnifyingglass"
          size={20}
          tintColor={colors.gray.DEFAULT}
        />
        <TextInput
          className="flex-1 h-12 dark:text-white"
          placeholder="Search stories and objectives..."
          placeholderTextColor={colors.gray.DEFAULT}
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
              tintColor={colors.gray.DEFAULT}
            />
          </Pressable>
        )}
      </Row>
    </Col>
  );
};
