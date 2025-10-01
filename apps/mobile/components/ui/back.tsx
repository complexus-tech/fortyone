import React from "react";
import { TouchableOpacity } from "react-native";
import { useRouter } from "expo-router";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";

export const Back = () => {
  const router = useRouter();
  const canGoBack = router.canGoBack();
  if (!canGoBack) return null;
  return (
    <TouchableOpacity onPress={() => router.back()}>
      <SymbolView
        name="arrow.backward"
        weight="semibold"
        size={20}
        tintColor={colors.dark[50]}
      />
    </TouchableOpacity>
  );
};
