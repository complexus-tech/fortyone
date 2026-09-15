import React from "react";
import { useRouter } from "expo-router";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks";
import { Host, Button, Image } from "@expo/ui/swift-ui";
import {
  frame,
  glassEffect,
  accessibilityLabel,
  buttonStyle,
} from "@expo/ui/swift-ui/modifiers";

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
    <Host matchContents style={{ width: 44, height: 44 }}>
      <Button
        modifiers={[
          accessibilityLabel("Back"),
          buttonStyle("plain"),
          frame({ width: 44, height: 44 }),
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
          color={themeColors[resolvedTheme].foreground}
        />
      </Button>
    </Host>
  );
};
