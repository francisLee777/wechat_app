import { request } from '../utils/request';
import type { Message } from '../models/index';

export async function addMessage(content: string, parentId: string | null = null, replyToUserId?: string, replyToUserName?: string): Promise<Message> {
  const data: any = { content };
  if (parentId) data.parentId = parseInt(parentId);
  if (replyToUserId) data.replyToUserId = replyToUserId;
  if (replyToUserName) data.replyToUserName = replyToUserName;

  const res = await request<any>('/message/add', 'POST', data);
  return mapMessage(res);
}

export async function getMessages(): Promise<Message[]> {
  const res = await request<any[]>('/message/list', 'GET');
  return res.map(mapMessage);
}

function mapMessage(m: any): Message {
  return {
    id: String(m.id),
    familyId: String(m.family_id ?? m.familyId),
    userId: m.user_id ?? m.userId,
    userName: m.user_name ?? m.userName,
    userAvatar: m.user_avatar ?? m.userAvatar,
    content: m.content,
    createTime: new Date(m.create_time ?? m.createTime).getTime(),
    parentId: (m.parent_id ?? m.parentId) ? String(m.parent_id ?? m.parentId) : null,
    replyToUserId: m.reply_to_user_id ?? m.replyToUserId,
    replyToUserName: m.reply_to_user_name ?? m.replyToUserName,
    replies: m.replies ? m.replies.map(mapMessage) : []
  };
}

export async function deleteMessage(msgId: string): Promise<void> {
  await request(`/message/delete?id=${msgId}`, 'POST');
}
