import { SignupProviderButton } from "./signup-provider-button";

export const GoogleSignupButton = () => {
  return (
    <SignupProviderButton
      className="bg-state-hover hover:bg-state-active relative z-1 px-3 md:px-4"
      provider="google"
    />
  );
};
