use anyhow::Result;
use clap::{Parser, Subcommand};
use cloudhub::shared::{Command, Response};
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use tokio::net::UnixStream;
use std::path::Path;

#[derive(Parser)]
#[command(
    name = "client",
    about = "CloudHub Client - Interact with the CloudHub daemon"
)]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    Up { services: Vec<String> },
    Down { services: Vec<String> },
    Restart { services: Vec<String> },
    List { verbose: bool },
    Status,
}

#[tokio::main]
async fn main() -> Result<()> {
    let cli = Cli::parse();
    let socket_path = "/run/cloudhub.sock";

    if !Path::new(socket_path).exists() {
        eprintln!("Daemon socket not found at {}", socket_path);
        std::process::exit(1);
    }

    let mut stream = UnixStream::connect(socket_path).await?;
    println!("Connected to CloudHub daemon at {}", socket_path);

    let command = match cli.command {
        Commands::Up { services } => Command::Up { services },
        Commands::Down { services } => Command::Down { services },
        Commands::Restart { services } => Command::Restart { services },
        Commands::List { verbose } => Command::List { verbose },
        Commands::Status => Command::Status,
    };
    let cmd_bytes = serde_json::to_vec(&command).expect("Failed to serialize command");
    stream.write_all(&cmd_bytes).await?;

    let mut buffer = Vec::new();
    stream.read_to_end(&mut buffer).await?;
    
    if buffer.is_empty() {
        println!("No response from daemon");
    } else {
        let response: Response = serde_json::from_slice(&buffer)?;
        match response {
            Response::Success(msg) => println!("✅ {}", msg),
            Response::Error(msg) => eprintln!("❌ {}", msg),
        }
    }

    Ok(())
}
