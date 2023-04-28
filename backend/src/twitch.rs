struct TwitchClient {
    client: reqwest::Client,
    client_id: String,
    client_secret: String,
}

impl TwitchClient {
    pub fn new(client_id: String, client_secret: String) -> TwitchClient {
        TwitchClient {
            client: reqwest::Client::new(),
            client_id,
            client_secret,
        }
    }
}
