import { request, toRelativePath } from '../utils/request';
import type { User, UserRole } from '../models/index';

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
  const token = wx.getStorageSync('AUTH_TOKEN');
  if (!token) return null;
  try {
    const dbUser = await request<any>('/user/getUserInfo', 'GET');
    if (!dbUser || !dbUser.openId) return null;
    const user: User = {
      id: dbUser.openId,
      nickName: dbUser.user_nickName,
      avatarUrl: dbUser.user_icon_url,
      familyId: dbUser.family_id ? String(dbUser.family_id) : null,
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
        const token = wx.getStorageSync('AUTH_TOKEN');
        resolve(!!token);
      },
      fail: () => {
        resolve(false);
      }
    });
  });
}

export interface WechatUserInfo {
  nickName: string;
  avatarUrl: string;
}

export async function loginUser(): Promise<User> {
  return new Promise((resolve, reject) => {
    wx.login({
      success: async (res) => {
        if (res.code) {
          try {
            const payload = { code: res.code };
            console.log(payload)
            // 将 code 同时放入 URL 参数中，确保后端能获取到
            const loginRes = await request<{token: string, openid: string}>(`/user/login?code=${res.code}`, 'POST', payload);
            wx.setStorageSync('AUTH_TOKEN', loginRes.token);
            wx.setStorageSync('OPENID', loginRes.openid);
            console.log("登录后成功设置, AUTH_TOKEN ",loginRes.token,"和OPENID",loginRes.openid)
            
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

export async function saveNickName(nickName: string): Promise<void> {
  await request('/user/saveNickName', 'POST', { nickName });
  await fetchCurrentUser();
}

export async function saveAvatarUrl(avatarUrl: string): Promise<void> {
  const rel = toRelativePath(avatarUrl);
  await request(`/user/saveIconURL?iconURL=${encodeURIComponent(rel)}`, 'POST');
  await fetchCurrentUser();
}
