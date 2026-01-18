export const clearAllStorage = () => {
  // Clear NextAuth session cookies with proper domain configuration
  const domain = process.env.NEXT_PUBLIC_DOMAIN ? `.${process.env.NEXT_PUBLIC_DOMAIN}` : "";
  if (domain) {
    document.cookie = "__Secure-next-auth.session-token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/; domain=" + domain + "; secure;";
  }

  // Clear other client-side cookies
  document.cookie.split(";").forEach((c) => {
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
