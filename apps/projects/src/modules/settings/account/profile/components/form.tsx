import type { FormEvent } from "react";
import React, { useEffect, useState } from "react";
import { Box, Input, Button } from "ui";
import { SectionHeader } from "@/modules/settings/components";
import { useProfile } from "@/lib/hooks/profile";
import { useUpdateProfileMutation } from "@/lib/hooks/update-profile-mutation";
import { ProfilePicture } from "./profile-picture";

export const Form = () => {
  const { data: profile } = useProfile();
  const { mutate: updateProfile } = useUpdateProfileMutation();

  const [form, setForm] = useState({
    fullName: profile?.fullName || "",
    username: profile?.username || "",
  });

  const hasChanged = () => {
    return (
      form.fullName !== profile?.fullName || form.username !== profile.username
    );
  };

  const handleUpdateProfile = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    updateProfile(form);
  };

  useEffect(() => {
    setForm({
      fullName: profile?.fullName || "",
      username: profile?.username || "",
    });
  }, [profile]);

  return (
    <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
      <SectionHeader
        action={<ProfilePicture />}
        description="Update your personal information and profile picture."
        title="Personal Information"
      />
      <form className="p-6" onSubmit={handleUpdateProfile}>
        <Box className="mb-4 grid gap-4 md:grid-cols-2 md:gap-6">
          <Input
            label="Full name"
            name="fullName"
            onChange={(e) => {
              setForm({ ...form, fullName: e.target.value });
            }}
            placeholder="Enter your full name"
            value={form.fullName}
          />
          <Input
            label="Username"
            name="username"
            onChange={(e) => {
              setForm({ ...form, username: e.target.value });
            }}
            placeholder="Enter your username"
            value={form.username}
          />
        </Box>
        <Button disabled={!hasChanged()} type="submit">
          Save changes
        </Button>
      </form>
    </Box>
  );
};
