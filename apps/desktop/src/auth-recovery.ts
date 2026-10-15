export type AuthRecovery = "sign-in" | "credential-access";

// These native error prefixes are a stable renderer contract. A network or
// workspace-permission error must never be treated as a missing account session.
export function authRecovery(message?: string | null): AuthRecovery | null {
  if (
    message?.startsWith(
      "Secure credential access was not granted or is unavailable.",
    )
  )
    return "credential-access";
  if (
    message?.startsWith("Sign in again to synchronize.") ||
    message?.startsWith("Your saved desktop session is invalid.")
  )
    return "sign-in";
  return null;
}
