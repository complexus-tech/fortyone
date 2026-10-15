"use client";

import { useEffect, useState } from "react";
import { Box, Button, Dialog, Flex, Text } from "ui";
import { getDesktopBridge } from "./bridge";
import { useDesktop } from "./desktop-provider";

type UpdateStatus = { configured: boolean; currentVersion: string };
const message = (reason: unknown) =>
  reason instanceof Error ? reason.message : String(reason);

export function DesktopUpdateSettings() {
  const { isDesktop } = useDesktop();
  const bridge = getDesktopBridge();
  const supported = Boolean(
    isDesktop &&
      bridge?.getUpdateStatus &&
      bridge.checkForUpdate &&
      bridge.installUpdate,
  );
  const [status, setStatus] = useState<UpdateStatus | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [checking, setChecking] = useState(false);
  const [availableVersion, setAvailableVersion] = useState<string | null>(null);
  const [checked, setChecked] = useState(false);
  const [open, setOpen] = useState(false);
  const [installing, setInstalling] = useState(false);
  const [progress, setProgress] = useState(
    "Downloading and verifying the update…",
  );
  const [restarting, setRestarting] = useState(false);
  useEffect(() => {
    if (!supported) return;
    let active = true;
    void bridge!.getUpdateStatus!().then(
      (value) => {
        if (active) setStatus(value);
      },
      (reason: unknown) => {
        if (active) setError(message(reason));
      },
    );
    return () => {
      active = false;
    };
  }, [bridge, supported]);
  useEffect(() => {
    if (!installing) return;
    const onProgress = (event: Event) => {
      const detail = (
        event as CustomEvent<{
          phase: string;
          received: number;
          total: number | null;
        }>
      ).detail;
      if (detail.phase === "saving") setProgress("Saving your local work…");
      else if (detail.phase === "installing")
        setProgress("Installing the verified update…");
      else if (detail.phase === "downloading") {
        const percent =
          detail.total && detail.total > 0
            ? Math.min(
                100,
                Math.max(0, Math.round((detail.received / detail.total) * 100)),
              )
            : null;
        setProgress(
          percent === null
            ? "Downloading and verifying the update…"
            : `Downloading update… ${percent}%`,
        );
      }
    };
    window.addEventListener("fortyone:desktop-update-progress", onProgress);
    return () => {
      window.removeEventListener(
        "fortyone:desktop-update-progress",
        onProgress,
      );
    };
  }, [installing]);
  if (!supported) return null;

  const check = async () => {
    setChecking(true);
    setError(null);
    try {
      const current = await bridge!.getUpdateStatus!();
      setStatus(current);
      if (!current.configured) return;
      const result = await bridge!.checkForUpdate!();
      setAvailableVersion(
        result.available && result.version ? result.version : null,
      );
      setChecked(true);
    } catch (reason) {
      setError(message(reason));
    } finally {
      setChecking(false);
    }
  };
  const install = async () => {
    if (!availableVersion || installing) return;
    setInstalling(true);
    setError(null);
    try {
      await bridge!.installUpdate!({ version: availableVersion });
      setRestarting(true);
    } catch (reason) {
      setInstalling(false);
      setError(message(reason));
    }
  };

  let installMessage =
    "FortyOne will save your local drafts, install the signed update, and restart. Keep this window open until it finishes.";
  if (installing) installMessage = progress;
  if (restarting) installMessage = "Restarting FortyOne…";
  return (
    <Box className="border-border border-t pt-6">
      <Flex align="center" gap={4} justify="between">
        <Box>
          <Text fontWeight="medium">App updates</Text>
          <Text color="muted">
            {status
              ? `Version ${status.currentVersion}`
              : "Reading app version…"}
          </Text>
        </Box>
        {status?.configured || error ? (
          <Button
            disabled={checking || installing}
            onClick={() => {
              void check();
            }}
            variant="outline"
          >
            {checking ? "Checking…" : "Check for updates"}
          </Button>
        ) : null}
      </Flex>
      {status && !status.configured ? (
        <Text className="mt-2" color="muted">
          Automatic updates are available in a signed release build. This
          development build updates through the local build tools.
        </Text>
      ) : null}
      {checked && !availableVersion ? (
        <Text className="mt-2" color="muted" role="status">
          You’re up to date.
        </Text>
      ) : null}
      {availableVersion ? (
        <Button
          className="mt-3"
          onClick={() => {
            setError(null);
            setOpen(true);
          }}
        >
          Review version {availableVersion}
        </Button>
      ) : null}
      {error && !open ? (
        <Text className="mt-2" color="danger" role="alert">
          {error}
        </Text>
      ) : null}
      <Dialog
        onOpenChange={(next) => {
          if (!installing) setOpen(next);
        }}
        open={open}
      >
        <Dialog.Content
          hideClose={installing}
          onEscapeKeyDown={(event) => {
            if (installing) event.preventDefault();
          }}
          onInteractOutside={(event) => {
            if (installing) event.preventDefault();
          }}
          size="sm"
        >
          <Dialog.Header>
            <Dialog.Title>Update FortyOne</Dialog.Title>
            <Dialog.Description>Version {availableVersion}</Dialog.Description>
          </Dialog.Header>
          <Dialog.Body>
            <Text role={installing ? "status" : undefined}>
              {installMessage}
            </Text>
            {error ? (
              <Text className="mt-3" color="danger" role="alert">
                {error}
              </Text>
            ) : null}
          </Dialog.Body>
          <Dialog.Footer className="justify-end gap-2">
            <Button
              color="tertiary"
              disabled={installing}
              onClick={() => {
                setOpen(false);
              }}
            >
              Cancel
            </Button>
            <Button
              disabled={installing}
              onClick={() => {
                void install();
              }}
            >
              {installing ? "Installing…" : "Install and restart"}
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </Box>
  );
}
