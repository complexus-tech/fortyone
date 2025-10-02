import React from "react";
import { TouchableOpacity } from "react-native";
import { useRouter } from "expo-router";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";

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
    <TouchableOpacity onPress={handleBack}>
      <SymbolView
        name="arrow.backward"
        weight="semibold"
        size={20}
        tintColor={colors.dark[50]}
      />
    </TouchableOpacity>
  );
};
