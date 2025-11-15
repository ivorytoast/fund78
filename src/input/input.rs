use crate::Event;
use std::sync::mpsc::{self, SendError};

/// Handle for sending events to the engine
/// Users interact with this abstraction, never seeing the underlying channel
pub struct InputHandle {
    sender: mpsc::Sender<Event>,
}

impl InputHandle {
    /// Create a new InputHandle (only Engine should call this)
    pub(crate) fn new(sender: mpsc::Sender<Event>) -> Self {
        InputHandle { sender }
    }

    /// Send an event to the engine
    /// This is the only way inputs can communicate with the engine
    pub fn send_event(&self, event: Event) -> Result<(), SendError<Event>> {
        self.sender.send(event)
    }
}

/// Trait that all inputs must implement
/// Inputs can ONLY send events to the engine, never receive
pub trait Input: Send {
    /// Start the input with a handle to send events to the engine
    fn start(self: Box<Self>, handle: InputHandle);
}
