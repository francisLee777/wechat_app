import { getCurrentUser, loginUser, checkSession, User } from '../../utils/db';

Page({
  data: {
    isLoggedIn: false,
    hasFamily: false,
    user: null as User | null,
  },

  async onShow() {
    // 检查会话是否有效
    const isValid = await checkSession();
    if (isValid) {
      this.refreshUser();
    } else {
      this.setData({
        isLoggedIn: false,
        user: null,
        hasFamily: false
      });
    }
  },

  refreshUser() {
    const user = getCurrentUser();
    this.setData({
      isLoggedIn: !!user,
      user,
      hasFamily: !!(user && user.familyId)
    });
  },

  // 登录
  async handleLogin() {
    try {
      wx.showLoading({ title: '登录中...' });
      const user = await loginUser();
      this.refreshUser();
      wx.hideLoading();
      wx.showToast({ title: '登录成功', icon: 'success' });
    } catch (e: any) {
      wx.hideLoading();
      wx.showToast({ title: '登录失败', icon: 'none' });
    }
  },

  // 跳转到创建家庭页面
  goToCreateFamily() {
    wx.navigateTo({ url: '/pages/family/create' });
  },

  // 跳转到加入家庭页面
  goToJoinFamily() {
    wx.navigateTo({ url: '/pages/family/join' });
  },

  // 跳转到食谱列表
  goToRecipes() {
    wx.navigateTo({ url: '/pages/recipe/list' });
  },

  // 跳转到留言板
  goToMessages() {
    wx.navigateTo({ url: '/pages/message/board' });
  },

  // 跳转到家庭管理
  goToManage() {
    wx.navigateTo({ url: '/pages/family/manage' });
  },
  
  // 复制用户ID (方便测试)
  copyUserId() {
    if (this.data.user) {
      wx.setClipboardData({
        data: this.data.user.id,
        success: () => wx.showToast({ title: '用户ID已复制' })
      });
    }
  }
});
