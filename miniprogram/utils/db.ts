// miniprogram/utils/db.ts

// 配置后端服务地址
const BASE_URL = 'http://localhost:8080/api';

// ================= 类型定义 =================

export type UserRole = 'admin' | 'member';

export interface User {
  id: string;          // OpenID
  nickName: string;
  avatarUrl: string;
  familyId: string | null;
  role: UserRole;
}

export interface Family {
  id: string;
  name: string;
  ownerId: string;
  createTime: number;
}

export interface Recipe {
  id: string;
  familyId: string;
  name: string;
  images: string[];
  content: string;
  createTime: number;
  updateTime: number;
  sortOrder: number;
}

export interface Message {
  id: string;
  familyId: string;
  userId: string;
  userName: string;
  userAvatar: string;
  content: string;
  createTime: number;
  parentId: string | null;
  replyToUserId?: string;
  replyToUserName?: string;
  replies?: Message[];
}

// ================= 网络请求封装 =================

const request = <T>(
  url: string, 
  method: 'GET' | 'POST' | 'PUT' | 'DELETE', 
  data: any = {},
  headers: any = {}
): Promise<T> => {
  return new Promise((resolve, reject) => {
    const openId = wx.getStorageSync('OPENID');
    const header = {
      ...headers,
      'Content-Type': 'application/json',
      ...(openId ? { 'X-WX-OPENID': openId } : {})
    };

    wx.request({
      url: `${BASE_URL}${url}`,
      method,
      data,
      header,
      success: (res: any) => {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve(res.data as T);
        } else {
          reject(new Error(res.data?.error || `Request failed with status ${res.statusCode}`));
        }
      },
      fail: (err) => {
        reject(new Error(err.errMsg || 'Network request failed'));
      }
    });
  });
};

// ================= 用户认证 =================

let cachedUser: User | null = null;

function updateCachedUser(user: User | null) {
  cachedUser = user;
  if (user) {
    wx.setStorageSync('CACHED_USER', user);
  } else {
    wx.removeStorageSync('CACHED_USER');
  }
}

// 同步获取当前用户（优先返回内存/缓存）
export function getCurrentUser(): User | null {
  if (!cachedUser) {
    const stored = wx.getStorageSync('CACHED_USER');
    if (stored) cachedUser = stored;
  }
  return cachedUser;
}

// 异步从服务器获取最新用户信息并更新缓存
export async function fetchCurrentUser(): Promise<User | null> {
  const openId = wx.getStorageSync('OPENID');
  if (!openId) return null;
  try {
    const dbUser = await request<any>('/user/getUserInfo', 'GET');
    if (!dbUser || !dbUser.openId) return null;
    const user: User = {
      id: dbUser.openId,
      nickName: dbUser.user_nickName,
      avatarUrl: dbUser.user_icon_url,
      familyId: dbUser.familyId ? String(dbUser.familyId) : null,
      role: dbUser.role as UserRole
    };
    updateCachedUser(user);
    return user;
  } catch (e) {
    console.error('fetchCurrentUser failed', e);
    return null;
  }
}

export function checkSession(): Promise<boolean> {
  return new Promise((resolve) => {
    wx.checkSession({
      success: () => {
        const openId = wx.getStorageSync('OPENID');
        resolve(!!openId);
      },
      fail: () => {
        resolve(false);
      }
    });
  });
}

export async function loginUser(): Promise<User> {
  return new Promise((resolve, reject) => {
    wx.login({
      success: async (res) => {
        if (res.code) {
          try {
            const loginRes = await request<{openid: string}>(`/user/login?code=${res.code}`, 'POST');
            wx.setStorageSync('OPENID', loginRes.openid);
            
            const user = await fetchCurrentUser();
            if (user) {
              resolve(user);
            } else {
              reject(new Error('Failed to retrieve user info'));
            }
          } catch (e: any) {
            reject(new Error('Login failed: ' + e.message));
          }
        } else {
          reject(new Error('wx.login failed: ' + res.errMsg));
        }
      },
      fail: (err) => reject(err)
    });
  });
}

// ================= 家庭相关 =================

export async function createFamily(name: string): Promise<Family> {
  const res = await request<any>('/family/create', 'POST', { name });
  const family = {
    id: String(res.id),
    name: res.name,
    ownerId: res.ownerId,
    createTime: new Date(res.createTime).getTime()
  };
  await fetchCurrentUser(); // Refresh role/familyId
  return family;
}

export async function joinFamily(familyId: string): Promise<Family> {
  const res = await request<any>('/family/join', 'POST', { familyId: parseInt(familyId) });
  const family = {
    id: String(res.id),
    name: res.name,
    ownerId: res.ownerId,
    createTime: new Date(res.createTime).getTime()
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
    familyId: String(u.familyId),
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

export async function getFamilyById(id: string): Promise<Family | undefined> {
  try {
    const res = await request<any>(`/family/info?id=${id}`, 'GET');
    return {
      id: String(res.id),
      name: res.name,
      ownerId: res.ownerId,
      createTime: new Date(res.createTime).getTime()
    };
  } catch (e) {
    return undefined;
  }
}

// ================= 食谱相关 =================

export async function getRecipes(keyword: string = ''): Promise<Recipe[]> {
  const res = await request<any[]>('/recipe/list', 'GET');
  let recipes = res.map(r => ({
    id: String(r.id),
    familyId: String(r.familyId),
    name: r.name,
    images: r.images, 
    content: r.content,
    createTime: r.createTime, 
    updateTime: r.updateTime,
    sortOrder: r.sortOrder
  }));

  if (keyword) {
    recipes = recipes.filter(r => r.name.includes(keyword) || r.content.includes(keyword));
  }
  return recipes;
}

export async function getRecipeById(id: string): Promise<Recipe | undefined> {
  try {
    const res = await request<any>(`/recipe/info?id=${id}`, 'GET');
    return {
      id: String(res.id),
      familyId: String(res.familyId),
      name: res.name,
      images: res.images,
      content: res.content,
      createTime: res.createTime,
      updateTime: res.updateTime,
      sortOrder: res.sortOrder
    };
  } catch (e) {
    return undefined;
  }
}

export async function addRecipe(name: string, images: string[], content: string): Promise<Recipe> {
  const res = await request<any>('/recipe/add', 'POST', { name, images, content });
  return {
    id: String(res.id),
    familyId: String(res.familyId),
    name: res.name,
    images: res.images,
    content: res.content,
    createTime: res.createTime,
    updateTime: res.updateTime,
    sortOrder: res.sortOrder
  };
}

export async function updateRecipe(id: string, name: string, images: string[], content: string): Promise<void> {
  await request('/recipe/update', 'POST', { id: parseInt(id), name, images, content });
}

export async function deleteRecipe(id: string): Promise<void> {
  await request(`/recipe/delete?id=${id}`, 'POST');
}

export async function reorderRecipe(id: string, direction: 'up' | 'down'): Promise<void> {
  await request('/recipe/reorder', 'POST', { id: parseInt(id), direction });
}

export async function batchUpdateRecipes(updatedRecipes: Recipe[]): Promise<void> {
  const payload = updatedRecipes.map(r => ({
    id: parseInt(r.id),
    sortOrder: r.sortOrder
  }));
  await request('/recipe/batchUpdate', 'POST', { recipes: payload });
}

// ================= 留言板相关 =================

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
    familyId: String(m.familyId),
    userId: m.userId,
    userName: m.userName,
    userAvatar: m.userAvatar,
    content: m.content,
    createTime: new Date(m.createTime).getTime(),
    parentId: m.parentId ? String(m.parentId) : null,
    replyToUserId: m.replyToUserId,
    replyToUserName: m.replyToUserName,
    replies: m.replies ? m.replies.map(mapMessage) : []
  };
}

export async function deleteMessage(msgId: string): Promise<void> {
  await request(`/message/delete?id=${msgId}`, 'POST');
}
