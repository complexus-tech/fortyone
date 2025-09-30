import React from "react";
import { Pressable } from "react-native";
import { Row, Text } from "@/components/ui";

export const Header = () => {
  return (
    <Row justify="between" asContainer>
      <Text fontSize="2xl" fontWeight="semibold" color="black">
        Notifications
      </Text>
      <Pressable
        className="p-2 rounded-md"
        style={({ pressed }) => [pressed && { backgroundColor: "#F2F2F7" }]}
      >
        <Text fontSize="lg" fontWeight="semibold" color="black">
          ⋯
        </Text>
      </Pressable>
    </Row>
  );
};
