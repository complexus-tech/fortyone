import React from "react";
import { Pressable } from "react-native";
import { SymbolView } from "expo-symbols";
import { Avatar, Row, Text } from "@/components/ui";
import { useRouter } from "expo-router";
import { colors } from "@/constants";

export const Header = () => {
  const router = useRouter();

  return (
    <Row align="center" justify="between" className="mb-3">
      <Row align="center" gap={2}>
        <Avatar
          name="John Doe"
          size="md"
          src="https://lh3.googleusercontent.com/a/ACg8ocIUt7Dv7aHtGSeygW70yxWRryGSXgddIq5NaVrg7ofoXO8uM5jt=s576-c-no"
        />
        <Text fontSize="2xl" fontWeight="semibold" numberOfLines={1}>
          Hello, Joseph
        </Text>
      </Row>
      <Pressable
        onPress={() => {
          router.push("/settings");
        }}
      >
        <SymbolView name="gear" size={28} tintColor={colors.dark[50]} />
      </Pressable>
    </Row>
  );
};
