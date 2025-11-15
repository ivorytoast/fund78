use fund78::engine_process;
use fund78::sample::{sample_events, sample_workers};

fn main() {
    let events = sample_events();
    let workers = sample_workers();

    engine_process(events, workers);
}
