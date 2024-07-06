"use server";

export async function authenticateUser({ email = "", password = "" }) {
  //   const user = await post<{ email: string; password: string }, User>(
  //     `${apiUrl}/login/`,
  //     { email, password }
  //   );

  return {
    id: "1",
    name: "Joseph Mukorivo",
    email: "josemukorivo@gmail.com",
    jwt: "fake-jwt",
    customField: "custom-field",
    image:
      "https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo",
  };
}
