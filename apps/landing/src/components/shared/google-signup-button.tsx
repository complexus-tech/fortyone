import { Button } from "ui";
import { GoogleIcon } from "@/components/ui";
import { GOOGLE_SIGNUP_URL } from "@/lib/app-url";

export const GoogleSignupButton = () => {
  return (
    <Button
      className="bg-state-hover hover:bg-state-active relative z-1 px-3 md:px-4"
      color="tertiary"
      href={GOOGLE_SIGNUP_URL}
      leftIcon={<GoogleIcon aria-hidden="true" className="size-5 shrink-0" />}
      prefetch={false}
      rounded="lg"
      size="lg"
      variant="naked"
    >
      Continue with Google
    </Button>
  );
};
