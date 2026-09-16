import { useRouter } from "expo-router";
import { Avatar, ScreenHeader } from "@/components/ui";
import { GlassIconButton } from "@/components/ui/glass-icon-button";
import { useProfile } from "@/modules/users/hooks/use-profile";
import { NewStoryButton } from "./new-story";

export const Header = () => {
  const router = useRouter();
  const { data: profile } = useProfile();

  return (
    <ScreenHeader
      title="Home"
      trailing={
        <>
          <NewStoryButton />
          <GlassIconButton
            label="Open settings"
            onPress={() => router.push("/settings")}
          >
            <Avatar
              name={profile?.fullName || profile?.username}
              src={profile?.avatarUrl}
              style={{ width: 30, height: 30 }}
            />
          </GlassIconButton>
        </>
      }
    />
  );
};
