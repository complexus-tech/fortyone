import { useState } from "react";
import { View, TouchableOpacity, TextInput, ScrollView } from "react-native";
import { Button } from "@/components/ui/button";
import { Text } from "@/components/ui/text";
import { Input } from "@/components/ui/Input";
import { useAuthStore } from "@/store";
import { Col, Row, SafeContainer } from "@/components/ui";

export const Auth = () => {
  const setAuthData = useAuthStore((state) => state.setAuthData);
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);

  const handleEmailSubmit = () => {
    // Mock magic link sending
    setLoading(true);
    setTimeout(() => {
      setLoading(false);
    }, 1000);
  };

  const handleOTPSubmit = () => {
    // Mock OTP verification
    setAuthData(
      "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI4YTc5ODExMi05MGZlLTQ5NWUtOWYxYy1mMzY2NTVlM2Q4YWIiLCJleHAiOjE3NjI4ODMyMDMsIm5iZiI6MTc1OTQyNzIwMywiaWF0IjoxNzU5NDI3MjAzfQ.K05W85tEEWQ5dFqu7bgXjjowkk_zYowwKSJ_VMXR7_o",
      "complexus"
    );
  };

  return (
    <ScrollView
      contentContainerStyle={{
        justifyContent: "space-between",
        flex: 1,
        paddingBottom: 60,
        paddingHorizontal: 16,
      }}
    >
      <View className="size-14 bg-primary rounded-lg mb-8" />
      <Col>
        <Text
          className="mb-6 uppercase text-[14px] tracking-wider"
          fontSize="sm"
        >
          [Built for builders]
        </Text>
        <Text fontSize="4xl" fontWeight="semibold" className="mb-4">
          Plan, track, deliver with the project management tool your team will
          love.
        </Text>
      </Col>
      <Col>
        <Button
          size="lg"
          rounded="lg"
          className="w-full"
          onPress={handleEmailSubmit}
        >
          Get Started
        </Button>
        <Text fontSize="sm" align="center" className="mt-4">
          © {new Date().getFullYear()} • Product of Complexus LLC • All Rights
          Reserved.
        </Text>
      </Col>
    </ScrollView>
  );
};
