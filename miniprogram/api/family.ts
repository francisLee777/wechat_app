import { request } from '../utils/request';
import type { Family, User, UserRole } from '../models/index';
import { fetchCurrentUser } from './auth';

export async function createFamily(name: string): Promise<Family> {
  const res = await request<any>('/family/create', 'POST', { name });
  const family = {
    id: String(res.id),
    name: res.name,
    ownerId: res.owner_id,
    createTime: new Date(res.create_time).getTime()
  };
  await fetchCurrentUser(); // Refresh role/familyId
  return family;
}

export async function joinFamily(familyId: string): Promise<Family> {
  const res = await request<any>('/family/join', 'POST', { familyId: parseInt(familyId) });
  const family = {
    id: String(res.id),
    name: res.name,
    ownerId: res.owner_id,
    createTime: new Date(res.create_time).getTime()
  };
  await fetchCurrentUser();
  return family;
}

export async function getFamilyMembers(familyId: string): Promise<User[]> {
  const res = await request<any[]>('/family/members', 'GET');
  return res.map(u => ({
    id: u.openId,
    nickName: u.user_nickName,
    avatarUrl: u.user_icon_url,
    familyId: String(u.family_id),
    role: u.role as UserRole
  }));
}

export async function quitFamily(): Promise<void> {
  await request('/family/quit', 'POST');
  await fetchCurrentUser();
}

export async function removeMember(targetUserId: string): Promise<void> {
  await request('/family/removeMember', 'POST', { memberOpenId: targetUserId });
}

export async function deleteFamily(): Promise<void> {
  await request('/family/delete', 'POST');
  await fetchCurrentUser();
}

export async function getFamilyById(id: string): Promise<Family | undefined> {
  try {
    const res = await request<any>(`/family/info?id=${id}`, 'GET');
    return {
      id: String(res.id),
      name: res.name,
      ownerId: res.owner_id,
      createTime: new Date(res.create_time).getTime()
    };
  } catch (e) {
    return undefined;
  }
}
