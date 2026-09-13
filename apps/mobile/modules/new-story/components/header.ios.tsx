import { Back, Badge, Row, Text } from "@/components/ui";
import React from "react";
import { Button, Host, Image, ProgressView } from "@expo/ui/swift-ui";
import {
  accessibilityLabel,
  disabled as disabledModifier,
  frame,
  glassEffect,
  opacity,
  tint,
} from "@expo/ui/swift-ui/modifiers";
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

      <Host matchContents style={{ width: 44, height: 44 }}>
        <Button
          modifiers={[
            frame({ width: 44, height: 44 }),
            accessibilityLabel(loading ? "Creating task" : "Create task"),
            disabledModifier(Boolean(disabled || loading)),
            opacity(disabled ? 0.45 : 1),
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
            <ProgressView modifiers={[tint(colors.primary)]} />
          ) : (
            <Image systemName="checkmark" size={18} />
          )}
        </Button>
      </Host>
    </Row>
  );
};
