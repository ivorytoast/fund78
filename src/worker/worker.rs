use crate::Event;

/// Trait that all workers must implement
/// Workers are designed to handle events that match their specific topic/task
/// and execute a function that returns another event
pub trait Worker: Send + Sync {
    /// Returns the task/topic this worker handles (e.g., "bitcoin_price")
    /// Workers only process events where event.task matches this value
    fn handles_task(&self) -> &str;

    /// Process an event and return a new event
    /// This is the core logic of the worker - it receives an event matching
    /// its topic and produces a new event as output
    fn process(&self, event: Event) -> Event;
}

/// A simple function-based worker implementation
/// Useful for quickly creating workers from closures or function pointers
pub struct FunctionWorker {
    task: String,
    job: Box<dyn Fn(Event) -> Event + Send + Sync>,
}

impl FunctionWorker {
    /// Create a new FunctionWorker
    ///
    /// # Arguments
    /// * `task` - The task/topic this worker handles
    /// * `job` - The function that processes events
    pub fn new<F>(task: String, job: F) -> Self
    where
        F: Fn(Event) -> Event + Send + Sync + 'static,
    {
        FunctionWorker {
            task,
            job: Box::new(job),
        }
    }
}

impl Worker for FunctionWorker {
    fn handles_task(&self) -> &str {
        &self.task
    }

    fn process(&self, event: Event) -> Event {
        (self.job)(event)
    }
}
