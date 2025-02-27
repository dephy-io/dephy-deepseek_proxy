pub mod message;
pub mod node;
pub mod utils;
mod relay_client;

/// chat-controller version
pub const VERSION: &str = env!("CARGO_PKG_VERSION");

pub use relay_client::RelayClient;
