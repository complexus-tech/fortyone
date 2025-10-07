import React, { useEffect, useState } from "react";
import { TextInput } from "react-native";
import { Button } from "@/components/ui";
import { useProfile } from "@/modules/users/hooks/use-profile";
import { useUpdateProfileMutation } from "@/modules/users/hooks/use-update-profile-mutation";

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
    if (!hasChanged) return;

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
        className="border rounded-[0.7rem] border-gray-100 bg-gray-50 px-4 h-14 mx-4 mb-4 dark:border-dark-50 dark:bg-dark-300 dark:text-white"
      />
      <TextInput
        placeholder="Username"
        autoComplete="username"
        value={form.username}
        onChangeText={(text) => setForm({ ...form, username: text })}
        className="border rounded-[0.7rem] border-gray-100 bg-gray-50 px-4 h-14 mx-4 dark:border-dark-50 dark:bg-dark-300 dark:text-white"
      />
      <Button
        onPress={handleUpdateProfile}
        disabled={updateMutation.isPending}
        className="mt-5 mx-4"
      >
        {updateMutation.isPending ? "Saving..." : "Save changes"}
      </Button>
    </>
  );
};
