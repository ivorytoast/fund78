use ::serde::Deserialize;
use ::std::sync::Arc;
use dotenv::dotenv;
use fund78::{Engine, Event, Worker};
use futures_util::{SinkExt, StreamExt};
use std::collections::VecDeque;
use std::env;
use std::sync::mpsc;
use std::thread;
use tokio::net::{TcpListener, TcpStream};
use tokio::sync::broadcast;
use tokio_tungstenite::connect_async;
use tokio_tungstenite::{accept_async, tungstenite::Message};

#[derive(Debug, Deserialize)]
struct CryptoTrade {
    ev: String,
    pair: String,
    p: f64,
    t: i64,
    s: f64,
    c: Vec<i32>,
    i: String,
    x: i32,
    r: i64,
}

fn main() {
    dotenv().ok();

    let (polygon_tx, engine_rx) = mpsc::channel::<Event>();

    let (engine_tx, _) = broadcast::channel::<String>(100);
    let engine_tx = Arc::new(engine_tx);

    let polygon_handle = thread::spawn(move || {
        let rt = tokio::runtime::Runtime::new().unwrap();
        rt.block_on(get_market_data(polygon_tx));
    });

    let engine_tx_clone = engine_tx.clone();
    let engine_handle = thread::spawn(move || {
        run_engine(engine_rx, engine_tx_clone);
    });

    let ws_handle = thread::spawn(|| {
        let rt = tokio::runtime::Runtime::new().unwrap();
        rt.block_on(run_websocket_server(engine_tx));
    });

    println!("All systems running...");
    println!("Data flow: Polygon -> Engine -> WebSocket Clients");

    polygon_handle.join().unwrap();
    engine_handle.join().unwrap();
    ws_handle.join().unwrap();
}

fn run_engine(engine_rx: mpsc::Receiver<Event>, broadcast_tx: Arc<broadcast::Sender<String>>) {
    let mut workers: Vec<Worker> = Vec::new();
    workers.push(Worker {
        handles_task: "bitcoin_price".to_string(),
        job: handle_bitcoin_price,
    });

    let mut engine = Engine::new(workers).expect("Failed to create engine");

    let mut pending_events = VecDeque::new();

    loop {
        match engine_rx.recv() {
            Ok(event) => {
                pending_events.push_back(event.clone());

                while let Ok(event) = engine_rx.try_recv() {
                    pending_events.push_back(event);
                }

                engine.process(pending_events.clone());

                for event in &pending_events {
                    let json = serde_json::to_string(&event).unwrap();
                    let _ = broadcast_tx.send(json);
                }

                pending_events.clear();
            }
            Err(_) => {
                println!("Polygon channel closed, shutting down engine");
                break;
            }
        }
    }
}

fn handle_bitcoin_price(payload: i32) -> Event {
    let timestamp = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_secs();

    Event {
        task: "bitcoin_price_accepted".to_string(),
        payload: format!("bitcoin price accepted at time: {}", timestamp)
            .parse()
            .unwrap_or(0),
    }
}

async fn get_market_data(event_tx: mpsc::Sender<Event>) {
    let api_key = env::var("POLYGON_API_KEY").expect("POLYGON_API_KEY must be set in .env file");
    let url = format!("wss://socket.massive.com/crypto");

    let (ws_stream, _) = connect_async(&url)
        .await
        .expect("Failed to connect to Polygon");
    println!("Connected to Polygon.io");

    let (mut write, mut read) = ws_stream.split();

    let auth_msg = format!(r#"{{"action":"auth","params":"{}"}}"#, api_key);
    write.send(Message::Text(auth_msg)).await.unwrap();
    println!("Sent authentication");

    if let Some(Ok(Message::Text(response))) = read.next().await {
        println!("Auth response: {}", response);
    }

    let subscribe_msg = r#"{"action":"subscribe","params":"XT.X:BTC-USD"}"#;
    write
        .send(Message::Text(subscribe_msg.to_string()))
        .await
        .unwrap();

    println!("Sent the subscribe message");

    while let Some(Ok(msg)) = read.next().await {
        if let Message::Text(text) = msg {
            if text.contains("\"ev\":\"XT\"") {
                match serde_json::from_str::<Vec<CryptoTrade>>(&text) {
                    Ok(trades) => {
                        for trade in trades {
                            println!(
                                "Bitcoin trade: price=${}, size={}, time={}",
                                trade.p, trade.s, trade.t
                            );

                            let event = Event {
                                task: "bitcoin_price".to_string(),
                                payload: trade.p as i32,
                            };

                            if event_tx.send(event).is_err() {
                                eprint!("Failed to send event - engine thread may have stopped");
                                return;
                            }
                        }
                    }
                    Err(e) => {
                        eprintln!("Failed to deserialize trade data: {}", e);
                    }
                }
            } else {
                println!("Received admin message from polygon: {}", text);
            }
        }
    }
}

async fn run_websocket_server(broadcast_tx: Arc<broadcast::Sender<String>>) {
    let listener = TcpListener::bind("127.0.0.1:8080").await.unwrap();
    println!("Websocket server running on ws://127.0.0.1:8080");
    println!("Client will receive processed events from Engine");

    while let Ok((stream, _)) = listener.accept().await {
        let rx = broadcast_tx.subscribe();
        tokio::spawn(handle_connection(stream, rx));
    }
}

async fn handle_connection(stream: TcpStream, mut rx: broadcast::Receiver<String>) {
    let ws_stream = match accept_async(stream).await {
        Ok(ws) => ws,
        Err(e) => {
            eprint!("WebSocket handshake failed: {}", e);
            return;
        }
    };

    let (mut write, _read) = ws_stream.split();
    println!("Client connected - will receive processed engine outputs");

    while let Ok(processed_event) = rx.recv().await {
        if write.send(Message::Text(processed_event)).await.is_err() {
            println!("Client disconnected");
            break;
        }
    }
}
