use actix_web::{web, HttpRequest, HttpResponse};
use chrono::prelude::*;
use serde::Serialize;
use sqlx::MySqlPool;

#[derive(sqlx::FromRow, Serialize)]
struct User {
    id: i64,
    name: String,
    twitch_id: String,
    twitch_login: String,
    created_at: chrono::DateTime<Utc>,
    updated_at: chrono::DateTime<Utc>,
}

pub async fn get_user(req: HttpRequest, pool: web::Data<MySqlPool>) -> HttpResponse {
    let name = req.match_info().get("name").unwrap_or("ChandyMan");

    match sqlx::query_as::<_, User>("SELECT id, name, twitch_id, twitch_login, created_at, updated_at FROM users where lookup_name = ?")
        .bind(name)
        .fetch_one(pool.get_ref())
        .await
        {
            Ok(user) => {
                HttpResponse::Ok().json(user)
            },
            Err(e) => {
                match e {
                    sqlx::Error::RowNotFound => {
                        HttpResponse::NotFound().finish()
                    }
                    _ => {
                        HttpResponse::InternalServerError().finish()
                    }
                }
            }
        }
}
