import { View } from "react-native";
import { Button } from "@/components/ui/button";
import { Text } from "@/components/ui/text";
import { useAuthStore } from "@/store";
import { SafeContainer } from "@/components/ui";

export default function Login() {
  const setAuthData = useAuthStore((state) => state.setAuthData);

  const handleLogin = () => {
    // Mock login - replace with real auth later
    setAuthData(
      "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI4YTc5ODExMi05MGZlLTQ5NWUtOWYxYy1mMzY2NTVlM2Q4YWIiLCJleHAiOjE3NjI4ODMyMDMsIm5iZiI6MTc1OTQyNzIwMywiaWF0IjoxNzU5NDI3MjAzfQ.K05W85tEEWQ5dFqu7bgXjjowkk_zYowwKSJ_VMXR7_o",
      "complexus"
    );
  };

  return (
    <SafeContainer className="flex-1 justify-center items-center px-6">
      <View className="w-full max-w-sm">
        <Text fontSize="3xl" fontWeight="bold" align="center" className="mb-8">
          Welcome to Complexus
        </Text>

        <Text fontSize="md" color="muted" align="center" className="mb-8">
          Sign in to access your workspace
        </Text>

        <Button size="lg" rounded="lg" onPress={handleLogin} className="w-full">
          Login
        </Button>
      </View>
    </SafeContainer>
  );
}
