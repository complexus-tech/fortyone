import React from "react";
import { View, Pressable } from "react-native";
import { Row, Text } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";
import type { Team as TeamType } from "@/modules/teams/types";
import { useRouter } from "expo-router";

export const Team = ({ id, name, color }: TeamType) => {
  const router = useRouter();
  return (
    <Pressable
      style={({ pressed }) => [
        pressed && { backgroundColor: colors.gray[50] },
        { borderRadius: 10 },
      ]}
      onPress={() => router.push(`/teams/${id}/objectives`)}
    >
      <Row
        align="center"
        justify="between"
        className="py-3.5 px-2 min-h-[44px]"
      >
        <Row align="center">
          <View
            className="size-3.5 rounded-full mr-2"
            style={{ backgroundColor: color }}
          />
          <Text>{name}</Text>
        </Row>
        <SymbolView
          name="chevron.forward"
          weight="semibold"
          size={12}
          tintColor={colors.gray.DEFAULT}
        />
      </Row>
    </Pressable>
  );
};
