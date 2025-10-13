import React from "react";
import { TouchableOpacity } from "react-native";
import { useRouter } from "expo-router";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";
import { useColorScheme } from "nativewind";

export const Back = () => {
  const { colorScheme } = useColorScheme();
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
        name="chevron.backward"
        weight="medium"
        size={22}
        tintColor={colorScheme === "light" ? colors.dark[50] : colors.gray[200]}
      />
    </TouchableOpacity>
  );
};
