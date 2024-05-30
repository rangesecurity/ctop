use anyhow::anyhow;
use clap::{Arg, Command};
use ws_client::WsClient;
#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let matches = Command::new("ctop")
    .subcommands(vec![
        Command::new("subscribe")
        .arg(
            Arg::new("url")
            .long("url")
        )
    ]).get_matches();
    match matches.subcommand() {
        Some(("subscribe", sub)) => {
            let url = sub.get_one::<String>("url").unwrap();
            let client = WsClient::new(url).await?;
            client.subscribe_votes().await?;
            Ok(())
        }
        _ => Err(anyhow!("invalid subcommand"))
    }
}