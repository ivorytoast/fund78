use crate::{Output, OutputHandle};
use futures_util::{SinkExt, StreamExt};
use tokio::net::{TcpListener, TcpStream};
use tokio_tungstenite::{accept_async, tungstenite::Message};

pub struct WebSocketOutput {
    bind_address: String,
}

impl WebSocketOutput {
    pub fn new(bind_address: String) -> Self {
        WebSocketOutput { bind_address }
    }
}

impl Output for WebSocketOutput {
    fn start(self: Box<Self>, handle: OutputHandle) {
        let rt = tokio::runtime::Runtime::new().unwrap();
        rt.block_on(async {
            let listener = TcpListener::bind(&self.bind_address).await.unwrap();
            println!("Websocket server running on ws://{}", self.bind_address);
            println!("Client will receive processed events from Engine");

            while let Ok((stream, _)) = listener.accept().await {
                let rx = handle.subscribe();
                tokio::spawn(async move {
                    handle_connection(stream, rx).await;
                });
            }
        });
    }
}

async fn handle_connection(stream: TcpStream, mut rx: tokio::sync::broadcast::Receiver<String>) {
    let ws_stream = match accept_async(stream).await {
        Ok(ws) => ws,
        Err(e) => {
            eprintln!("WebSocket handshake failed: {}", e);
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
