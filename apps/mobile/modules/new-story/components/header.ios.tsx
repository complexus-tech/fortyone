import { Back, Badge, Row, Text } from "@/components/ui";
import React from "react";
import { ActivityIndicator } from "react-native";
import { Host, HStack, Image } from "@expo/ui/swift-ui";
import { frame, glassEffect } from "@expo/ui/swift-ui/modifiers";
import { colors } from "@/constants";

type HeaderProps = {
  disabled?: boolean;
  loading?: boolean;
  onSubmit: () => void;
};

export const Header = ({ disabled, loading, onSubmit }: HeaderProps) => {
  return (
    <Row justify="between" align="center" className="mb-4">
      <Back />

      <Badge color="tertiary" className="px-3">
        <Text>Create task</Text>
      </Badge>

      <Host matchContents style={{ width: 40, height: 40 }}>
        <HStack
          modifiers={[
            frame({ width: 40, height: 40 }),
            glassEffect({
              glass: {
                interactive: true,
                variant: "regular",
              },
            }),
          ]}
          onPress={disabled || loading ? undefined : onSubmit}
        >
          {loading ? (
            <ActivityIndicator size="small" color={colors.primary} />
          ) : (
            <Image systemName="checkmark" size={18} />
          )}
        </HStack>
      </Host>
    </Row>
  );
};
