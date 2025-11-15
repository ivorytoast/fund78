use std::sync::Arc;
use tokio::sync::broadcast::{self, error::RecvError};

/// Handle for receiving events from the engine
/// Users interact with this abstraction, never seeing the underlying channel
pub struct OutputHandle {
    sender: Arc<broadcast::Sender<String>>,
}

impl OutputHandle {
    /// Create a new OutputHandle (only Engine should call this)
    pub(crate) fn new(sender: Arc<broadcast::Sender<String>>) -> Self {
        OutputHandle { sender }
    }

    /// Subscribe to receive events from the engine
    /// Returns a receiver that can be used to get events
    /// Useful for outputs that need multiple receivers (e.g., WebSocket with multiple clients)
    pub fn subscribe(&self) -> broadcast::Receiver<String> {
        self.sender.subscribe()
    }

    /// Receive an event from the engine (async)
    /// Creates a new subscription and receives one event
    /// For multiple receives, use subscribe() once and reuse the receiver
    pub async fn recv(&self) -> Result<String, RecvError> {
        let mut rx = self.subscribe();
        rx.recv().await
    }

    /// Receive an event from the engine (blocking)
    /// Creates a new subscription and receives one event
    /// For multiple receives, use subscribe() once and reuse the receiver
    pub fn blocking_recv(&self) -> Result<String, RecvError> {
        let mut rx = self.subscribe();
        rx.blocking_recv()
    }
}

/// Trait that all outputs must implement
/// Outputs can ONLY receive events from the engine, never send
pub trait Output: Send {
    /// Start the output with a handle to receive events from the engine
    fn start(self: Box<Self>, handle: OutputHandle);
}
