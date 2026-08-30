"use client";

import type { FormEvent } from "react";
import { useState } from "react";
import { Box, Button, Input } from "ui";
import { SectionHeader } from "@/components/ui/section-header";
import { useProfile } from "@/lib/hooks/profile";
import { useUpdateProfileMutation } from "@/lib/hooks/update-profile-mutation";
import type { User } from "@/types/user";
import { ProfilePicture } from "./profile-picture";

type ProfileFormProps = {
  initialProfile?: User;
};

export const ProfileForm = ({ initialProfile }: ProfileFormProps) => {
  const { data: profile } = useProfile(initialProfile);
  const { isPending, mutate: updateProfile } = useUpdateProfileMutation();
  const profileFullName = profile?.fullName ?? "";
  const profileUsername = profile?.username ?? "";
  const [draft, setDraft] = useState<{
    fullName?: string;
    username?: string;
  }>({});
  const form = {
    fullName: draft.fullName ?? profileFullName,
    username: draft.username ?? profileUsername,
  };
  const hasChanged =
    form.fullName !== profileFullName || form.username !== profileUsername;

  const handleUpdateProfile = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const submittedForm = form;
    updateProfile(
      {
        fullName: submittedForm.fullName.trim(),
        username: submittedForm.username.trim(),
      },
      {
        onSuccess: (response) => {
          if (!response.error?.message) {
            setDraft((current) => ({
              ...(current.fullName !== undefined &&
              current.fullName !== submittedForm.fullName
                ? { fullName: current.fullName }
                : {}),
              ...(current.username !== undefined &&
              current.username !== submittedForm.username
                ? { username: current.username }
                : {}),
            }));
          }
        },
      },
    );
  };

  return (
    <Box className="border-border bg-surface rounded-2xl border">
      <SectionHeader
        action={<ProfilePicture initialProfile={profile ?? initialProfile} />}
        description="Update your personal information and profile picture."
        title="Personal Information"
      />
      <form className="p-6" onSubmit={handleUpdateProfile}>
        <Box className="mb-2 grid gap-4 md:grid-cols-2 md:gap-6">
          <Input
            label="Full name"
            name="fullName"
            onChange={(event) => {
              setDraft((current) => ({
                ...current,
                fullName: event.target.value,
              }));
            }}
            placeholder="Enter your full name"
            value={form.fullName}
          />
          <Input
            label="Username"
            name="username"
            onChange={(event) => {
              setDraft((current) => ({
                ...current,
                username: event.target.value,
              }));
            }}
            placeholder="Enter your username"
            value={form.username}
          />
        </Box>
        <Button
          className="mt-4"
          disabled={!hasChanged || isPending}
          loading={isPending}
          loadingText="Saving..."
          type="submit"
        >
          Save changes
        </Button>
      </form>
    </Box>
  );
};
