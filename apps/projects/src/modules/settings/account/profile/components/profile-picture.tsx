"use client";

import { Avatar } from "ui";
import { useProfile } from "@/lib/hooks/profile";

export const ProfilePicture = () => {
  const { data: profile } = useProfile();
  return (
    <Avatar
      className="h-10"
      name={profile?.fullName || profile?.username}
      src={profile?.avatarUrl}
    />
  );
};
