"use server";

import { cookies } from "next/headers";
import ky from "ky";
import { signOut } from "@/auth";
import { getApiUrl } from "@/lib/api-url";
import { buildAuthHeaders } from "@/lib/http/auth-headers";
import { getCookieHeader } from "@/lib/http/header";
import { switchWorkspace } from "@/lib/actions/users/switch-workspace";

const SESSION_COOKIE_NAMES = [
  "__Secure-next-auth.session-token",
  "next-auth.session-token",
  "fortyone_session",
];

export const logOut = async () => {
  const apiUrl = getApiUrl();
  const cookieHeader = await getCookieHeader();

  try {
    await ky.delete(`${apiUrl}/users/session`, {
      credentials: "include",
      headers: buildAuthHeaders({ cookieHeader }),
    });
  } catch {
    // Best-effort cookie clearing; continue logout flow.
  }

  await signOut({ redirect: false });

  const cookieStore = await cookies();
  const domain = process.env.NEXT_PUBLIC_DOMAIN
    ? `.${process.env.NEXT_PUBLIC_DOMAIN}`
    : undefined;

  SESSION_COOKIE_NAMES.forEach((name) => {
    cookieStore.set(name, "", {
      expires: new Date(0),
      httpOnly: true,
      path: "/",
      secure: true,
    });
    if (domain) {
      cookieStore.set(name, "", {
        domain,
        expires: new Date(0),
        httpOnly: true,
        path: "/",
        secure: true,
      });
    }
  });
};

export const changeWorkspace = async (workspaceId: string) => {
  await switchWorkspace(workspaceId);
};
