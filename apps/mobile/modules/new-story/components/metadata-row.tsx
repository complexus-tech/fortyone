import React from "react";
import { Pressable } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { Row, Text } from "@/components/ui";
import { colors } from "@/constants";
import { useTheme } from "@/hooks";

type MetadataRowProps = {
  label: string;
  value: string;
  required?: boolean;
  onPress: () => void;
};

export const MetadataRow = ({
  label,
  value,
  required,
  onPress,
}: MetadataRowProps) => {
  const { resolvedTheme } = useTheme();
  const iconColor =
    resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[300];

  return (
    <Pressable onPress={onPress}>
      <Row
        align="center"
        justify="between"
        className="border-b border-gray-100 py-3 dark:border-dark-100"
      >
        <Text color="muted">
          {label}
          {required ? " *" : ""}
        </Text>
        <Row align="center" gap={1} className="max-w-[62%]">
          <Text align="right">{value}</Text>
          <Ionicons name="chevron-forward" size={15} color={iconColor} />
        </Row>
      </Row>
    </Pressable>
  );
};
