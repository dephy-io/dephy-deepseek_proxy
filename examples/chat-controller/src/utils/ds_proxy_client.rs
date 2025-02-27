use nostr::message::RelayMessage;
use nostr::EventBuilder;
use nostr::EventId;
use nostr::Filter;
use nostr::Keys;
use nostr::SingleLetterTag;
use nostr::Tag;
use nostr_sdk::Client;
use nostr_sdk::RelayPoolNotification;
use serde::{Deserialize, Serialize};
use crate::relay_client::Error;

const EVENT_KIND: nostr::Kind = nostr::Kind::Custom(1573);
const MENTION_TAG: SingleLetterTag = SingleLetterTag::lowercase(nostr::Alphabet::P);
const SESSION_TAG: SingleLetterTag = SingleLetterTag::lowercase(nostr::Alphabet::S);

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
#[repr(u8)]
enum DephyDsProxyStatus {
    Available = 1,
    Working = 2,
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
#[repr(u8)]
enum DephyDsProxyStatusReason {
    UserRequest = 1,
    AdminRequest = 2,
    UserBehaviour = 3,
    Reset = 4,
    LockFailed = 5,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
enum DephyDsProxyMessage {
    Request {
        to_status: DephyDsProxyStatus,
        reason: DephyDsProxyStatusReason,
        initial_request: EventId,
        payload: String,
    },
    Status {
        status: DephyDsProxyStatus,
        reason: DephyDsProxyStatusReason,
        initial_request: EventId,
        payload: String,
    },
    Account {
        user: String,
        tokens: u32,
    },
}

#[derive(Clone)]
pub struct DsProxyClient {
    client: Client,
    session: String,
    mention: String,
}

impl DsProxyClient {
    pub async fn new(
        nostr_relay: &str,
        keys: &Keys,
        session: &str,
        mention: &str,
        max_notification_size: usize,
    ) -> Result<Self, Error> {
        let client_opts =
            nostr_sdk::Options::default().notification_channel_size(max_notification_size);

        let client = Client::builder()
            .signer(keys.clone())
            .opts(client_opts)
            .build();

        client.add_relay(nostr_relay).await?;
        client.connect().await;

        Ok(Self {
            client,
            session: session.to_string(),
            mention: mention.to_string()
        })
    }

    pub async fn fetch_user_tokens(
        &self,
        target_user: &str,
        // mention: &str,
    ) -> Result<u32, Error> {
        let filter = Filter::new()
            .kind(EVENT_KIND)
            .custom_tag(SESSION_TAG, [self.session.clone()])
            .custom_tag(MENTION_TAG, [self.mention.clone()]);

        // ["dephy-dsproxy-controller"]
        // ["d041ea9854f2117b82452457c4e6d6593a96524027cd4032d2f40046deb78d93"]

        let output = self.client.subscribe(vec![filter], None).await?;

        let sub_id = output.id().clone();

        let mut target_tokens: Option<u32> = None;

        let mut notifications = self.client.notifications();

        loop {
            let notification = notifications
                .recv()
                .await
                .expect("Failed to receive notification");

            match notification {
                RelayPoolNotification::Shutdown => panic!("Relay pool shutdown"),

                RelayPoolNotification::Message {
                    message:
                        RelayMessage::Closed {
                            message,
                            subscription_id,
                        },
                    ..
                } if subscription_id == sub_id => {
                    tracing::error!("Subscription closed: {}", message);
                    panic!("Subscription closed: {message}");
                }

                RelayPoolNotification::Message {
                    message: RelayMessage::EndOfStoredEvents(subscription_id),
                    ..
                } if subscription_id == sub_id => {
                    break;
                }

                RelayPoolNotification::Message {
                    message:
                        RelayMessage::Event {
                            event,
                            subscription_id,
                        },
                    ..
                } => {
                    if subscription_id == sub_id {
                        let Ok(message) =
                            serde_json::from_str::<DephyDsProxyMessage>(&event.content)
                        else {
                            tracing::error!("Failed to parse message: {:?}", event);
                            continue;
                        };

                        match message {
                            DephyDsProxyMessage::Account { user, tokens } => {
                                if user == target_user {
                                    target_tokens = Some(tokens);
                                }
                            }

                            _ => {}
                        }
                    }
                }

                _ => {}
            }
        }

        return Ok(target_tokens.unwrap_or(0));
    }

    pub async fn update_user_account(
        &self,
        user: String,
        tokens: u32,
    ) -> Result<(), Error> {
        self.send_event(&self.mention, &DephyDsProxyMessage::Account { user, tokens })
            .await?;

        Ok(())
    }

    async fn send_event<M>(&self, to: &str, message: &M) -> Result<(), Error>
    where
        M: serde::Serialize + std::fmt::Debug,
    {
        let content = serde_json::to_string(message)?;

        let event_builder = EventBuilder::new(EVENT_KIND, content).tags([
            Tag::parse(["s".to_string(), self.session.clone()]).unwrap(),
            Tag::parse(["p".to_string(), to.to_string()]).unwrap(),
        ]);

        let res = self.client.send_event_builder(event_builder).await?;

        if !res.failed.is_empty() {
            for (relay_url, err) in res.failed.iter() {
                tracing::error!("failed to send event to {} err: {:?}", relay_url, err);
            }
            return Err(Error::SendEvent(format!(
                "Failed to send event {message:?} to relay"
            )));
        }

        Ok(())
    }
}
