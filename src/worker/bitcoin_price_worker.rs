use crate::{Event, Worker};

/// Worker that processes bitcoin price events
/// Handles events with task "bitcoin_price" and produces "bitcoin_price_accepted" events
pub struct BitcoinPriceWorker;

impl BitcoinPriceWorker {
    pub fn new() -> Self {
        BitcoinPriceWorker
    }
}

impl Default for BitcoinPriceWorker {
    fn default() -> Self {
        Self::new()
    }
}

impl Worker for BitcoinPriceWorker {
    fn handles_task(&self) -> &str {
        "bitcoin_price"
    }

    fn process(&self, event: Event) -> Event {
        let timestamp = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap()
            .as_secs();

        Event {
            task: "bitcoin_price_accepted".to_string(),
            payload: format!("bitcoin price {} accepted at time: {}", event.payload, timestamp)
                .parse()
                .unwrap_or(0),
        }
    }
}
