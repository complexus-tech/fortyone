import { Inter, Oswald, IBM_Plex_Sans as IBMPlexSans } from "next/font/google";

export const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
  display: "swap",
  weight: ["400", "500", "600", "700"],
});

export const oswald = Oswald({
  variable: "--font-oswald",
  subsets: ["latin"],
  display: "swap",
  weight: ["300", "400", "500", "600", "700"],
});

export const ibmplexsans = IBMPlexSans({
  variable: "--font-ibmplexsans",
  subsets: ["latin"],
  display: "swap",
  weight: ["400", "500", "600", "700"],
});
