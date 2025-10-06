import React, { useState } from "react";
import { Row, Text } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { Pressable } from "react-native";
import { colors } from "@/constants";
import { useColorScheme } from "nativewind";
import { OptionsSheet } from "./options-sheet";

export const Header = () => {
  const { colorScheme } = useColorScheme();
  const [isOpened, setIsOpened] = useState(false);
  return (
    <>
      <Row className="mb-2" asContainer justify="between" align="center">
        <Text fontSize="2xl" fontWeight="semibold">
          My Work
        </Text>
        <Pressable
          className="p-2 rounded-xl active:bg-gray-50 dark:active:bg-dark-300"
          onPress={() => setIsOpened(true)}
        >
          <SymbolView
            name="line.3.horizontal.decrease"
            size={26}
            tintColor={
              colorScheme === "light" ? colors.dark[50] : colors.gray[300]
            }
          />
        </Pressable>
      </Row>
      <OptionsSheet isOpened={isOpened} setIsOpened={setIsOpened} />
    </>
  );
};
