import React, { useEffect, useState } from "react";
import { View } from "react-native";
import { Input, Button, Text, Col } from "@/components/ui";
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
    <Col asContainer>
      <Input
        label="Full name"
        placeholder="Enter your full name"
        value={form.fullName}
        onChangeText={(text) => setForm({ ...form, fullName: text })}
      />
      <Input
        label="Username"
        placeholder="Enter your username"
        value={form.username}
        onChangeText={(text) => setForm({ ...form, username: text })}
      />

      <Button
        onPress={handleUpdateProfile}
        disabled={updateMutation.isPending}
        className="mt-6"
      >
        {updateMutation.isPending ? "Saving..." : "Save changes"}
      </Button>
    </Col>
  );
};
