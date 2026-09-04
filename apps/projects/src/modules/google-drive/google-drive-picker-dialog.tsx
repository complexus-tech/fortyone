"use client";

import { useEffect, useRef } from "react";
import { toast } from "sonner";
import {
  useAttachGoogleDriveFiles,
  useCreateGoogleDriveConnectSession,
  useCreateGoogleDrivePickerSession,
  useGoogleDriveIntegration,
} from "./hooks";
import type {
  GoogleDriveFileReference,
  GoogleDrivePickerFile,
  GoogleDriveTarget,
} from "./types";

const GOOGLE_PICKER_SCRIPT_ID = "google-drive-picker-api";
const GOOGLE_PICKER_SCRIPT_URL = "https://apis.google.com/js/api.js";
const MAX_GOOGLE_DRIVE_SELECTIONS = 20;
const GOOGLE_DRIVE_MIME_TYPES = [
  "application/vnd.google-apps.document",
  "application/vnd.google-apps.spreadsheet",
  "application/vnd.google-apps.presentation",
  "application/pdf",
  "image/jpeg",
  "image/png",
  "image/webp",
].join(",");

type PickerDocument = {
  id: string;
  mimeType?: string;
  name?: string;
  resourceKey?: string;
};

type PickerCallbackData = {
  action?: string;
  docs?: PickerDocument[];
};

type GooglePickerApi = {
  Action: { CANCEL: string; PICKED: string };
  DocsView: new (viewId?: string) => {
    setEnableDrives: (enabled: boolean) => unknown;
    setFileIds: (fileIds: string) => unknown;
    setIncludeFolders: (enabled: boolean) => unknown;
    setMimeTypes: (mimeTypes: string) => unknown;
    setMode: (mode: string) => unknown;
    setSelectFolderEnabled: (enabled: boolean) => unknown;
  };
  DocsViewMode: { LIST: string };
  Feature: {
    MULTISELECT_ENABLED: string;
  };
  PickerBuilder: new () => {
    addView: (view: unknown) => unknown;
    build: () => { setVisible: (visible: boolean) => void };
    enableFeature: (feature: string) => unknown;
    setAppId: (appId: string) => unknown;
    setCallback: (callback: (data: PickerCallbackData) => void) => unknown;
    setDeveloperKey: (apiKey: string) => unknown;
    setMaxItems: (maximum: number) => unknown;
    setOAuthToken: (accessToken: string) => unknown;
    setOrigin: (origin: string) => unknown;
  };
  ViewId: { DOCS: string };
};

declare global {
  interface Window {
    gapi?: {
      load: (
        library: string,
        options: { callback: () => void; onerror: () => void },
      ) => void;
    };
    google?: { picker?: GooglePickerApi };
  }
}

let pickerApiPromise: Promise<void> | null = null;

const loadGooglePickerApi = () => {
  if (window.google?.picker) return Promise.resolve();
  if (pickerApiPromise) return pickerApiPromise;

  pickerApiPromise = new Promise<void>((resolve, reject) => {
    const loadPickerLibrary = () => {
      if (!window.gapi) {
        reject(new Error("Google Picker did not finish loading."));
        return;
      }
      window.gapi.load("picker", {
        callback: resolve,
        onerror: () => {
          reject(new Error("Google Picker could not be loaded."));
        },
      });
    };

    const existingScript = document.getElementById(
      GOOGLE_PICKER_SCRIPT_ID,
    ) as HTMLScriptElement | null;
    if (existingScript) {
      if (window.gapi) loadPickerLibrary();
      else
        existingScript.addEventListener("load", loadPickerLibrary, {
          once: true,
        });
      return;
    }

    const script = document.createElement("script");
    script.async = true;
    script.defer = true;
    script.id = GOOGLE_PICKER_SCRIPT_ID;
    script.src = GOOGLE_PICKER_SCRIPT_URL;
    script.addEventListener("load", loadPickerLibrary, { once: true });
    script.addEventListener(
      "error",
      () => {
        script.remove();
        reject(new Error("Google Picker could not be loaded."));
      },
      { once: true },
    );
    document.head.append(script);
  }).catch((error) => {
    pickerApiPromise = null;
    throw error;
  });

  return pickerApiPromise;
};

const mapPickerFiles = (documents: PickerDocument[]): GoogleDrivePickerFile[] =>
  documents.map(({ id, mimeType, name, resourceKey }) => ({
    id,
    ...(mimeType ? { mimeType } : {}),
    ...(name ? { name } : {}),
    ...(resourceKey ? { resourceKey } : {}),
  }));

export type GoogleDrivePickerDialogProps = {
  fileIds?: string[];
  onAttached?: (files: GoogleDriveFileReference[]) => void;
  onClose: () => void;
  target: GoogleDriveTarget;
};

export const GoogleDrivePickerDialog = ({
  fileIds,
  onAttached,
  onClose,
  target,
}: GoogleDrivePickerDialogProps) => {
  const integration = useGoogleDriveIntegration();
  const connect = useCreateGoogleDriveConnectSession();
  const pickerSession = useCreateGoogleDrivePickerSession();
  const attachFiles = useAttachGoogleDriveFiles(target);
  const connectRef = useRef(connect.mutate);
  const createSessionRef = useRef(pickerSession.mutateAsync);
  const attachFilesRef = useRef(attachFiles.mutateAsync);
  const onAttachedRef = useRef(onAttached);
  const onCloseRef = useRef(onClose);
  const focusedFileIds = fileIds?.filter(Boolean) ?? [];
  const focusedFileIdsRef = useRef(focusedFileIds);
  let pickerAvailability:
    | "connection_required"
    | "pending"
    | "ready"
    | "unavailable" = "unavailable";
  if (integration.isPending) {
    pickerAvailability = "pending";
  } else if (
    integration.data?.configured &&
    integration.data.connected &&
    !integration.data.requiresReauthorization
  ) {
    pickerAvailability = "ready";
  } else if (integration.data?.configured) {
    pickerAvailability = "connection_required";
  }
  const pickerConnectionLabel = integration.data?.connected
    ? "Reconnect"
    : "Connect";

  connectRef.current = connect.mutate;
  createSessionRef.current = pickerSession.mutateAsync;
  attachFilesRef.current = attachFiles.mutateAsync;
  onAttachedRef.current = onAttached;
  onCloseRef.current = onClose;
  focusedFileIdsRef.current = focusedFileIds;

  useEffect(() => {
    if (pickerAvailability === "pending") return;
    if (pickerAvailability !== "ready") {
      if (pickerAvailability === "connection_required") {
        toast.info("Connect Google Drive to continue", {
          action: {
            label: pickerConnectionLabel,
            onClick: (event) => {
              event.preventDefault();
              connectRef.current(window.location.href);
            },
          },
          description:
            "Your link is saved. Connect your Google account to add its private preview.",
        });
      } else {
        toast.error("Google Drive", {
          description: "Google Drive is not configured for this environment.",
        });
      }
      onCloseRef.current();
      return;
    }

    let active = true;
    let picker: { setVisible: (visible: boolean) => void } | null = null;

    const openPicker = async () => {
      try {
        const [session] = await Promise.all([
          createSessionRef.current(),
          loadGooglePickerApi(),
        ]);
        if (!active) return;
        const pickerApi = window.google?.picker;
        if (!pickerApi) throw new Error("Google Picker is unavailable.");

        const view = new pickerApi.DocsView(pickerApi.ViewId.DOCS);
        const requestedFileIds = focusedFileIdsRef.current;
        const isFocusedPicker = requestedFileIds.length > 0;
        view.setIncludeFolders(!isFocusedPicker);
        view.setSelectFolderEnabled(false);
        if (isFocusedPicker) {
          view.setFileIds(requestedFileIds.join(","));
        } else {
          view.setEnableDrives(true);
        }
        view.setMimeTypes(GOOGLE_DRIVE_MIME_TYPES);
        view.setMode(pickerApi.DocsViewMode.LIST);

        const builder = new pickerApi.PickerBuilder();
        builder.addView(view);
        if (!isFocusedPicker) {
          builder.enableFeature(pickerApi.Feature.MULTISELECT_ENABLED);
        }
        builder.setAppId(session.appId);
        builder.setDeveloperKey(session.apiKey);
        builder.setMaxItems(isFocusedPicker ? 1 : MAX_GOOGLE_DRIVE_SELECTIONS);
        builder.setOAuthToken(session.accessToken);
        builder.setOrigin(session.origin ?? window.location.origin);
        builder.setCallback((data) => {
          if (!active) return;
          if (data.action === pickerApi.Action.CANCEL) {
            onCloseRef.current();
            return;
          }
          if (data.action !== pickerApi.Action.PICKED || !data.docs?.length) {
            return;
          }
          if (
            isFocusedPicker &&
            (data.docs.length !== 1 ||
              !requestedFileIds.includes(data.docs[0]?.id ?? ""))
          ) {
            toast.error("Google Drive", {
              description: "Choose the Google file from the pasted link.",
            });
            onCloseRef.current();
            return;
          }
          void attachFilesRef.current(mapPickerFiles(data.docs)).then(
            (files) => {
              if (!active) return;
              onAttachedRef.current?.(files);
              onCloseRef.current();
            },
            () => {
              if (!active) return;
              // The mutation displays the provider error and leaves the picker
              // lifecycle in a known closed state.
              onCloseRef.current();
            },
          );
        });
        picker = builder.build();
        picker.setVisible(true);
      } catch (error) {
        if (!active) return;
        toast.error("Google Drive", {
          description:
            error instanceof Error
              ? error.message
              : "Google Picker could not be opened.",
        });
        onCloseRef.current();
      }
    };

    void openPicker();
    return () => {
      active = false;
      picker?.setVisible(false);
    };
  }, [pickerAvailability, pickerConnectionLabel]);

  return null;
};
