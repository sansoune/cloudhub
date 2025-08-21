use tokio::net::UnixStream;
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use anyhow::Result;
// use serde_json;
#[tokio::main]
async fn main() -> Result<()> {
    let socket_path = "/run/cloudhub.sock";

    let mut stream = UnixStream::connect(socket_path).await?;
    println!("Connected to CloudHub daemon at {}", socket_path);


    let command = "status";
    stream.write_all(command.as_bytes()).await?;

    let mut buffer = [0; 1024];
    let n = stream.read(&mut buffer).await?;
    if n == 0 {
        println!("No response from daemon.");
        return Ok(());
    }

    let response = String::from_utf8_lossy(&buffer[..n]);
    println!("Received response: {}", response);

    Ok(())
    
}