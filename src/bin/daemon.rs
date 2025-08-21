use anyhow::Result;
use tokio::net::UnixListener;
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use std::path::Path;

#[tokio::main]
async fn main() -> Result<()> {
    let socket_path = "/run/cloudhub.sock";

    if Path::new(socket_path).exists() {
        std::fs::remove_file(socket_path)?;
    }

    let listener = UnixListener::bind(socket_path)?;
    println!("Daemon running on {}", socket_path);

    loop {
        let (mut socket, _) = listener.accept().await?;
        let mut buffer = [0; 1024];

        match socket.read(&mut buffer).await {
            Ok(0) => break, // Connection closed
            Ok(n) => {
                let request = String::from_utf8_lossy(&buffer[..n]);
                println!("Received request: {}", request);

                // Here you would handle the request and send a response
                let response = "Response from daemon";
                socket.write_all(response.as_bytes()).await?;
            }
            Err(e) => {
                eprintln!("Failed to read from socket: {}", e);
            }
        }
    }
    Ok(())
}