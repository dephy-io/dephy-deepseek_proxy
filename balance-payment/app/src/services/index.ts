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

interface SSEMessage {
    content: string;
}

interface ApiResponse<T> {
    data?: T;
    error?: string;
}

const BASE_URL = "/api";

export async function getUser(publicKey: string): Promise<ApiResponse<User>> {
    try {
        const response = await fetch(`${BASE_URL}/user?user_pubkey=${encodeURIComponent(publicKey)}`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
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
export async function createConversation(publicKey: string): Promise<ApiResponse<Conversation>> {
    try {
        const response = await fetch(`${BASE_URL}/conversations`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ user_pubkey: publicKey }),
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
export async function getConversations(publicKey: string): Promise<ApiResponse<Conversation[]>> {
    try {
        const response = await fetch(`${BASE_URL}/conversations?user_pubkey=${encodeURIComponent(publicKey)}`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
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
    handler: (content: string) => Promise<void> // 新增 UI 处理函数
): Promise<ApiResponse<Message>> {
    try {
        const response = await fetch(`${BASE_URL}/messages`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
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
            if(chunk.startsWith("event:done")){
                break
            }
            const sseMsgStr = chunk.split("event:message\ndata:")[1]
            const ssgMsg: SSEMessage = JSON.parse(sseMsgStr);
            const content = ssgMsg.content
            fullContent += content
            await handler(content)
        }

        return doneMessage ? { data: doneMessage } : { error: "No message returned" };
    } catch (error) {
        return { error: error instanceof Error ? error.message : "Network error" };
    }
}

// 获取消息列表 (GET /messages)
export async function getMessages(conversationId: string): Promise<ApiResponse<Message[]>> {
    try {
        const response = await fetch(`${BASE_URL}/messages?conversation_id=${encodeURIComponent(conversationId)}`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
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
