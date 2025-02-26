pub mod message;
pub mod node;
mod relay_client;

/// dephy-dsproxy-controller version
pub const VERSION: &str = env!("CARGO_PKG_VERSION");

pub use relay_client::RelayClient;
