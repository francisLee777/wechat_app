import { compressImage } from './image-compress';

// 配置后端服务地址
// 模拟器中使用 localhost，真机调试需换成局域网IP或公网域名
export const BASE_URL = 'http://localhost/api';
// 图片资源的 Base URL (对应后端 Static 目录映射)
export const IMG_BASE_URL = 'http://localhost/api';

const handleUnauthorized = () => {
  wx.removeStorageSync('AUTH_TOKEN');
  // Avoid multiple modals
  // @ts-ignore
  if (getApp().globalData?.isShowingAuthModal) return;
  // @ts-ignore
  if (!getApp().globalData) getApp().globalData = {};
  // @ts-ignore
  getApp().globalData.isShowingAuthModal = true;

  wx.showModal({
    title: '登录过期',
    content: '您的登录状态已过期，请重新登录',
    showCancel: false,
    confirmText: '重新登录',
    success: () => {
      // @ts-ignore
      getApp().globalData.isShowingAuthModal = false;
      wx.reLaunch({ url: '/pages/profile/index' });
    }
  });
};

export const request = <T>(
  url: string, 
  method: 'GET' | 'POST' | 'PUT' | 'DELETE', 
  data: any = {},
  headers: any = {}
): Promise<T> => {
  return new Promise((resolve, reject) => {
    const token = wx.getStorageSync('AUTH_TOKEN');
    const header = {
      ...headers,
      'Content-Type': 'application/json',
      ...(token ? { 'Authorization': `Bearer ${token}` } : {})
    };

    wx.request({
      url: `${BASE_URL}${url}`,
      method,
      data,
      header,
      success: (res: any) => {
        // Check for token refresh
        const newToken = res.header['New-Token'] || res.header['new-token'];
        if (newToken) {
          wx.setStorageSync('AUTH_TOKEN', newToken);
          console.log('Token refreshed automatically (upload)');
        }

        if (res.statusCode === 401) {
          handleUnauthorized();
          reject(new Error('Unauthorized'));
          return;
        }

        if (res.statusCode >= 200 && res.statusCode < 300) {
          // Check backend business code
          if (res.data && typeof res.data.code === 'number') {
            if (res.data.code === 0) {
              resolve(res.data.data as T);
            } else {
              reject(new Error(res.data.errorMsg || res.data.errMsg || `Business Error: ${res.data.code}`));
            }
          } else {
            // Fallback for non-standard response or raw data (if any)
             resolve(res.data as T);
          }
          return;
        }
        var errorMsg = 'Request failed with status ' + res.statusCode;
        if (res && res.data && res.data.error) {
          errorMsg = res.data.error;
        }
        reject(new Error(errorMsg));
      },
      fail: (err) => {
        reject(new Error(err.errMsg || 'Network request failed'));
      }
    });
  });
};

// 将完整图片 URL 归一化为相对路径，形如 "/api/media/xxx"
export const toRelativePath = (url: string): string => {
  if (!url) return url;
  let rel = url;
  // 去掉 host 前缀
  if (rel.startsWith(IMG_BASE_URL)) {
    rel = rel.slice(IMG_BASE_URL.length);
  } else if (rel.startsWith('http://') || rel.startsWith('https://')) {
    // 粗略提取 path 部分
    const idx = rel.indexOf('://');
    if (idx >= 0) {
      const slash = rel.indexOf('/', idx + 3);
      if (slash >= 0) rel = rel.slice(slash);
    }
  }
  // 去掉查询参数
  const q = rel.indexOf('?');
  if (q >= 0) rel = rel.slice(0, q);
  // 确保以 / 开头
  if (rel && !rel.startsWith('/')) rel = '/' + rel;
  return rel;
};

export const uploadFile = (filePath: string, scope?: 'user' | 'family', shouldCompress: boolean = true): Promise<string> => {
  return new Promise(async (resolve, reject) => {
    let finalPath = filePath;
    
    // 尝试压缩
    if (shouldCompress) {
      try {
        finalPath = await compressImage(filePath);
      } catch (e) {
        console.warn('Compression failed, using original image', e);
      }
    }

    const token = wx.getStorageSync('AUTH_TOKEN');
    wx.uploadFile({
      url: `${BASE_URL}/upload${scope ? `?scope=${scope}` : ''}`,
      filePath: finalPath,
      name: 'file',
      header: {
        ...(token ? { 'Authorization': `Bearer ${token}` } : {})
      },
      success: (res) => {
        // Check for token refresh (uploadFile response also has header)
        // Note: wx.uploadFile response header might be inconsistent across platforms, but we try.
        // @ts-ignore
        const header = res.header || {};
        const newToken = header['New-Token'] || header['new-token'];
        if (newToken) {
          wx.setStorageSync('AUTH_TOKEN', newToken);
          console.log('Token refreshed automatically');
        }

        if (res.statusCode === 401) {
          handleUnauthorized();
          reject(new Error('Unauthorized'));
          return;
        }

        if (res.statusCode >= 200 && res.statusCode < 300) {
          try {
            // wx.uploadFile returns data as string
            const data = JSON.parse(res.data);
            if (data.code === 0 && data.data && data.data.url) {
              // 为图片访问追加 token 查询参数，适配 <image src> 无法设置 Header 的场景
              const token = wx.getStorageSync('AUTH_TOKEN');
              const base = IMG_BASE_URL + data.data.url;
              const sep = base.indexOf('?') >= 0 ? '&' : '?';
              console.log(base,token);
              resolve(token ? `${base}${sep}token=${token}` : base);
            } else {
              reject(new Error(data.errorMsg || 'Upload failed'));
            }
          } catch (e) {
            reject(new Error('Invalid JSON response'));
          }
        } else {
          reject(new Error('Upload failed with status ' + res.statusCode));
        }
      },
      fail: (err) => {
        reject(new Error(err.errMsg || 'Upload request failed'));
      }
    });
  });
};
