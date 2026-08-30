"use client";

import { useState } from "react";
import { ArrowLeft2Icon } from "icons";
import { Box, Flex, Input, Switch, Text } from "ui";
import type { PublicPortal } from "@/shared/feedback-widget/types";
import {
  confirmWidgetFeedbackVerificationAction,
  requestWidgetFeedbackVerificationAction,
  type WidgetParticipantSession,
} from "../actions";
import { WidgetBackButton } from "./widget-ui";

export const IdentityGate = ({
  onBack,
  onVerified,
  portal,
}: {
  onBack: () => void;
  onVerified: (session: WidgetParticipantSession) => void;
  portal: PublicPortal;
}) => {
  const [displayName, setDisplayName] = useState("");
  const [email, setEmail] = useState("");
  const [code, setCode] = useState("");
  const [hideNamePublicly, setHideNamePublicly] = useState(
    portal.guestIdentityPolicy === "always_mask_guests",
  );
  const [step, setStep] = useState<"details" | "verify">("details");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState("");

  const requestCode = async () => {
    if (!email.trim() || isSubmitting) return;
    setIsSubmitting(true);
    setError("");
    const response = await requestWidgetFeedbackVerificationAction({
      displayName,
      email,
      hideNamePublicly,
      portalSlug: portal.slug,
    })
      .catch(() => null)
      .finally(() => {
        setIsSubmitting(false);
      });
    if (!response) {
      setError("Unable to send a verification code");
      return;
    }
    if (response.error?.message || !response.data) {
      setError(response.error?.message ?? "Unable to send a verification code");
      return;
    }
    setStep("verify");
  };

  const confirmCode = async () => {
    if (!code.trim() || isSubmitting) return;
    setIsSubmitting(true);
    setError("");
    const response = await confirmWidgetFeedbackVerificationAction({
      code,
      email,
      portalSlug: portal.slug,
    })
      .catch(() => null)
      .finally(() => {
        setIsSubmitting(false);
      });
    if (!response) {
      setError("That code could not be verified");
      return;
    }
    if (response.error?.message || !response.data) {
      setError(response.error?.message ?? "That code could not be verified");
      return;
    }
    onVerified(response.data);
  };

  const submitLabel =
    step === "details" ? "Send verification code" : "Verify and continue";
  const actionLabel = isSubmitting ? "Please wait…" : submitLabel;

  return (
    <Box className="bg-background absolute inset-0 z-40 flex min-h-0 flex-col">
      <Flex align="center" className="h-16 shrink-0 px-4" gap={2}>
        <WidgetBackButton aria-label="Go back" onClick={onBack}>
          <ArrowLeft2Icon className="h-5" />
        </WidgetBackButton>
        <Text className="text-[16px]" fontWeight="semibold">
          {step === "details" ? "Join the conversation" : "Check your email"}
        </Text>
      </Flex>
      <Box className="min-h-0 flex-1 overflow-y-auto px-6 pt-5">
        {step === "details" ? (
          <>
            <Text className="text-[19px] leading-7" fontWeight="semibold">
              Verify once, then stay in the widget
            </Text>
            <Text className="mt-2 text-[12px] leading-5" color="muted">
              We’ll send a short code so you can vote, comment, and receive
              relevant feedback updates without creating a FortyOne account.
            </Text>
            <Box className="mt-7 space-y-4">
              <label className="block space-y-2" htmlFor="feedback-widget-name">
                <Text
                  as="span"
                  className="block text-[11px]"
                  color="muted"
                  fontWeight="semibold"
                >
                  Name
                </Text>
                <Input
                  id="feedback-widget-name"
                  onChange={(event) => {
                    setDisplayName(event.target.value);
                  }}
                  placeholder="Your name"
                  value={displayName}
                />
              </label>
              <label
                className="block space-y-2"
                htmlFor="feedback-widget-email"
              >
                <Text
                  as="span"
                  className="block text-[11px]"
                  color="muted"
                  fontWeight="semibold"
                >
                  Email
                </Text>
                <Input
                  autoFocus
                  id="feedback-widget-email"
                  onChange={(event) => {
                    setEmail(event.target.value);
                  }}
                  placeholder="you@company.com"
                  type="email"
                  value={email}
                />
              </label>
              {portal.guestIdentityPolicy === "allow_public_masking" ? (
                <Flex
                  align="center"
                  className="border-border/70 rounded-lg border p-3"
                  justify="between"
                >
                  <Box className="pr-4">
                    <Text className="text-[12px]" fontWeight="medium">
                      Post as Anonymous
                    </Text>
                    <Text
                      className="mt-0.5 text-[10px] leading-4"
                      color="muted"
                    >
                      Your email stays private either way.
                    </Text>
                  </Box>
                  <Switch
                    checked={hideNamePublicly}
                    onCheckedChange={setHideNamePublicly}
                  />
                </Flex>
              ) : null}
            </Box>
          </>
        ) : (
          <>
            <Text className="text-[19px] leading-7" fontWeight="semibold">
              Enter your verification code
            </Text>
            <Text className="mt-2 text-[12px] leading-5" color="muted">
              We sent a code to {email}. It may take a moment to arrive.
            </Text>
            <Input
              aria-label="Verification code"
              autoComplete="one-time-code"
              autoFocus
              className="mt-7 text-center tracking-[0.28em]"
              inputMode="numeric"
              maxLength={8}
              onChange={(event) => {
                setCode(event.target.value);
              }}
              placeholder="000000"
              value={code}
            />
            <button
              className="text-text-muted hover:text-foreground mt-4 text-[11px] transition-colors"
              onClick={() => {
                setCode("");
                setError("");
                setStep("details");
              }}
              type="button"
            >
              Use a different email
            </button>
          </>
        )}
        {error ? (
          <Text className="mt-4 text-[12px] text-red-600 dark:text-red-400">
            {error}
          </Text>
        ) : null}
      </Box>
      <Box className="shrink-0 px-6 py-5">
        <button
          className="bg-foreground text-background focus-visible:ring-ring inline-flex h-10 w-full items-center justify-center rounded-lg px-5 text-[13px] font-semibold focus-visible:ring-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-40"
          disabled={
            step === "details"
              ? !email.trim() || isSubmitting
              : !code.trim() || isSubmitting
          }
          onClick={() =>
            void (step === "details" ? requestCode() : confirmCode())
          }
          type="button"
        >
          {actionLabel}
        </button>
      </Box>
    </Box>
  );
};
