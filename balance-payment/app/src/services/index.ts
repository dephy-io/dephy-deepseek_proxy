export interface Conversation {
    id: string;
    user_pubkey: string;
    total_tokens: number;
    created_at: string;
    updated_at: string;
}

export interface User {
    id: number;
    public_key: string;
    tokens: number;
    tokens_consumed: number;
    created_at: string;
    updated_at: string;
}

export interface Message {
    id: number;
    conversation_id: string;
    role: string;
    content: string;
    created_at: string;
}

interface LoginResponse {
    token: string;
    user: string;
    expire_at: string; // ISO 8601 格式
}

interface SSEMessage {
    content: string;
}

interface ApiResponse<T> {
    data?: T;
    error?: string;
}

const BASE_URL = "/api";

const AUTH_DATA_KEY = "auth_data";

export const getAuthData = (): LoginResponse | null => {
    const authData = localStorage.getItem(AUTH_DATA_KEY);
    if (!authData) return null;

    const data = JSON.parse(authData) as LoginResponse;
    const expireAt = new Date(data.expire_at);
    if (isNaN(expireAt.getTime()) || expireAt <= new Date()) {
        // Token 已过期，清除本地存储
        localStorage.removeItem(AUTH_DATA_KEY);
        return null;
    }

    return data;
};

const setAuthData = (data: LoginResponse) => {
    localStorage.setItem(AUTH_DATA_KEY, JSON.stringify(data));
};

export async function login(
    publicKey: string,
    message: string,
    signature: string
): Promise<ApiResponse<{ token: string; user: User }>> {
    try {
        const response = await fetch(`${BASE_URL}/user/login`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ user_pubkey: publicKey, message, signature }),
        });
        const data = await response.json();
        if (!response.ok) {
            return { error: data.error || "Failed to login" };
        }
        setAuthData(data as LoginResponse);
        return { data };
    } catch (error) {
        return { error: error instanceof Error ? error.message : "Network error" };
    }
}

// 获取用户 (GET /user)
export async function getUser(): Promise<ApiResponse<User>> {
    try {
        const authData = getAuthData();
        if (!authData) {
            return { error: "No valid authentication data found, please login again" };
        }

        const response = await fetch(`${BASE_URL}/user`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${authData.token}`,
            },
        });
        const data = await response.json();
        if (!response.ok) {
            return { error: data.error || "Failed to get user" };
        }
        return { data };
    } catch (error) {
        return { error: error instanceof Error ? error.message : "Network error" };
    }
}

// 创建会话 (POST /conversations)
export async function createConversation(): Promise<ApiResponse<Conversation>> {
    try {
        const authData = getAuthData();
        if (!authData) {
            return { error: "No valid authentication data found, please login again" };
        }

        const response = await fetch(`${BASE_URL}/conversations`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${authData.token}`,
            },
            // 假设后端从 token 中提取 user_pubkey
            body: JSON.stringify({}),
        });
        const data = await response.json();
        if (!response.ok) {
            return { error: data.error || "Failed to create conversation" };
        }
        return { data };
    } catch (error) {
        return { error: error instanceof Error ? error.message : "Network error" };
    }
}

// 获取会话列表 (GET /conversations)
export async function getConversations(): Promise<ApiResponse<Conversation[]>> {
    try {
        const authData = getAuthData();
        if (!authData) {
            return { error: "No valid authentication data found, please login again" };
        }

        const response = await fetch(`${BASE_URL}/conversations`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${authData.token}`,
            },
        });
        const data = await response.json();
        if (!response.ok) {
            return { error: data.error || "Failed to get conversations" };
        }
        return { data };
    } catch (error) {
        return { error: error instanceof Error ? error.message : "Network error" };
    }
}

// 添加消息 (POST /messages)，支持自定义 UI 处理函数
export async function addMessage(
    conversationId: string,
    content: string,
    model: string,
    handler: (content: string) => Promise<void>
): Promise<ApiResponse<Message>> {
    try {
        const authData = getAuthData();
        if (!authData) {
            return { error: "No valid authentication data found, please login again" };
        }

        const response = await fetch(`${BASE_URL}/messages`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${authData.token}`,
            },
            body: JSON.stringify({
                conversation_id: conversationId,
                content,
                model,
            }),
        });

        if (!response.ok) {
            const data = await response.json();
            return { error: data.error || "Failed to add message" };
        }

        const reader = response.body?.getReader();
        if (!reader) {
            return { error: "Failed to read stream" };
        }

        const decoder = new TextDecoder();
        let fullContent = "";
        let doneMessage: Message | undefined;

        while (true) {
            const { done, value } = await reader.read();
            if (done) break;

            const chunk = decoder.decode(value);
            if (chunk.startsWith("event:done")) {
                break;
            }
            const sseMsgStr = chunk.split("event:message\ndata:")[1];
            const sseMsg: SSEMessage = JSON.parse(sseMsgStr);
            const content = sseMsg.content;
            fullContent += content;
            await handler(content);
        }

        return doneMessage ? { data: doneMessage } : { error: "No message returned" };
    } catch (error) {
        return { error: error instanceof Error ? error.message : "Network error" };
    }
}

// 获取消息列表 (GET /messages)
export async function getMessages(conversationId: string): Promise<ApiResponse<Message[]>> {
    try {
        const authData = getAuthData();
        if (!authData) {
            return { error: "No valid authentication data found, please login again" };
        }

        const response = await fetch(`${BASE_URL}/messages?conversation_id=${encodeURIComponent(conversationId)}`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${authData.token}`,
            },
        });
        const data = await response.json();
        if (!response.ok) {
            return { error: data.error || "Failed to get messages" };
        }
        return { data };
    } catch (error) {
        return { error: error instanceof Error ? error.message : "Network error" };
    }
}