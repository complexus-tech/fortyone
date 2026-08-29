const nextJest = require("next/jest");

const createJestConfig = nextJest({
  // Provide the path to the Next.js app so next/jest can load its config and env files.
  dir: "./",
});

/** @type {import("jest").Config} */
const config = {
  coverageProvider: "v8",
  modulePathIgnorePatterns: ["<rootDir>/.next/"],
  moduleNameMapper: {
    "^@/(.*)$": "<rootDir>/src/$1",
    "^chrono-node/en$":
      "<rootDir>/node_modules/chrono-node/dist/cjs/locales/en/index.js",
    "^react$": "<rootDir>/node_modules/react",
    "^react/(.*)$": "<rootDir>/node_modules/react/$1",
    "^react-dom$": "<rootDir>/node_modules/react-dom",
    "^react-dom/(.*)$": "<rootDir>/node_modules/react-dom/$1",
    "^react-resizable-panels$":
      "<rootDir>/jest.mocks/react-resizable-panels.cjs",
  },
  testEnvironment: "jsdom",
  setupFilesAfterEnv: ["<rootDir>/jest.setup.ts"],
};

// Export through next/jest because loading the Next.js configuration is asynchronous.
module.exports = createJestConfig(config);
