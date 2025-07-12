import { Instrument_Sans as InstrumentSans, Inter } from "next/font/google";
import localFont from "next/font/local";

export const satoshi = localFont({
  src: "../styles/Satoshi-Variable.woff2",
  // variable: "--font-satoshi",
  weight: "450 900",
  display: "swap",
  style: "normal",
});

export const instrumentSans = InstrumentSans({
  variable: "--font-sans",
  subsets: ["latin"],
  display: "swap",
  weight: "variable",
});

export const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
  display: "swap",
  weight: ["400", "500", "600", "700"],
});
