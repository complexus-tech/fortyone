"use client";

import { Avatar } from "ui";
import { ProfileUploadDialog } from "ui/src/ProfileUploadDialog/ProfileUploadDialog";
import { useState } from "react";
import { useProfile } from "@/lib/hooks/profile";

export const ProfilePicture = () => {
  const { data: profile } = useProfile();

  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [profileImage, setProfileImage] = useState<string | undefined>();

  const handleUpload = (file: File) => {
    // In a real app, you would upload the file to your server
    // For this example, we'll just create a local URL
    const imageUrl = URL.createObjectURL(file);
    setProfileImage(imageUrl);
    // console.log("Uploading file:", file.name, file.size, file.type);
  };

  const handleRemove = () => {
    if (profileImage) {
      URL.revokeObjectURL(profileImage);
      setProfileImage(undefined);
    }
    // console.log("Removing profile image");
  };

  return (
    <>
      <button
        onClick={() => {
          setIsDialogOpen(true);
        }}
        type="button"
      >
        <Avatar
          className="h-10"
          name={profile?.fullName || profile?.username}
          src={profile?.avatarUrl}
        />
        <span className="sr-only">Change profile picture</span>
      </button>
      <ProfileUploadDialog
        currentImage={profile?.avatarUrl}
        isOpen={isDialogOpen}
        maxSizeInMB={5}
        onOpenChange={setIsDialogOpen}
        onRemove={handleRemove}
        onUpload={handleUpload}
      />
    </>
  );
};
