import { Back, Badge, Row, Text } from "@/components/ui";
import React from "react";
import { colors } from "@/constants";
import { Pressable } from "react-native";
import { Ionicons } from "@expo/vector-icons";

export const Header = () => {
  return (
    <Row justify="between" align="center" className="mb-6">
      <Back />

      <Badge color="tertiary" className="px-3">
        <Text>Mobile</Text>
      </Badge>

      <Pressable
        style={{
          width: 40,
          height: 40,
          borderRadius: 20,
          backgroundColor: "rgba(255, 255, 255, 0.8)",
          justifyContent: "center",
          alignItems: "center",
          shadowColor: "#000",
          shadowOffset: {
            width: 0,
            height: 2,
          },
          shadowOpacity: 0.1,
          shadowRadius: 4,
          elevation: 3,
        }}
      >
        <Ionicons name="checkmark" size={18} color={colors.primary} />
      </Pressable>
    </Row>
  );
};
