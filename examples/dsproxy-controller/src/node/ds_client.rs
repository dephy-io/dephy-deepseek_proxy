use reqwest::Client;
use serde::{Deserialize, Serialize};
use std::{collections::HashMap, error::Error};

const API_BASE_URL: &str = "https://api.ppinfra.com/v3/openai";

#[derive(Debug, Serialize, Deserialize)]
pub struct AskMessage {
    pub role: String,
    pub content: Option<String>,
    pub name: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct AnwserMessage {
    pub role: String,
    pub content: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum ResponseFormat {
    Text,
    JsonObject,
    JsonSchema { schema: serde_json::Value },
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ChatCompletionRequest {
    pub model: String, 
    pub messages: Vec<AskMessage>, 
    pub max_tokens: u32, 

    pub temperature: Option<f32>, 
    pub top_p: Option<f32>, 
    pub top_k: Option<u32>, 
    pub min_p: Option<f32>, 
    pub n: Option<u32>, 
    pub stream: Option<bool>, 
    pub stop: Option<Vec<String>>, 
    pub presence_penalty: Option<f32>, 
    pub frequency_penalty: Option<f32>, 
    pub repetition_penalty: Option<f32>, 
    pub logit_bias: Option<HashMap<String, i32>>, 
    pub logprobs: Option<bool>, 
    pub top_logprobs: Option<u32>, 

    pub response_format: Option<ResponseFormat>, 
    pub seed: Option<u32>, 
    pub user: Option<String>, 
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ChatChoice {
    pub index: u32,
    pub message: AnwserMessage,
    pub finish_reason: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ChatCompletionResponse {
    pub id: String,
    pub object: String,
    pub created: u64,
    pub model: String,
    pub choices: Vec<ChatChoice>,
    pub usage: Option<Usage>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Usage {
    pub prompt_tokens: u32,
    pub completion_tokens: u32,
    pub total_tokens: u32,
}

pub struct DsClient {
    client: Client,
    api_key: String,
}

impl DsClient {
    pub fn new(api_key: String) -> Self {
        Self {
            client: Client::new(),
            api_key,
        }
    }

    async fn post<T: Serialize + ?Sized, U: for<'de> Deserialize<'de>>(
        &self,
        endpoint: &str,
        body: &T,
    ) -> Result<U, Box<dyn Error>> {
        let url = format!("{}/{}", API_BASE_URL, endpoint);
        let res = self
            .client
            .post(&url)
            .header("Authorization", format!("Bearer {}", self.api_key))
            .header("Content-Type", "application/json")
            .json(body)
            .send()
            .await?;

        if res.status().is_success() {
            let response_body = res.json::<U>().await?;
            Ok(response_body)
        } else {
            let status = res.status();
            let error_text = res.text().await?;
            Err(format!("Request failed with status {}: {}", status, error_text).into())
        }
    }

    pub async fn create_chat_completion(
        &self,
        request: ChatCompletionRequest,
    ) -> Result<ChatCompletionResponse, Box<dyn Error>> {
        self.post("chat/completions", &request).await
    }

    // Add more methods as needed for other endpoints
}
