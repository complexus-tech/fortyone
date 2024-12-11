"use server";

export async function authenticateUser({ email = "", password = "" }) {
  //   const user = await post<{ email: string; password: string }, User>(
  //     `${apiUrl}/login/`,
  //     { email, password }
  //   );

  return {
    id: "8a798112-90fe-495e-9f1c-f36655e3d8ab",
    name: "Joseph Mukorivo",
    email: "josemukorivo@gmail.com",
    token:
      "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI4YTc5ODExMi05MGZlLTQ5NWUtOWYxYy1mMzY2NTVlM2Q4YWIiLCJleHAiOjE3MzI0NDc1MjAsIm5iZiI6MTczMjM2MTEyMCwiaWF0IjoxNzMyMzYxMTIwfQ.zK4bDNPCvOYNfLq7FXTpbgMk7MD48EiP9HadtTaloW8",
    workspaces: [
      {
        id: "3589aaa4-f1f4-40bb-ae1c-9104dd537d8c",
        name: "Complexus",
        isActive: true,
      },
    ],
    image:
      "https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo",
  };
}
