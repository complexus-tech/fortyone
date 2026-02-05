const SESSION_COOKIE_NAMES = [
  "__Secure-next-auth.session-token",
  "next-auth.session-token",
  "fortyone_session",
];

export const clearAllStorage = () => {
  const domain =
    process.env.NEXT_PUBLIC_DOMAIN === "fortyone.app"
      ? ".fortyone.app"
      : undefined;

  // Clear session cookies with domain
  SESSION_COOKIE_NAMES.forEach((name) => {
    // Clear without domain (for cookies set without domain)
    document.cookie = `${name}=;expires=${new Date(0).toUTCString()};path=/`;
    // Clear with domain (for cookies set with domain)
    if (domain) {
      document.cookie = `${name}=;expires=${new Date(0).toUTCString()};path=/;domain=${domain}`;
    }
  });

  // Clear other client-side cookies
  document.cookie.split(";").forEach((c) => {
    const cookieName = c.replace(/^ +/, "").split("=")[0];
    // Skip session cookies (already handled above)
    if (SESSION_COOKIE_NAMES.includes(cookieName)) return;
    document.cookie = c
      .replace(/^ +/, "")
      .replace(/=.*/, `=;expires=${new Date().toUTCString()};path=/`);
  });

  localStorage.clear();

  if ("caches" in window) {
    caches.keys().then((names) => {
      names.forEach((name) => {
        caches.delete(name);
      });
    });
  }
};
