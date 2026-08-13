"use client";

import { useId, useState } from "react";
import { Box, Button, Checkbox, Dialog, Flex, Input, Text } from "ui";
import { toast } from "sonner";
import { OTPInput } from "@/components/ui/otp-input";
import {
  confirmFeedbackVerificationAction,
  requestFeedbackVerificationAction,
} from "./actions";
import type { PublicPortal, PublicPortalGuestParticipant } from "./types";

type VerificationStage = "identity" | "code";

const getIdentityPolicyCopy = (portal: PublicPortal) => {
  if (portal.guestIdentityPolicy === "always_mask_guests") {
    return "Your email stays private and your public name will be Anonymous.";
  }
  if (portal.guestIdentityPolicy === "allow_public_masking") {
    return "Your email stays private. You can also hide your name publicly.";
  }
  return "Your email stays private. Your display name will appear publicly.";
};

export const FeedbackGuestVerification = ({
  onBack,
  onVerified,
  portal,
  purpose = "continue",
}: {
  onBack?: () => void;
  onVerified: (participant: PublicPortalGuestParticipant) => void;
  portal: PublicPortal;
  purpose?: string;
}) => {
  const maskIdentityId = useId();
  const [stage, setStage] = useState<VerificationStage>("identity");
  const [email, setEmail] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [hideNamePublicly, setHideNamePublicly] = useState(false);
  const [code, setCode] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const requestVerification = async () => {
    if (!email.trim()) return;
    setIsSubmitting(true);
    try {
      const response = await requestFeedbackVerificationAction({
        portalSlug: portal.slug,
        email,
        displayName,
        ...(portal.guestIdentityPolicy === "allow_public_masking"
          ? { hideNamePublicly }
          : {}),
      });
      if (response.error?.message) {
        toast.error("Unable to send verification email", {
          description: response.error.message,
        });
        return;
      }
      setCode("");
      setStage("code");
    } catch {
      toast.error("Unable to send verification email", {
        description: "Check your connection and try again.",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const confirmVerification = async () => {
    const normalizedCode = code.replace(/\s/g, "");
    if (normalizedCode.length !== 6) return;
    setIsSubmitting(true);
    try {
      const response = await confirmFeedbackVerificationAction({
        portalSlug: portal.slug,
        code: normalizedCode,
        email,
      });
      if (!response.data?.participant) {
        toast.error("That code could not be verified", {
          description:
            response.error?.message ?? "Request a new code and try again.",
        });
        return;
      }
      onVerified(response.data.participant);
    } catch {
      toast.error("That code could not be verified", {
        description: "Check your connection and try again.",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  if (stage === "code") {
    return (
      <Box className="px-6 py-6">
        <Text as="h2" className="text-xl" fontWeight="semibold">
          Check your email
        </Text>
        <Text className="mt-2 max-w-xl leading-6" color="muted">
          Enter the six-digit code sent to <strong>{email}</strong>. The link in
          that email works too, but this draft will stay here while you enter
          the code.
        </Text>
        <Box className="mt-6">
          <OTPInput
            disabled={isSubmitting}
            hasError={false}
            onChange={setCode}
            value={code}
          />
        </Box>
        <Flex className="mt-6 flex-wrap gap-2" justify="between">
          <Button
            color="tertiary"
            disabled={isSubmitting}
            onClick={() => {
              setStage("identity");
              setCode("");
            }}
            variant="naked"
          >
            Change email
          </Button>
          <Flex gap={2}>
            <Button
              color="tertiary"
              disabled={isSubmitting}
              onClick={() => {
                void requestVerification();
              }}
              variant="outline"
            >
              Send a new code
            </Button>
            <Button
              color="invert"
              disabled={code.replace(/\s/g, "").length !== 6}
              loading={isSubmitting}
              loadingText="Verifying..."
              onClick={() => {
                void confirmVerification();
              }}
            >
              Verify and {purpose}
            </Button>
          </Flex>
        </Flex>
      </Box>
    );
  }

  return (
    <Box className="px-6 py-6">
      <Text as="h2" className="text-xl" fontWeight="semibold">
        Continue with email
      </Text>
      <Text className="mt-2 max-w-xl leading-6" color="muted">
        Verify your email to {purpose} and receive replies and meaningful status
        updates. This does not create a FortyOne account.
      </Text>
      <Box className="mt-6 grid gap-4 sm:grid-cols-2">
        <Input
          autoComplete="email"
          label="Email address"
          onChange={(event) => {
            setEmail(event.target.value);
          }}
          placeholder="you@example.com"
          required
          type="email"
          value={email}
        />
        <Input
          autoComplete="name"
          label="Display name (optional)"
          maxLength={100}
          onChange={(event) => {
            setDisplayName(event.target.value);
          }}
          placeholder="How others should see you"
          value={displayName}
        />
      </Box>
      <Box
        className="border-border/70 bg-surface-muted/40 mt-4 rounded-xl border px-4 py-3"
        role="note"
      >
        <Text className="text-sm leading-5" color="muted">
          {getIdentityPolicyCopy(portal)}
        </Text>
        {portal.guestIdentityPolicy === "allow_public_masking" ? (
          <label
            className="mt-3 flex cursor-pointer items-center gap-2 text-sm"
            htmlFor={maskIdentityId}
          >
            <Checkbox
              checked={hideNamePublicly}
              id={maskIdentityId}
              onCheckedChange={(checked) => {
                setHideNamePublicly(checked === true);
              }}
            />
            Hide my name publicly
          </label>
        ) : null}
      </Box>
      <Flex className="mt-6 gap-2" justify="end">
        {onBack ? (
          <Button color="tertiary" disabled={isSubmitting} onClick={onBack}>
            Back
          </Button>
        ) : null}
        <Button
          color="invert"
          disabled={!email.trim() || isSubmitting}
          loading={isSubmitting}
          loadingText="Sending..."
          onClick={() => {
            void requestVerification();
          }}
        >
          Email me a code
        </Button>
      </Flex>
    </Box>
  );
};

export const FeedbackGuestVerificationDialog = ({
  onOpenChange,
  onVerified,
  open,
  portal,
  purpose,
}: {
  onOpenChange: (open: boolean) => void;
  onVerified: (participant: PublicPortalGuestParticipant) => void;
  open: boolean;
  portal: PublicPortal;
  purpose: string;
}) => (
  <Dialog onOpenChange={onOpenChange} open={open}>
    <Dialog.Content className="max-w-xl" hideClose>
      <FeedbackGuestVerification
        onBack={() => {
          onOpenChange(false);
        }}
        onVerified={(participant) => {
          onVerified(participant);
          onOpenChange(false);
        }}
        portal={portal}
        purpose={purpose}
      />
    </Dialog.Content>
  </Dialog>
);
