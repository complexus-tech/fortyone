import "@testing-library/jest-dom";
import { TextDecoder, TextEncoder } from "node:util";

Object.defineProperties(globalThis, {
  TextDecoder: { configurable: true, value: TextDecoder, writable: true },
  TextEncoder: { configurable: true, value: TextEncoder, writable: true },
});

if (typeof window !== "undefined") {
  if (
    typeof globalThis.PointerEvent === "undefined" &&
    typeof globalThis.MouseEvent !== "undefined"
  ) {
    Object.defineProperty(globalThis, "PointerEvent", {
      configurable: true,
      value: globalThis.MouseEvent,
      writable: true,
    });
  }

  Object.defineProperty(window, "__NEXT_PUBLIC_ENV", {
    configurable: true,
    value: {
      API_URL: process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8000",
      NODE_ENV: "test",
    },
  });
}
