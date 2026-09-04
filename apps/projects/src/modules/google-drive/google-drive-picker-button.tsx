"use client";

import type { ComponentPropsWithoutRef, ReactNode } from "react";
import { useState } from "react";
import dynamic from "next/dynamic";
import { Button } from "ui";
import { toast } from "sonner";
import {
  useCreateGoogleDriveConnectSession,
  useGoogleDriveIntegration,
} from "./hooks";
import type { GoogleDriveFileReference, GoogleDriveTarget } from "./types";

const GoogleDrivePickerDialog = dynamic(
  () =>
    import("./google-drive-picker-dialog").then(
      (module) => module.GoogleDrivePickerDialog,
    ),
  { ssr: false },
);

type ButtonProps = ComponentPropsWithoutRef<typeof Button>;

export type GoogleDrivePickerButtonProps = Omit<
  ButtonProps,
  "onClick" | "target"
> & {
  children?: ReactNode;
  fileIds?: string[];
  onAttached?: (files: GoogleDriveFileReference[]) => void;
  target: GoogleDriveTarget;
};

export const GoogleDrivePickerButton = ({
  children = "Attach from Drive",
  disabled,
  fileIds,
  loading,
  onAttached,
  target,
  ...buttonProps
}: GoogleDrivePickerButtonProps) => {
  const [isPickerOpen, setIsPickerOpen] = useState(false);
  const integration = useGoogleDriveIntegration();
  const connect = useCreateGoogleDriveConnectSession();

  const openPicker = () => {
    if (!integration.data) {
      toast.error("Google Drive", {
        description: "The Google Drive connection could not be checked.",
      });
      return;
    }
    if (!integration.data.configured) {
      toast.error("Google Drive", {
        description: "Google Drive is not configured for this environment.",
      });
      return;
    }
    if (
      !integration.data.connected ||
      integration.data.requiresReauthorization
    ) {
      connect.mutate(window.location.href);
      return;
    }
    setIsPickerOpen(true);
  };

  return (
    <>
      <Button
        {...buttonProps}
        disabled={disabled || integration.isPending}
        loading={loading || connect.isPending}
        onClick={openPicker}
      >
        {children}
      </Button>
      {isPickerOpen ? (
        <GoogleDrivePickerDialog
          fileIds={fileIds}
          onAttached={onAttached}
          onClose={() => {
            setIsPickerOpen(false);
          }}
          target={target}
        />
      ) : null}
    </>
  );
};
