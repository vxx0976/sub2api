import { apiClient } from '../client'
import type { ChatConversation, ChatMessage } from '@/types'

export async function listConversations(
  page = 1,
  pageSize = 20,
  filters?: { status?: string; search?: string }
) {
  const { data } = await apiClient.get('/admin/chat/conversations', {
    params: { page, page_size: pageSize, ...filters },
  })
  return data
}

export async function getConversation(id: number): Promise<ChatConversation> {
  const { data } = await apiClient.get(`/admin/chat/conversations/${id}`)
  return data
}

export async function getMessages(
  id: number,
  limit = 50,
  offset?: number
): Promise<{ messages: ChatMessage[]; total: number }> {
  const { data } = await apiClient.get(`/admin/chat/conversations/${id}/messages`, {
    params: { limit, ...(offset === undefined ? {} : { offset }) },
  })
  return data
}

export async function sendReply(id: number, content: string): Promise<ChatMessage> {
  const { data } = await apiClient.post(`/admin/chat/conversations/${id}/messages`, { content })
  return data
}

export async function closeConversation(id: number): Promise<void> {
  await apiClient.post(`/admin/chat/conversations/${id}/close`)
}

export async function markRead(id: number): Promise<void> {
  await apiClient.post(`/admin/chat/conversations/${id}/read`)
}

export async function getUnreadCount(): Promise<{ count: number }> {
  const { data } = await apiClient.get('/admin/chat/unread-count')
  return data
}

export default {
  listConversations,
  getConversation,
  getMessages,
  sendReply,
  closeConversation,
  markRead,
  getUnreadCount,
}
