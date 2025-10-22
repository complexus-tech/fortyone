import { Back, Badge, Row, Text } from "@/components/ui";
import React from "react";
import { Host, HStack, Image } from "@expo/ui/swift-ui";
import { frame, glassEffect } from "@expo/ui/swift-ui/modifiers";

export const Header = () => {
  return (
    <Row justify="between" align="center" className="mb-6">
      <Back />

      <Badge color="tertiary" className="px-3">
        <Text>Mobile</Text>
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
        >
          <Image systemName="checkmark" size={18} />
        </HStack>
      </Host>
    </Row>
  );
};
