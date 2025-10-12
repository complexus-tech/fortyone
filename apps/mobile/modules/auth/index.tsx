import { View } from "react-native";
import { Button, Col, Text } from "@/components/ui";
import { useAuthStore } from "@/store";

export const Auth = () => {
  const setAuthData = useAuthStore((state) => state.setAuthData);

  const handleOTPSubmit = () => {
    // Mock OTP verification
    setAuthData(
      "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI4YTc5ODExMi05MGZlLTQ5NWUtOWYxYy1mMzY2NTVlM2Q4YWIiLCJleHAiOjE3NjI4ODMyMDMsIm5iZiI6MTc1OTQyNzIwMywiaWF0IjoxNzU5NDI3MjAzfQ.K05W85tEEWQ5dFqu7bgXjjowkk_zYowwKSJ_VMXR7_o",
      "complexus"
    );
  };

  return (
    <View
      style={{
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
        <Button size="lg" rounded="lg" className="w-full">
          Get Started
        </Button>
        <Text fontSize="sm" align="center" className="mt-4">
          © {new Date().getFullYear()} • Product of Complexus LLC • All Rights
          Reserved.
        </Text>
      </Col>
    </View>
  );
};
