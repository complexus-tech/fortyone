import type { ComponentPropsWithoutRef, ReactNode } from "react";
import { Button } from "ui";
import { GoogleIcon, MicrosoftIcon } from "@/components/ui";
import { GOOGLE_SIGNUP_URL, MICROSOFT_SIGNUP_URL } from "@/lib/app-url";

type SignupProvider = "google" | "microsoft";

const PROVIDERS: Record<
  SignupProvider,
  { href: string; icon: ReactNode; name: string }
> = {
  google: {
    href: GOOGLE_SIGNUP_URL,
    icon: <GoogleIcon aria-hidden="true" className="size-5 shrink-0" />,
    name: "Google",
  },
  microsoft: {
    href: MICROSOFT_SIGNUP_URL,
    icon: <MicrosoftIcon aria-hidden="true" className="size-5 shrink-0" />,
    name: "Microsoft",
  },
};

interface SignupProviderButtonProps
  extends Pick<ComponentPropsWithoutRef<typeof Button>, "className"> {
  emphasized?: boolean;
  label?: "Continue" | "Sign up";
  provider: SignupProvider;
}

export const SignupProviderButton = ({
  className,
  emphasized = false,
  label = "Continue",
  provider,
}: SignupProviderButtonProps) => {
  const providerConfig = PROVIDERS[provider];

  return (
    <Button
      align="center"
      className={className}
      color={emphasized ? "invert" : "tertiary"}
      href={providerConfig.href}
      leftIcon={providerConfig.icon}
      prefetch={false}
      rounded="lg"
      size="lg"
      variant={emphasized ? "solid" : "naked"}
    >
      {label} with {providerConfig.name}
    </Button>
  );
};
