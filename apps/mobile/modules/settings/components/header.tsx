import { useRouter } from "expo-router";
import { ScreenHeader } from "@/components/ui";
import { GlassIconButton } from "@/components/ui/glass-icon-button";

export const Header = ({ onClose }: { onClose?: () => void }) => {
  const router = useRouter();

  const close = () => {
    if (onClose) {
      onClose();
      return;
    }
    if (router.canGoBack()) router.back();
    else router.replace("/");
  };

  return (
    <ScreenHeader
      title="Settings"
      compact
      trailing={
        <GlassIconButton
          icon="close"
          systemImage="xmark"
          label="Close settings"
          onPress={close}
        />
      }
    />
  );
};
