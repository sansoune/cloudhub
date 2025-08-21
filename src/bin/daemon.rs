use anyhow::{Result, Context};
use tokio::net::UnixListener;
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use std::path::{Path, PathBuf};
use cloudhub::shared::{Command, Response};
use cloudhub::service_manager::ServiceManager;
use serde_json::Result as JsonResult;

#[tokio::main]
async fn main() -> Result<()> {
    let path = PathBuf::from("/opt/cloudhub");
    let manager = ServiceManager::new(&path).await.context("Failed to initialize service manager")?;
    let socket_path = "/run/cloudhub.sock";

    if Path::new(socket_path).exists() {
        std::fs::remove_file(socket_path)?;
    }

    let listener = UnixListener::bind(socket_path)?;
    println!("Daemon running on {}", socket_path);

    loop {
        let (mut socket, _) = listener.accept().await?;
        let manager = manager.clone();
        tokio::spawn(async move {
            let mut buf = [0u8; 4096];
            if let Ok(n) = socket.read(&mut buf).await {
                if n == 0 {
                    return;
                }
                let cmd: JsonResult<Command> = serde_json::from_slice(&buf[..n]);
                let response = match cmd {
                    Ok(Command::Status) => {
                        // Response::Success("Service status is OK".to_string())
                        match manager.show_status(Vec::new()).await {
                            Ok(res) => Response::Success(res),
                            Err(e) => Response::Error(e.to_string())
                        }

                    },
                    Ok(Command::Up { services }) => {
                        Response::Success(format!("Started services: {:?}", services))
                    },
                    Ok(Command::Down { services }) => {
                        Response::Success(format!("Stopped services: {:?}", services))
                    },
                    Ok(Command::Restart { services }) => {
                        Response::Success(format!("Restarted services: {:?}", services))
                    },
                    Ok(Command::List { verbose }) => {
                        if verbose {
                            Response::Success("List of all services with details".to_string())
                        } else {
                            Response::Success("List of all services".to_string())
                        }
                    },
                    Err(e) => Response::Error(format!("Failed to parse command: {}", e)),
                };

                let resp_bytes = serde_json::to_vec(&response).expect("Failed to serialize response");
                let _ = socket.write_all(&resp_bytes).await;
            }
        });
    }
}