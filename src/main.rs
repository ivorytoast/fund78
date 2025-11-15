use dotenv::dotenv;
use fund78::{BitcoinPriceWorker, Engine, Input, Output, PolygonInput, WebSocketOutput, Worker};
use std::env;

fn main() {
    dotenv().ok();

    let workers: Vec<Box<dyn Worker>> = vec![Box::new(BitcoinPriceWorker::new())];

    let inputs: Vec<Box<dyn Input>> = vec![Box::new(PolygonInput::new(
        env::var("POLYGON_API_KEY").expect("POLYGON_API_KEY must be set in .env file"),
    ))];

    let outputs: Vec<Box<dyn Output>> =
        vec![Box::new(WebSocketOutput::new("127.0.0.1:8080".to_string()))];

    let engine = Engine::new(workers, inputs, outputs).expect("Failed to create engine");

    println!("Data flow: Polygon (Input) -> Engine -> WebSocket (Output)");

    engine.run();
}
