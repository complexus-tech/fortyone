import assert from "node:assert/strict";
import { execFileSync, spawnSync } from "node:child_process";
import {
  mkdtempSync,
  mkdirSync,
  renameSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import { createRequire } from "node:module";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";
import type { TestContext } from "node:test";

const verifier = fileURLToPath(
  new URL("./verify-ios-privacy.py", import.meta.url),
);
const plist = (body: string) =>
  `<?xml version="1.0"?><plist version="1.0"><dict>${body}</dict></plist>`;
const vendorPolicy = plist(`<key>NSPrivacyTracking</key><false/>
  <key>NSPrivacyAccessedAPITypes</key><array><dict>
  <key>NSPrivacyAccessedAPIType</key><string>NSPrivacyAccessedAPICategoryFileTimestamp</string>
  <key>NSPrivacyAccessedAPITypeReasons</key><array><string>C617.1</string></array>
  </dict></array>`);

function artifactFixture(t: TestContext) {
  const directory = mkdtempSync(join(tmpdir(), "fortyone-privacy-"));
  t.after(() => rmSync(directory, { recursive: true, force: true }));
  const app = join(directory, "FortyOne.app");
  mkdirSync(app);
  writeFileSync(join(app, "PrivacyInfo.xcprivacy"), plist(""));
  const vendor = (location = "SDWebImage.bundle", content = vendorPolicy) => {
    mkdirSync(join(app, location), { recursive: true });
    writeFileSync(join(app, location, "PrivacyInfo.xcprivacy"), content);
  };
  const check = (...args: string[]) =>
    spawnSync("python3", [verifier, app, ...args], { encoding: "utf8" });
  return { directory, app, vendor, check };
}

test("app-level API reasons do not conceal a missing SDK manifest", (t) => {
  const { app, check } = artifactFixture(t);
  writeFileSync(join(app, "PrivacyInfo.xcprivacy"), vendorPolicy);
  const result = check();
  assert.equal(result.status, 1);
  assert.match(
    result.stderr,
    /SDWebImage's bundled PrivacyInfo.xcprivacy is missing/,
  );
});

test("the vendor resource bundle and self-contained framework are both supported", (t) => {
  const { vendor, check } = artifactFixture(t);
  vendor();
  vendor("Frameworks/SDWebImage.framework");
  const result = check();
  assert.equal(result.status, 0, result.stderr);
  assert.match(result.stdout, /SDWebImage.bundle/);
  assert.match(result.stdout, /SDWebImage.framework/);
});

test("a present but empty vendor manifest fails closed", (t) => {
  const { vendor, check } = artifactFixture(t);
  vendor("SDWebImage.bundle", plist(""));
  const result = check();
  assert.equal(result.status, 1);
  assert.match(result.stderr, /differs from the reviewed vendor policy/);
});

test("malformed bundled privacy manifests cannot pass validation", (t) => {
  const { vendor, check } = artifactFixture(t);
  vendor("SDWebImage.bundle", "not a plist");
  const result = check();
  assert.equal(result.status, 1);
  assert.match(result.stderr, /Cannot read valid plist/);
});

test("a simulator build cannot pass the release gate", (t) => {
  const { app, vendor, check } = artifactFixture(t);
  vendor();
  writeFileSync(
    join(app, "Info.plist"),
    plist("<key>DTSDKName</key><string>iphonesimulator27.0</string>"),
  );
  const result = check("--release");
  assert.equal(result.status, 1);
  assert.match(result.stderr, /requires a device build/);
});

test("device release metadata passes without claiming signing or store validation", (t) => {
  const { app, vendor, check } = artifactFixture(t);
  vendor();
  writeFileSync(
    join(app, "Info.plist"),
    plist(`<key>DTSDKName</key><string>iphoneos27.0</string>
    <key>CFBundleIdentifier</key><string>com.fortyone.mobile</string>
    <key>CFBundleShortVersionString</key><string>1.0.0</string>
    <key>CFBundleVersion</key><string>1</string>`),
  );
  const result = check("--release");
  assert.equal(result.status, 0, result.stderr);
  assert.match(result.stdout, /Signing, SDK signatures.*are not validated/);
});

test("archives automatically enforce device release checks", (t) => {
  const { directory, app, vendor } = artifactFixture(t);
  vendor();
  writeFileSync(
    join(app, "Info.plist"),
    plist("<key>DTSDKName</key><string>iphonesimulator27.0</string>"),
  );
  const archive = join(directory, "FortyOne.xcarchive");
  const applications = join(archive, "Products/Applications");
  mkdirSync(applications, { recursive: true });
  renameSync(app, join(applications, "FortyOne.app"));
  const result = spawnSync("python3", [verifier, archive], {
    encoding: "utf8",
  });
  assert.equal(result.status, 1);
  assert.match(result.stderr, /requires a device build/);
});

test("installed Expo autolinking applies the source-build fix only to iOS", () => {
  const require = createRequire(import.meta.url);
  const expoRequire = createRequire(require.resolve("expo/package.json"));
  const cli = expoRequire.resolve(
    "expo-modules-autolinking/bin/expo-modules-autolinking.js",
  );
  const resolve = (platform: string) => {
    const output = execFileSync(
      process.execPath,
      [cli, "resolve", "--platform", platform, "--json"],
      {
        cwd: fileURLToPath(new URL("..", import.meta.url)),
        encoding: "utf8",
      },
    );
    return JSON.parse(output) as {
      configuration?: { buildFromSource?: string[] };
    };
  };
  assert.ok(
    resolve("apple").configuration?.buildFromSource?.includes("expo-image"),
  );
  assert.ok(
    !resolve("android").configuration?.buildFromSource?.includes("expo-image"),
  );
});
