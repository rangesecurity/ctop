use {
    anyhow::Context, futures::StreamExt, std::sync::Arc, tendermint_rpc::{query::{EventType, Query}, Client, Error, SubscriptionClient, WebSocketClient}, tokio::task::JoinHandle
};

pub struct WsClient {
    c: WebSocketClient,
    driver_handle: JoinHandle<Result<(), Error>>
}

impl WsClient {
    pub async fn new(url: &str) -> anyhow::Result<Arc<Self>> {
        let (client, driver) = WebSocketClient::new(url).await?;
        let driver_handle = tokio::spawn(async move { driver.run().await });
        Ok(Arc::new(Self {
            c: client,
            driver_handle
        }))
    }
    pub async fn subscribe_votes(&self) -> anyhow::Result<()> {
        let query: Query = "tm.event='Vote'".parse()?;
        let mut sub = self.c.subscribe(query).await?;
        while let Some(res) = sub.next().await {
            println!("{res:#?}");
        }
        Ok(())
    }
    pub async fn unsubscribe_votes(&self) -> anyhow::Result<()> {
        let query: Query = "tm.event='Vote'".parse()?;
        self.c.unsubscribe(query).await.with_context(|| "failed to unsubscribe")
    }
}
