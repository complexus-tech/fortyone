import React from "react";
import { useRouter } from "expo-router";
import { colors } from "@/constants";
import { useTheme } from "@/hooks";
import { Host, HStack, Image } from "@expo/ui/swift-ui";
import { frame, glassEffect } from "@expo/ui/swift-ui/modifiers";

export const Back = () => {
  const { resolvedTheme } = useTheme();
  const router = useRouter();
  const canGoBack = router.canGoBack();

  const handleBack = () => {
    if (canGoBack) {
      router.back();
    } else {
      router.replace("/");
    }
  };

  return (
    <Host matchContents style={{ width: 45, height: 45 }}>
      <HStack
        modifiers={[
          frame({ width: 45, height: 45 }),
          glassEffect({
            glass: {
              interactive: true,
              variant: "regular",
            },
          }),
        ]}
        onPress={handleBack}
      >
        <Image
          systemName="chevron.backward"
          size={20}
          color={resolvedTheme === "light" ? colors.dark[50] : colors.gray[200]}
        />
      </HStack>
    </Host>
  );
};
