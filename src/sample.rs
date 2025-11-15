use crate::{Event, Worker};

use std::collections::VecDeque;

pub fn sample_events() -> VecDeque<Event> {
    let mut events: VecDeque<Event> = VecDeque::new();
    events.push_back(Event {
        task: String::from("price"),
        payload: 10,
    });
    events.push_back(Event {
        task: String::from("price"),
        payload: 20,
    });
    events.push_back(Event {
        task: String::from("price"),
        payload: 30,
    });
    return events;
}

pub fn sample_workers() -> Vec<Worker> {
    let mut workers: Vec<Worker> = Vec::new();
    workers.push(Worker {
        handles_task: String::from("price"),
        job: price_job,
    });
    return workers;
}

fn price_job(payload: i32) -> Event {
    println!("Got the price of {}", payload);
    Event {
        task: String::from("price"),
        payload: payload + 5,
    }
}
