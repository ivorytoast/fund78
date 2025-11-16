use crate::{Event, Input, InputHandle};
use futures_util::{SinkExt, StreamExt};
use serde::Deserialize;
use tokio_tungstenite::{connect_async, tungstenite::Message};

#[derive(Debug, Deserialize)]
struct CryptoTrade {
    #[serde(rename = "ev")]
    _ev: String,
    #[serde(rename = "pair")]
    _pair: String,
    p: f64,
    t: i64,
    s: f64,
    #[serde(rename = "c")]
    _c: Vec<i32>,
    #[serde(rename = "i")]
    _i: String,
    #[serde(rename = "x")]
    _x: i32,
    #[serde(rename = "r")]
    _r: i64,
}

pub struct PolygonInput {
    api_key: String,
}

impl PolygonInput {
    pub fn new(api_key: String) -> Self {
        PolygonInput { api_key }
    }
}

impl Input for PolygonInput {
    fn start(self: Box<Self>, handle: InputHandle) {
        let rt = tokio::runtime::Runtime::new().unwrap();
        rt.block_on(async {
            let url = format!("wss://socket.massive.com/crypto");

            let (ws_stream, _) = connect_async(&url)
                .await
                .expect("Failed to connect to Polygon");
            println!("Connected to Polygon.io");

            let (mut write, mut read) = ws_stream.split();

            let auth_msg = format!(r#"{{"action":"auth","params":"{}"}}"#, self.api_key);
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
                                    // println!(
                                    //     "Bitcoin trade: price=${}, size={}, time={}",
                                    //     trade.p, trade.s, trade.t
                                    // );

                                    let event = Event {
                                        task: "bitcoin_price".to_string(),
                                        payload: trade.p as i32,
                                    };

                                    if handle.send_event(event).is_err() {
                                        eprintln!(
                                            "Failed to send event - engine thread may have stopped"
                                        );
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
        });
    }
}
