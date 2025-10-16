import React, { useEffect, useState } from "react";
import { TextInput } from "react-native";
import { Button } from "@/components/ui";
import { useProfile } from "@/modules/users/hooks/use-profile";
import { useUpdateProfileMutation } from "@/modules/users/hooks/use-update-profile-mutation";
import { toast } from "sonner-native";

export const ProfileForm = () => {
  const { data: profile } = useProfile();
  const updateMutation = useUpdateProfileMutation();

  const [form, setForm] = useState({
    fullName: "",
    username: "",
  });

  const [hasChanged, setHasChanged] = useState(false);

  useEffect(() => {
    if (profile) {
      setForm({
        fullName: profile.fullName || "",
        username: profile.username || "",
      });
    }
  }, [profile]);

  useEffect(() => {
    const changed =
      form.fullName !== (profile?.fullName || "") ||
      form.username !== (profile?.username || "");
    setHasChanged(changed);
  }, [form, profile]);

  const handleUpdateProfile = () => {
    if (!hasChanged) {
      toast.warning("No changes to save");
      return;
    }

    updateMutation.mutate({
      fullName: form.fullName,
      username: form.username,
    });
  };

  if (!profile) {
    return null;
  }

  return (
    <>
      <TextInput
        placeholder="Full name"
        autoComplete="name"
        value={form.fullName}
        onChangeText={(text) => setForm({ ...form, fullName: text })}
        className="rounded-[12px] text-[16px] bg-gray-100/70 px-4 h-12 font-medium mx-4.5 dark:bg-dark-100 dark:text-white mb-4"
      />
      <TextInput
        placeholder="Username"
        autoComplete="username"
        value={form.username}
        onChangeText={(text) => setForm({ ...form, username: text })}
        className="rounded-[12px] text-[16px] bg-gray-100/70 px-4 h-12 font-medium mx-4.5 dark:bg-dark-100 dark:text-white"
      />
      <Button
        onPress={handleUpdateProfile}
        disabled={updateMutation.isPending}
        className="mt-5 mx-4.5 rounded-[12px]"
        color="invert"
      >
        {updateMutation.isPending ? "Saving..." : "Save changes"}
      </Button>
    </>
  );
};
