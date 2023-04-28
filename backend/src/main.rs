use shindaggers::run;
use sqlx::MySqlPool;
use std::env;
use std::net::TcpListener;

#[tokio::main]
async fn main() -> Result<(), std::io::Error> {
    let listener = TcpListener::bind("127.0.0.1:8000")?;

    let connection_pool = MySqlPool::connect(
        env::var("CONN_STRING")
            .expect("DSN Env Var required")
            .as_str(),
    )
    .await
    .expect("Failed to connect to mysql");

    {
        let _res = sqlx::query("SELECT 1;")
            .execute(&connection_pool)
            .await
            .expect("Failed to connect to mysql");
    }

    run(listener, connection_pool)?.await
}
