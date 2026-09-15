import { useRouter } from "expo-router";
import { IconButton } from "./icon-button";

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

  return <IconButton icon="chevron-back" label="Back" onPress={handleBack} />;
};
