import React from "react";
import { useRouter } from "expo-router";
import { Button, Host } from "@expo/ui/swift-ui";
import { aspectRatio, frame } from "@expo/ui/swift-ui/modifiers";

export const Back = () => {
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
    <Host style={{ width: 36, height: 36, backgroundColor: "red" }}>
      <Button
        onPress={handleBack}
        systemImage="chevron.left"
        variant="glass"
        modifiers={[
          frame({ width: 36, height: 36 }),
          aspectRatio({ ratio: 1 }),
        ]}
      />
    </Host>
  );
};
