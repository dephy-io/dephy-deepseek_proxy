use nostr::Event;
use nostr::PublicKey;
use nostr::RelayMessage;
use nostr::Timestamp;
use nostr_sdk::RelayPoolNotification;

use crate::message::ChatMessage;
use crate::node::ds_client::RequestMessage;
use crate::relay_client::extract_mention;
use crate::RelayClient;

use super::ds_client::{ChatCompletionRequest, DsClient};

#[derive(Debug, thiserror::Error)]
#[non_exhaustive]
pub enum Error {
    #[error("Serde json error: {0}")]
    SerdeJson(#[from] serde_json::Error),
    #[error("Nostr key error: {0}")]
    NostrKey(#[from] nostr::key::Error),
    #[error("Relay client error: {0}")]
    RelayClient(#[from] crate::relay_client::Error),
}

pub struct MessageHandler {
    client: RelayClient,
    ds_client: DsClient,
    solana_keypair_path: String,
    controller_pubkey: PublicKey,
    admin_pubkey: PublicKey,
    started_at: Timestamp,
    request_cache: Vec<RequestMessage>,
}

impl MessageHandler {
    pub fn new(
        client: RelayClient,
        ds_client: DsClient,
        solana_keypair_path: &str,
        controller_pubkey: PublicKey,
        admin_pubkey: PublicKey,
    ) -> Self {
        let started_at = Timestamp::now();

        Self {
            client,
            ds_client,
            solana_keypair_path: solana_keypair_path.to_string(),
            controller_pubkey,
            admin_pubkey,
            started_at,
            request_cache: vec![],
        }
    }

    pub async fn run(mut self) {
        let mut notifications = self.client.notifications();

        let checking_client = self.client.clone();
        let relay_checker = async move {
            checking_client
                .run_relay_checker(std::time::Duration::from_secs(10))
                .await
        };

        let message_handler = async move {
            // TODO: get uuid from controllers‘ communication
            let uuid = String::from("664385e1a27240d7bbcd2ca83212445e");
            let mentions = vec![uuid.clone()];

            let sub_id = self
                .client
                .subscribe_all(None, Some(mentions))
                .await
                .expect("Failed to subscribe events");

            loop {
                let notification = notifications
                    .recv()
                    .await
                    .expect("Failed to receive notification");
                // tracing::debug!("Received notification: {:?}", notification);

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
                    } if subscription_id == sub_id => {}

                    RelayPoolNotification::Message {
                        message:
                            RelayMessage::Event {
                                event,
                                subscription_id,
                            },
                        ..
                    } => {
                        if subscription_id == sub_id {
                            let Ok(message) = serde_json::from_str::<ChatMessage>(&event.content)
                            else {
                                tracing::error!("Failed to parse message: {:?}", event);
                                continue;
                            };

                            self.handle_message(&event, &message)
                                .await
                                .expect("Failed to handle message");
                        }
                    }

                    _ => {}
                }
            }
        };

        futures::join!(relay_checker, message_handler);
    }

    async fn handle_message(&mut self, event: &Event, message: &ChatMessage) -> Result<(), Error> {
        match message {
            ChatMessage::Ask {
                role,
                content,
                name,
            } => {
                tracing::debug!("Ask: {:?}", content);

                let Some(mention) = extract_mention(event) else {
                    tracing::error!("Machine not mentioned in event, skip event: {:?}", event);
                    return Ok(());
                };
                
                let msg = RequestMessage {
                    role: role.into(),
                    content: content.clone(),
                    name: Some(name.into()),
                };

                if event.created_at < self.started_at {
                    // history messages
                    self.request_cache.push(msg)
                } else {
                    // new messages
                    let mut messages = self.request_cache.clone();
                    messages.push(msg.clone());
    
                    match self
                        .ds_client
                        .create_chat_completion(ChatCompletionRequest {
                            model: "deepseek/deepseek-r1/community".into(),
                            messages,
                            max_tokens: 10240,
                            temperature: None,
                            top_p: None,
                            n: None,
                            stream: None,
                            stop: None,
                            presence_penalty: None,
                            frequency_penalty: None,
                            logit_bias: None,
                            user: None,
                            top_k: None,
                            min_p: None,
                            repetition_penalty: None,
                            logprobs: None,
                            top_logprobs: None,
                            response_format: None,
                            seed: None,
                        })
                        .await
                    {
                        Ok(response) => {
                            self.client
                                .send_event(
                                    mention,
                                    &ChatMessage::Anwser {
                                        finish_reason: response.choices[0].finish_reason.clone(),
                                        role: response.choices[0].message.role.clone(),
                                        content: response.choices[0].message.content.clone(),
                                    },
                                )
                                .await?;
    
                            self.request_cache.push(msg);
                        }
                        Err(err) => {
                            tracing::error!("Failed to process anwser message: {:?}", err);
                        }
                    }
                }

            }

            ChatMessage::Anwser { role, content, finish_reason } => {
                if finish_reason == "stop" {
                    tracing::debug!("Anwser: {:?}", content);
                    let msg = RequestMessage {
                        role: role.into(),
                        content: content.clone(),
                        name: None,
                    };
    
                    self.request_cache.push(msg);
                }
            }
        }
        Ok(())
    }
}
