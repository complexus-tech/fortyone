import React, { useState } from "react";
import { TextInput, Pressable } from "react-native";
import { Row, Text } from "@/components/ui";
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
    <Row
      align="center"
      className="pb-4 bg-gray-50 rounded-lg px-3 py-2"
      gap={2}
    >
      <SymbolView
        name="magnifyingglass"
        size={18}
        tintColor={colors.gray.DEFAULT}
      />
      <TextInput
        className="flex-1 text-base"
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
            size={16}
            tintColor={colors.gray.DEFAULT}
          />
        </Pressable>
      )}
    </Row>
  );
};
