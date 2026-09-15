#!/usr/bin/env python3
"""Validate bundled privacy resources without building, signing, or uploading."""

import argparse
import plistlib
import re
import sys
from pathlib import Path
from xml.parsers.expat import ExpatError


class ValidationError(Exception):
    pass


def load_plist(path):
    try:
        with path.open("rb") as stream:
            value = plistlib.load(stream)
    except (OSError, ValueError, plistlib.InvalidFileException, ExpatError) as error:
        raise ValidationError(f"Cannot read valid plist: {path}") from error
    if not isinstance(value, dict):
        raise ValidationError(f"Expected a plist dictionary: {path}")
    return value


def application_path(path):
    if path.suffix == ".xcarchive":
        applications = list((path / "Products/Applications").glob("*.app"))
        if len(applications) != 1:
            raise ValidationError("The archive must contain exactly one main app.")
        return applications[0]
    if path.suffix != ".app" or not path.is_dir():
        raise ValidationError("Provide an existing .app or .xcarchive directory.")
    return path


def verify_resources(app):
    load_plist(app / "PrivacyInfo.xcprivacy")
    manifests = sorted(app.rglob("PrivacyInfo.xcprivacy"))
    vendor_manifests = []
    for path in manifests:
        manifest = load_plist(path)
        # An app-level aggregate does not replace the SDK's own resource. Both
        # CocoaPods' resource bundle and a vendor's self-contained framework work.
        relative = path.relative_to(app)
        if {"SDWebImage.bundle", "SDWebImage.framework"}.intersection(relative.parts):
            vendor_manifests.append((path, manifest))

    if not vendor_manifests:
        raise ValidationError(
            "SDWebImage's bundled PrivacyInfo.xcprivacy is missing. "
            "Regenerate pods with expo.autolinking.ios.buildFromSource=['expo-image']; "
            "do not substitute an app-only declaration for the vendor resource."
        )

    for path, manifest in vendor_manifests:
        accessed = manifest.get("NSPrivacyAccessedAPITypes", [])
        if not isinstance(accessed, list):
            raise ValidationError(f"Invalid accessed-API array: {path}")
        has_file_reason = any(
            isinstance(entry, dict)
            and entry.get("NSPrivacyAccessedAPIType")
            == "NSPrivacyAccessedAPICategoryFileTimestamp"
            and isinstance(entry.get("NSPrivacyAccessedAPITypeReasons"), list)
            and "C617.1" in entry.get("NSPrivacyAccessedAPITypeReasons", [])
            for entry in accessed
        )
        if not has_file_reason or manifest.get("NSPrivacyTracking") is not False:
            raise ValidationError(
                f"SDWebImage's manifest differs from the reviewed vendor policy: {path}. "
                "Review its file-timestamp reason and tracking declaration."
            )
    return [str(path.relative_to(app)) for path, _ in vendor_manifests]


def verify_release(app):
    info = load_plist(app / "Info.plist")
    sdk = str(info.get("DTSDKName", ""))
    version = re.fullmatch(r"iphoneos(\d+)(?:\.\d+)*", sdk)
    if not version or int(version.group(1)) < 26:
        raise ValidationError(
            "A release check requires a device build using iOS SDK 26 or later; "
            f"this app reports {sdk or 'no DTSDKName'}."
        )
    if info.get("CFBundleIdentifier") != "com.fortyone.mobile":
        raise ValidationError("Unexpected release bundle identifier.")
    for key in ("CFBundleShortVersionString", "CFBundleVersion"):
        if not info.get(key):
            raise ValidationError(f"Release metadata is missing {key}.")
    if "_expo._tcp" in info.get("NSBonjourServices", []):
        raise ValidationError("The release still declares Expo development discovery.")
    if "Expo Dev Launcher" in info.get("NSLocalNetworkUsageDescription", ""):
        raise ValidationError("The release still contains Expo's development permission text.")
    return sdk


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("artifact", type=Path, help="Built .app or .xcarchive")
    parser.add_argument(
        "--release", action="store_true", help="Also require device SDK and release metadata"
    )
    args = parser.parse_args()
    try:
        app = application_path(args.artifact.resolve())
        manifests = verify_resources(app)
        release = args.release or args.artifact.suffix == ".xcarchive"
        sdk = verify_release(app) if release else None
    except ValidationError as error:
        print(f"iOS privacy validation failed: {error}", file=sys.stderr)
        return 1
    print("iOS privacy resources verified:")
    for manifest in manifests:
        print(f"  {manifest}")
    if sdk:
        print(f"Release bundle metadata verified ({sdk}).")
    print("Signing, SDK signatures, App Store privacy labels, and runtime data collection are not validated.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
