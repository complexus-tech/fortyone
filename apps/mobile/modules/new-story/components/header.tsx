import { Back, Badge, Row, Text } from "@/components/ui";
import React from "react";
import { colors } from "@/constants";
import { ActivityIndicator, Pressable } from "react-native";
import { Ionicons } from "@expo/vector-icons";

type HeaderProps = {
  disabled?: boolean;
  loading?: boolean;
  onSubmit: () => void;
};

export const Header = ({ disabled, loading, onSubmit }: HeaderProps) => {
  return (
    <Row justify="between" align="center" className="mb-6">
      <Back />

      <Badge color="tertiary" className="px-3">
        <Text>Create task</Text>
      </Badge>

      <Pressable
        disabled={disabled || loading}
        onPress={onSubmit}
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
          opacity: disabled ? 0.45 : 1,
        }}
      >
        {loading ? (
          <ActivityIndicator size="small" color={colors.primary} />
        ) : (
          <Ionicons name="checkmark" size={18} color={colors.primary} />
        )}
      </Pressable>
    </Row>
  );
};
