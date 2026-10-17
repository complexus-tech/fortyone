//! UserNotifications is used only by the hosted macOS app. The legacy Tauri
//! desktop plugin cannot inspect/request macOS authorization or confirm delivery.
use std::{cell::RefCell, ptr::NonNull, sync::Mutex, time::Duration};

use block2::{DynBlock, RcBlock};
use objc2::{
    define_class, msg_send,
    rc::Retained,
    runtime::{AnyObject, Bool, ProtocolObject},
    AnyThread, DefinedClass, MainThreadMarker,
};
use objc2_foundation::{
    ns_string, NSBundle, NSDictionary, NSError, NSObject, NSObjectProtocol, NSString,
};
use objc2_user_notifications::{
    UNAuthorizationOptions, UNAuthorizationStatus, UNMutableNotificationContent, UNNotification,
    UNNotificationDefaultActionIdentifier, UNNotificationPresentationOptions,
    UNNotificationRequest, UNNotificationResponse, UNNotificationSettings, UNNotificationSound,
    UNUserNotificationCenter, UNUserNotificationCenterDelegate,
};
use serde_json::Value;
use sha2::{Digest, Sha256};
use tauri::{AppHandle, Emitter};
use tokio::sync::oneshot;

use super::{activation_payload, clean_text, Notification};

const CALLBACK_TIMEOUT: Duration = Duration::from_secs(15);
const AUTHORIZATION_TIMEOUT: Duration = Duration::from_secs(120);

struct DelegateState {
    app: AppHandle,
}

define_class!(
    #[unsafe(super(NSObject))]
    #[name = "FortyOneHostedNotificationDelegate"]
    #[ivars = DelegateState]
    struct NotificationDelegate;

    unsafe impl NSObjectProtocol for NotificationDelegate {}

    unsafe impl UNUserNotificationCenterDelegate for NotificationDelegate {
        #[unsafe(method(userNotificationCenter:willPresentNotification:withCompletionHandler:))]
        fn will_present(
            &self,
            _center: &UNUserNotificationCenter,
            _notification: &UNNotification,
            completion: &DynBlock<dyn Fn(UNNotificationPresentationOptions)>,
        ) {
            // The OS still applies the user's alert/sound/Focus settings.
            completion.call((foreground_options(),));
        }

        #[unsafe(method(userNotificationCenter:didReceiveNotificationResponse:withCompletionHandler:))]
        fn did_receive(
            &self,
            _center: &UNUserNotificationCenter,
            response: &UNNotificationResponse,
            completion: &DynBlock<dyn Fn()>,
        ) {
            // Delegate callbacks may run off the main thread. Copy only the
            // bounded target, finish promptly, and let existing routing verify it.
            let payload = if &*response.actionIdentifier()
                == unsafe { UNNotificationDefaultActionIdentifier }
            {
                decode_target(&response.notification().request().content().userInfo())
            } else {
                None
            };
            completion.call(());
            if let Some(payload) = payload {
                let app = self.ivars().app.clone();
                tauri::async_runtime::spawn(async move {
                    if let Err(error) = super::activate(app.clone(), payload).await {
                        let _ = app.emit("desktop-notification-error", error);
                    }
                });
            }
        }
    }
);

thread_local! {
    // UNUserNotificationCenter.delegate is weak. Setup runs on the main thread;
    // retaining here keeps the immutable delegate alive until process shutdown.
    static DELEGATE: RefCell<Option<Retained<NotificationDelegate>>> = const { RefCell::new(None) };
}

fn center() -> Result<Retained<UNUserNotificationCenter>, String> {
    // Apple's center can assert outside an application bundle. `tauri dev` and
    // unit-test executables must fail explicitly instead of crashing the process.
    if NSBundle::mainBundle().bundleIdentifier().is_none() {
        return Err("Open the installed FortyOne app to enable macOS notifications".into());
    }
    Ok(UNUserNotificationCenter::currentNotificationCenter())
}

pub(super) fn install(app: &AppHandle) -> Result<(), String> {
    MainThreadMarker::new().ok_or("Notification setup must run on the main thread")?;
    let Ok(center) = center() else {
        // Unbundled development remains usable; the explicit permission action
        // explains why notifications require a registered .app bundle.
        return Ok(());
    };
    DELEGATE.with(|owner| {
        if owner.borrow().is_none() {
            let allocated =
                NotificationDelegate::alloc().set_ivars(DelegateState { app: app.clone() });
            // SAFETY: initialize NSObject after assigning the Rust ivars once.
            let delegate: Retained<NotificationDelegate> =
                unsafe { msg_send![super(allocated), init] };
            center.setDelegate(Some(ProtocolObject::from_ref(&*delegate)));
            *owner.borrow_mut() = Some(delegate);
        }
    });
    Ok(())
}

fn foreground_options() -> UNNotificationPresentationOptions {
    UNNotificationPresentationOptions::Banner
        | UNNotificationPresentationOptions::List
        | UNNotificationPresentationOptions::Sound
}

fn authorization_options() -> UNAuthorizationOptions {
    UNAuthorizationOptions::Alert | UNAuthorizationOptions::Sound | UNAuthorizationOptions::Badge
}

fn permission_name(status: UNAuthorizationStatus) -> &'static str {
    match status {
        UNAuthorizationStatus::Authorized => "granted",
        UNAuthorizationStatus::NotDetermined => "prompt",
        // Quiet/provisional authorization cannot promise banners or sound.
        _ => "denied",
    }
}

// Objective-C completion blocks can be invoked from any queue. Sending a Rust
// value avoids retaining framework objects across await or blocking the UI loop.
fn complete<T>(sender: &Mutex<Option<oneshot::Sender<T>>>, value: T) {
    if let Ok(mut pending) = sender.lock() {
        if let Some(sender) = pending.take() {
            let _ = sender.send(value);
        }
    }
}

async fn receive<T>(receiver: oneshot::Receiver<T>, timeout: Duration) -> Result<T, String> {
    tokio::time::timeout(timeout, receiver)
        .await
        .map_err(|_| "macOS did not respond to the notification request in time".to_owned())?
        .map_err(|_| "macOS did not complete the notification request".to_owned())
}

pub(super) async fn permission() -> Result<&'static str, String> {
    let (sender, receiver) = oneshot::channel();
    {
        let center = center()?;
        let sender = Mutex::new(Some(sender));
        let callback = RcBlock::new(move |settings: NonNull<UNNotificationSettings>| {
            // SAFETY: Apple guarantees a valid settings object for this callback.
            let status = unsafe { settings.as_ref() }.authorizationStatus();
            complete(&sender, permission_name(status));
        });
        center.getNotificationSettingsWithCompletionHandler(&callback);
    }
    receive(receiver, CALLBACK_TIMEOUT).await
}

pub(super) async fn request_permission() -> Result<String, String> {
    let current = permission().await?;
    if current != "prompt" {
        return Ok(current.into());
    }
    let (sender, receiver) = oneshot::channel();
    {
        let center = center()?;
        let sender = Mutex::new(Some(sender));
        let callback = RcBlock::new(move |_granted: Bool, error: *mut NSError| {
            complete(&sender, native_result(error));
        });
        // This is called only by the existing explicit settings toggle.
        center
            .requestAuthorizationWithOptions_completionHandler(authorization_options(), &callback);
    }
    receive(receiver, AUTHORIZATION_TIMEOUT).await??;
    Ok(permission().await?.into())
}

fn native_result(error: *mut NSError) -> Result<(), String> {
    // SAFETY: framework completion callbacks provide either null or a live NSError.
    unsafe { error.as_ref() }
        .map(|error| Err(error.localizedDescription().to_string()))
        .unwrap_or(Ok(()))
}

fn request_identifier(notification: &Notification) -> String {
    let identity = format!(
        "{}:{}:{}:{}",
        notification.account_id,
        notification.workspace_id,
        notification.notification_id,
        notification.created_at
    );
    format!("fortyone-{:x}", Sha256::digest(identity.as_bytes()))
}

fn content(notification: &Notification) -> Retained<UNMutableNotificationContent> {
    let content = UNMutableNotificationContent::new();
    content.setTitle(&NSString::from_str(&clean_text(&notification.title, 256)));
    content.setBody(&NSString::from_str(&clean_text(&notification.body, 1024)));
    if !notification.silent {
        content.setSound(Some(&UNNotificationSound::defaultSound()));
    }
    let payload = NSString::from_str(&activation_payload(notification).to_string());
    let info = NSDictionary::from_slices(&[ns_string!("fortyoneHostedTarget")], &[&*payload]);
    // SAFETY: erasing generic dictionary types preserves NSString property-list
    // keys/values; UserNotifications copies them into the delivered notification.
    let info: Retained<NSDictionary> = unsafe { Retained::cast_unchecked(info) };
    unsafe { content.setUserInfo(&info) };
    content
}

fn decode_target(info: &NSDictionary) -> Option<Value> {
    let value = info.objectForKey(ns_string!("fortyoneHostedTarget") as &AnyObject)?;
    let encoded = value.downcast_ref::<NSString>()?.to_string();
    if encoded.len() > 8192 {
        return None;
    }
    let value: Value = serde_json::from_str(&encoded).ok()?;
    serde_json::from_value::<Notification>(value.clone())
        .ok()?
        .validate()
        .ok()?;
    Some(value)
}

pub(super) async fn show(notification: &Notification) -> Result<(), String> {
    let (sender, receiver) = oneshot::channel();
    {
        let center = center()?;
        let request = UNNotificationRequest::requestWithIdentifier_content_trigger(
            &NSString::from_str(&request_identifier(notification)),
            &content(notification),
            None,
        );
        let sender = Mutex::new(Some(sender));
        let callback =
            RcBlock::new(move |error: *mut NSError| complete(&sender, native_result(error)));
        center.addNotificationRequest_withCompletionHandler(&request, Some(&callback));
    }
    // Success means macOS accepted the request; Focus and user settings can still
    // suppress presentation. A failed submission releases the durable receipt.
    receive(receiver, CALLBACK_TIMEOUT).await?
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn permission_status_is_not_assumed_granted() {
        assert_eq!(
            permission_name(UNAuthorizationStatus::NotDetermined),
            "prompt"
        );
        assert_eq!(permission_name(UNAuthorizationStatus::Denied), "denied");
        assert_eq!(
            permission_name(UNAuthorizationStatus::Authorized),
            "granted"
        );
        assert_eq!(
            permission_name(UNAuthorizationStatus::Provisional),
            "denied"
        );
        assert_eq!(permission_name(UNAuthorizationStatus(99)), "denied");
    }

    #[test]
    fn foreground_presentation_and_authorization_request_sound_and_alerts() {
        assert!(foreground_options().contains(
            UNNotificationPresentationOptions::Banner
                | UNNotificationPresentationOptions::List
                | UNNotificationPresentationOptions::Sound
        ));
        assert_eq!(
            authorization_options(),
            UNAuthorizationOptions::Alert
                | UNAuthorizationOptions::Sound
                | UNAuthorizationOptions::Badge
        );
    }

    #[test]
    fn content_has_default_sound_and_validated_activation_payload() {
        let notification = super::super::tests::notification();
        let content = content(&notification);
        assert!(content.sound().is_some());
        assert_eq!(content.title().to_string(), notification.title);
        assert_eq!(
            decode_target(&content.userInfo()),
            Some(activation_payload(&notification))
        );
        assert!(
            content.badge().is_none(),
            "Unread count has its own scoped path"
        );
    }

    #[test]
    fn request_identity_is_stable_and_distinguishes_scope_and_coalesced_version() {
        let notification = super::super::tests::notification();
        let original = request_identifier(&notification);
        assert_eq!(original, request_identifier(&notification));
        let mut later = notification.clone();
        later.created_at = "2026-10-15T12:01:00Z".into();
        assert_ne!(original, request_identifier(&later));
        let mut other_account = notification.clone();
        other_account.account_id = uuid::Uuid::new_v4().to_string();
        assert_ne!(original, request_identifier(&other_account));
        let mut other_workspace = notification;
        other_workspace.workspace_id = uuid::Uuid::new_v4().to_string();
        assert_ne!(original, request_identifier(&other_workspace));
    }

    #[test]
    fn activation_payload_cannot_supply_an_arbitrary_destination() {
        let mut notification = super::super::tests::notification();
        notification.entity_type = "https://untrusted.example".into();
        assert!(decode_target(&content(&notification).userInfo()).is_none());
        notification.entity_type = "story".into();
        notification.workspace_id = "not-a-workspace".into();
        assert!(decode_target(&content(&notification).userInfo()).is_none());
    }

    #[test]
    fn chat_sound_respects_the_explicit_silent_preference() {
        let mut notification = super::super::tests::notification();
        notification.entity_type = "chat_message".into();
        assert!(content(&notification).sound().is_some());
        notification.silent = true;
        assert!(content(&notification).sound().is_none());
    }

    #[tokio::test]
    async fn callbacks_complete_once_and_missing_callbacks_are_bounded() {
        let (sender, receiver) = oneshot::channel();
        let sender = Mutex::new(Some(sender));
        complete(&sender, Ok::<(), String>(()));
        complete(&sender, Err("duplicate".into()));
        assert_eq!(receive(receiver, CALLBACK_TIMEOUT).await, Ok(Ok(())));
        let (sender, receiver) = oneshot::channel();
        complete(
            &Mutex::new(Some(sender)),
            Err::<(), String>("submission failed".into()),
        );
        assert_eq!(
            receive(receiver, CALLBACK_TIMEOUT).await,
            Ok(Err("submission failed".into()))
        );
        let (_sender, receiver) = oneshot::channel::<()>();
        assert!(receive(receiver, Duration::from_millis(1)).await.is_err());
    }
}
