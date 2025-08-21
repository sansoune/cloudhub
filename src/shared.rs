use serde::{Serialize, Deserialize};

#[derive(Serialize, Deserialize, Debug)]
pub enum Command {
    Up { services: Vec<String> },
    Down { services: Vec<String> },
    Restart { services: Vec<String> },
    List { verbose: bool },
    Status,
}

#[derive(Serialize, Deserialize, Debug)]
pub enum Response {
    Success(String),
    Error(String),
}