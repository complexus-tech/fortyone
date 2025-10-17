import React from "react";
import { Wrapper, Col, Text } from "@/components/ui";
import { SymbolView, SFSymbol } from "expo-symbols";
import { useRouter } from "expo-router";
import { Pressable } from "react-native";

type StatCardProps = {
  count?: number;
  label: string;
  icon: SFSymbol;
  iconColor?: string;
};

export const StatCard = ({ count, label, icon, iconColor }: StatCardProps) => {
  const router = useRouter();
  return (
    <Pressable onPress={() => router.push("/my-work")}>
      <Wrapper>
        <Col gap={3}>
          <Text fontSize="2xl" fontWeight="semibold">
            {count || 0}
          </Text>
          <Text color="muted" className="opacity-80">
            {label}
          </Text>
        </Col>
        <SymbolView
          name={icon}
          size={20}
          tintColor={iconColor}
          style={{ position: "absolute", top: 14, right: 16 }}
        />
      </Wrapper>
    </Pressable>
  );
};
