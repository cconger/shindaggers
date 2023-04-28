use sqlx::MySqlPool;
use std::env;
use std::net::TcpListener;

pub struct TestApp {
    pub address: String,
    pub db_pool: MySqlPool,
}

#[tokio::test]
async fn health_check_works() {
    let app = spawn_app().await;

    let client = reqwest::Client::new();

    let response = client
        .get(&format!("{}/health", &app.address))
        .send()
        .await
        .expect("Failed to execute requst.");

    assert!(response.status().is_success());
}

async fn spawn_app() -> TestApp {
    let listener = TcpListener::bind("127.0.0.1:0").expect("Failed to bind to random port");

    let connection_pool = MySqlPool::connect(
        env::var("CONN_STRING")
            .expect("CONN_STRING Env Var required")
            .as_str(),
    )
    .await
    .expect("Failed to connect to mysql");

    let port = listener.local_addr().unwrap().port();
    let server =
        shindaggers::run(listener, connection_pool.clone()).expect("Failed to bind address");
    let _ = tokio::spawn(server);

    let address = format!("http://127.0.0.1:{}", port);
    TestApp {
        address,
        db_pool: connection_pool,
    }
}
