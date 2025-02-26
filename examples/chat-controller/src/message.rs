// use nostr::EventId;
use serde::Deserialize;
use serde::Serialize;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ChatMessage {
    NewChat {
        uuid: String,
    },
    Ask {
        name: String,
        role: String,
        content: Option<String>,
    },
    Anwser {
        finish_reason: String,
        role: String,
        content: Option<String>,
    },
}
